package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hclparse"
	flag "github.com/spf13/pflag"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	"golang.org/x/crypto/ssh/terminal"
)

const versionStr = "0.0.1-dev"

// vars is populated from --vars arguments on the command line, via a flag
// registration in init() below.
var vars = &varSpecs{}

var (
	specFile    = flag.StringP("spec", "s", "", "path to spec file (required)")
	outputFile  = flag.StringP("out", "o", "", "write to the given file, instead of stdout")
	diagsFormat = flag.StringP("diags", "", "", "format any returned diagnostics in the given format; currently only \"json\" is accepted")
	showVarRefs = flag.BoolP("var-refs", "", false, "rather than decoding input, produce a JSON description of the variables referenced by it")
	withType    = flag.BoolP("with-type", "", false, "include an additional object level at the top describing the HCL-oriented type of the result value")
	showVersion = flag.BoolP("version", "v", false, "show the version number and immediately exit")
	keepNulls   = flag.BoolP("keep-nulls", "", false, "retain object properties that have null as their value (they are removed by default)")
)

var parser = hclparse.NewParser()
var diagWr hcl.DiagnosticWriter // initialized in init

func init() {
	flag.VarP(vars, "vars", "V", "provide variables to the given configuration file(s)")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Println(versionStr)
		os.Exit(0)
	}

	args := flag.Args()

	switch *diagsFormat {
	case "":
		color := terminal.IsTerminal(int(os.Stderr.Fd()))
		w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			w = 80
		}
		diagWr = hcl.NewDiagnosticTextWriter(os.Stderr, parser.Files(), uint(w), color)
	case "json":
		diagWr = &jsonDiagWriter{w: os.Stderr}
	default:
		fmt.Fprintf(os.Stderr, "Invalid diagnostics format %q: only \"json\" is supported.\n", *diagsFormat)
		os.Exit(2)
	}

	err := realmain(args)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
		os.Exit(1)
	}
}

