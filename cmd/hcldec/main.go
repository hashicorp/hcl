package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcldec"
	"github.com/hashicorp/hcl2/hclparse"
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
	showVersion = flag.BoolP("version", "v", false, "show the version number and immediately exit")
)

var parser = hclparse.NewParser()
var diagWr hcl.DiagnosticWriter // initialized in init

func init() {
	flag.VarP(vars, "vars", "V", "provide variables to the given configuration file(s)")

	color := terminal.IsTerminal(int(os.Stderr.Fd()))
	w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		w = 80
	}
	diagWr = hcl.NewDiagnosticTextWriter(os.Stderr, parser.Files(), uint(w), color)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Println(versionStr)
		os.Exit(0)
	}

	args := flag.Args()

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
			f, fDiags := parser.ParseHCLFile(filename)
			diags = append(diags, fDiags...)
			if !fDiags.HasErrors() {
				bodies = append(bodies, f.Body)
			}
		}
	}

	if diags.HasErrors() {
		diagWr.WriteDiagnostics(diags)
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

	val, decDiags := hcldec.Decode(body, spec, ctx)
	diags = append(diags, decDiags...)

	if diags.HasErrors() {
		diagWr.WriteDiagnostics(diags)
		os.Exit(2)
	}

	out, err := ctyjson.Marshal(val, val.Type())
	if err != nil {
		return err
	}

	// hcldec will include explicit nulls where an ObjectSpec has a spec
	// that refers to a missing item, but that'll probably be annoying for
	// a consumer of our output to deal with so we'll just strip those
	// out and reduce to only the non-null values.
	out = stripJSONNullProperties(out)

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

func stripJSONNullProperties(src []byte) []byte {
	var v interface{}
	err := json.Unmarshal(src, &v)
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