func realmain(args []string) error {

	if *specFile == "" {
		return fmt.Errorf("the --spec=... argument is required")
	}

	var diags hcl.Diagnostics

	specContent, specDiags := loadSpecFile(*specFile)
	diags = append(diags, specDiags...)
	if specDiags.HasErrors() {
		diagWr.WriteDiagnostics(diags)
		flush(diagWr)
		os.Exit(2)
	}

	spec := specContent.RootSpec

	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
		Functions: map[string]function.Function{},
	}
	for name, val := range specContent.Variables {
		ctx.Variables[name] = val
	}
	for name, f := range specContent.Functions {
		ctx.Functions[name] = f
	}
	if len(*vars) != 0 {
		for i, varsSpec := range *vars {
			var vals map[string]cty.Value
			var valsDiags hcl.Diagnostics
			if strings.HasPrefix(strings.TrimSpace(varsSpec), "{") {
				// literal JSON object on the command line
				vals, valsDiags = parseVarsArg(varsSpec, i)
			} else {
				// path to a file containing either HCL or JSON (by file extension)
				vals, valsDiags = parseVarsFile(varsSpec)
			}
			diags = append(diags, valsDiags...)
			for k, v := range vals {
				ctx.Variables[k] = v
			}
		}
	}

	// If we have empty context elements then we'll nil them out so that
	// we'll produce e.g. "variables are not allowed" errors instead of
	// "variable not found" errors.
	if len(ctx.Variables) == 0 {
		ctx.Variables = nil
	}
	if len(ctx.Functions) == 0 {
		ctx.Functions = nil
	}
	if ctx.Variables == nil && ctx.Functions == nil {
		ctx = nil
	}

	var bodies []hcl.Body

	if len(args) == 0 {
		src, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %s", err)
		}

		f, fDiags := parser.ParseHCL(src, "<stdin>")
		diags = append(diags, fDiags...)
		if !fDiags.HasErrors() {
			bodies = append(bodies, f.Body)
		}
	} else {
		for _, filename := range args {
			var f *hcl.File
			var fDiags hcl.Diagnostics
			if strings.HasSuffix(filename, ".json") {
				f, fDiags = parser.ParseJSONFile(filename)
			} else {
				f, fDiags = parser.ParseHCLFile(filename)
			}
			diags = append(diags, fDiags...)
			if !fDiags.HasErrors() {
				bodies = append(bodies, f.Body)
			}
		}
	}

	if diags.HasErrors() {
		diagWr.WriteDiagnostics(diags)
		flush(diagWr)
		os.Exit(2)
	}

	var body hcl.Body
	switch len(bodies) {
	case 0:
		// should never happen, but... okay?
		body = hcl.EmptyBody()
	case 1:
		body = bodies[0]
	default:
		body = hcl.MergeBodies(bodies)
	}

	if *showVarRefs {
		vars := hcldec.Variables(body, spec)
		return showVarRefsJSON(vars, ctx)
	}

	val, decDiags := hcldec.Decode(body, spec, ctx)
	diags = append(diags, decDiags...)

	if diags.HasErrors() {
		diagWr.WriteDiagnostics(diags)
		flush(diagWr)
		os.Exit(2)
	}

	wantType := val.Type()
	if *withType {
		// We'll instead ask to encode as dynamic, which will make the
		// marshaler include type information.
		wantType = cty.DynamicPseudoType
	}
	out, err := ctyjson.Marshal(val, wantType)
	if err != nil {
		return err
	}

	// hcldec will include explicit nulls where an ObjectSpec has a spec
	// that refers to a missing item, but that'll probably be annoying for
	// a consumer of our output to deal with so we'll just strip those
	// out and reduce to only the non-null values.
	if !*keepNulls {
		out = stripJSONNullProperties(out)
	}

	target := os.Stdout
	if *outputFile != "" {
		target, err = os.OpenFile(*outputFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return fmt.Errorf("can't open %s for writing: %s", *outputFile, err)
		}
	}

	fmt.Fprintf(target, "%s\n", out)

	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: hcldec --spec=<spec-file> [options] [hcl-file ...]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func showVarRefsJSON(vars []hcl.Traversal, ctx *hcl.EvalContext) error {
	type PosJSON struct {
		Line   int `json:"line"`
		Column int `json:"column"`
		Byte   int `json:"byte"`
	}
	type RangeJSON struct {
		Filename string  `json:"filename"`
		Start    PosJSON `json:"start"`
		End      PosJSON `json:"end"`
	}
	type StepJSON struct {
		Kind  string          `json:"kind"`
		Name  string          `json:"name,omitempty"`
		Key   json.RawMessage `json:"key,omitempty"`
		Range RangeJSON       `json:"range"`
	}
	type TraversalJSON struct {
		RootName string          `json:"root_name"`
		Value    json.RawMessage `json:"value,omitempty"`
		Steps    []StepJSON      `json:"steps"`
		Range    RangeJSON       `json:"range"`
	}

	ret := make([]TraversalJSON, 0, len(vars))
	for _, traversal := range vars {
		tJSON := TraversalJSON{
			Steps: make([]StepJSON, 0, len(traversal)),
		}

		for _, step := range traversal {
			var sJSON StepJSON
			rng := step.SourceRange()
			sJSON.Range.Filename = rng.Filename
			sJSON.Range.Start.Line = rng.Start.Line
			sJSON.Range.Start.Column = rng.Start.Column
			sJSON.Range.Start.Byte = rng.Start.Byte
			sJSON.Range.End.Line = rng.End.Line
			sJSON.Range.End.Column = rng.End.Column
			sJSON.Range.End.Byte = rng.End.Byte
			switch ts := step.(type) {
			case hcl.TraverseRoot:
				sJSON.Kind = "root"
				sJSON.Name = ts.Name
				tJSON.RootName = ts.Name
			case hcl.TraverseAttr:
				sJSON.Kind = "attr"
				sJSON.Name = ts.Name
			case hcl.TraverseIndex:
				sJSON.Kind = "index"
				src, err := ctyjson.Marshal(ts.Key, ts.Key.Type())
				if err == nil {
					sJSON.Key = json.RawMessage(src)
				}
			default:
				// Should never get here, since the above should be exhaustive
				// for all possible traversal step types.
				sJSON.Kind = "(unknown)"
			}
			tJSON.Steps = append(tJSON.Steps, sJSON)
		}

		// Best effort, we'll try to include the current known value of this
		// traversal, if any.
		val, diags := traversal.TraverseAbs(ctx)
		if !diags.HasErrors() {
			enc, err := ctyjson.Marshal(val, val.Type())
			if err == nil {
				tJSON.Value = json.RawMessage(enc)
			}
		}

		rng := traversal.SourceRange()
		tJSON.Range.Filename = rng.Filename
		tJSON.Range.Start.Line = rng.Start.Line
		tJSON.Range.Start.Column = rng.Start.Column
		tJSON.Range.Start.Byte = rng.Start.Byte
		tJSON.Range.End.Line = rng.End.Line
		tJSON.Range.End.Column = rng.End.Column
		tJSON.Range.End.Byte = rng.End.Byte

		ret = append(ret, tJSON)
	}

	out, err := json.MarshalIndent(ret, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal variable references as JSON: %s", err)
	}

	target := os.Stdout
	if *outputFile != "" {
		target, err = os.OpenFile(*outputFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return fmt.Errorf("can't open %s for writing: %s", *outputFile, err)
		}
	}

	fmt.Fprintf(target, "%s\n", out)

	return nil
}

func stripJSONNullProperties(src []byte) []byte {
	dec := json.NewDecoder(bytes.NewReader(src))
	dec.UseNumber()

	var v interface{}
	err := dec.Decode(&v)
	if err != nil {
		// We expect valid JSON
		panic(err)
	}

	v = stripNullMapElements(v)

	new, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return new
}

func stripNullMapElements(v interface{}) interface{} {
	switch tv := v.(type) {
	case map[string]interface{}:
		for k, ev := range tv {
			if ev == nil {
				delete(tv, k)
			} else {
				tv[k] = stripNullMapElements(ev)
			}
		}
		return v
	case []interface{}:
		for i, ev := range tv {
			tv[i] = stripNullMapElements(ev)
		}
		return v
	default:
		return v
	}
}
