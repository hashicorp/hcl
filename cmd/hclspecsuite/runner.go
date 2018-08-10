package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl2/ext/typeexpr"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type Runner struct {
	parser     *hclparse.Parser
	hcldecPath string
	baseDir    string
	log        LogCallback
}

func (r *Runner) Run() hcl.Diagnostics {
	return r.runDir(r.baseDir)
}

func (r *Runner) runDir(dir string) hcl.Diagnostics {
	var diags hcl.Diagnostics

	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to read test directory",
			Detail:   fmt.Sprintf("The directory %q could not be opened: %s.", dir, err),
		})
		return diags
	}

	var tests []string
	var subDirs []string
	for _, info := range infos {
		name := info.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		if info.IsDir() {
			subDirs = append(subDirs, name)
		}
		if strings.HasSuffix(name, ".t") {
			tests = append(tests, name)
		}
	}
	sort.Strings(tests)
	sort.Strings(subDirs)

	for _, filename := range tests {
		filename = filepath.Join(dir, filename)
		testDiags := r.runTest(filename)
		diags = append(diags, testDiags...)
	}

	for _, dirName := range subDirs {
		dir := filepath.Join(dir, dirName)
		dirDiags := r.runDir(dir)
		diags = append(diags, dirDiags...)
	}

	return diags
}

func (r *Runner) runTest(filename string) hcl.Diagnostics {
	prettyName := r.prettyTestName(filename)
	tf, diags := r.LoadTestFile(filename)
	if diags.HasErrors() {
		// We'll still log, so it's clearer which test the diagnostics belong to.
		if r.log != nil {
			r.log(prettyName, nil)
		}
		return diags
	}

	if r.log != nil {
		r.log(prettyName, tf)
	}

	basePath := filename[:len(filename)-2]
	specFilename := basePath + ".hcldec"
	nativeFilename := basePath + ".hcl"
	jsonFilename := basePath + ".hcl.json"

	if _, err := os.Stat(specFilename); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing .hcldec file",
			Detail:   fmt.Sprintf("No specification file for test %s: %s.", prettyName, err),
		})
		return diags
	}

	if _, err := os.Stat(nativeFilename); err == nil {
		moreDiags := r.runTestInput(specFilename, nativeFilename, tf)
		diags = append(diags, moreDiags...)
	}

	if _, err := os.Stat(jsonFilename); err == nil {
		moreDiags := r.runTestInput(specFilename, jsonFilename, tf)
		diags = append(diags, moreDiags...)
	}

	return diags
}

func (r *Runner) runTestInput(specFilename, inputFilename string, tf *TestFile) hcl.Diagnostics {
	// We'll add the source code of the input file to our own parser, even
	// though it'll actually be parsed by the hcldec child process, since that
	// way we can produce nice diagnostic messages if hcldec fails to process
	// the input file.
	if src, err := ioutil.ReadFile(inputFilename); err == nil {
		r.parser.AddFile(inputFilename, &hcl.File{
			Bytes: src,
		})
	}

	var diags hcl.Diagnostics

	val, moreDiags := r.hcldecTransform(specFilename, inputFilename)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		// If hcldec failed then there's no point in continuing.
		return diags
	}

	if errs := val.Type().TestConformance(tf.ResultType); len(errs) > 0 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Incorrect result type",
			Detail: fmt.Sprintf(
				"Input file %s produced %s, but was expecting %s.",
				inputFilename, typeexpr.TypeString(val.Type()), typeexpr.TypeString(tf.ResultType),
			),
		})
	}

	if tf.Result != cty.NilVal {
		cmpVal, err := convert.Convert(tf.Result, tf.ResultType)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Incorrect type for result value",
				Detail: fmt.Sprintf(
					"Result does not conform to the given result type: %s.", err,
				),
				Subject: &tf.ResultRange,
			})
		} else {
			if !val.RawEquals(cmpVal) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Incorrect result value",
					Detail: fmt.Sprintf(
						"Input file %s produced %#v, but was expecting %#v.",
						inputFilename, val, tf.Result,
					),
				})
			}
		}
	}

	return diags
}

func (r *Runner) hcldecTransform(specFile, inputFile string) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	cmd := &exec.Cmd{
		Path: r.hcldecPath,
		Args: []string{
			r.hcldecPath,
			"--spec=" + specFile,
			"--diags=json",
			"--with-type",
			inputFile,
		},
		Stdout: &outBuffer,
		Stderr: &errBuffer,
	}
	err := cmd.Run()
	if err != nil {
		if _, isExit := err.(*exec.ExitError); !isExit {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Failed to run hcldec",
				Detail:   fmt.Sprintf("Sub-program hcldec failed to start: %s.", err),
			})
			return cty.DynamicVal, diags
		}

		// If we exited unsuccessfully then we'll expect diagnostics on stderr
		moreDiags := decodeJSONDiagnostics(errBuffer.Bytes())
		diags = append(diags, moreDiags...)
		return cty.DynamicVal, diags
	} else {
		// Otherwise, we expect a JSON result value on stdout. Since we used
		// --with-type above, we can decode as DynamicPseudoType to recover
		// exactly the type that was saved, without the usual JSON lossiness.
		val, err := ctyjson.Unmarshal(outBuffer.Bytes(), cty.DynamicPseudoType)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Failed to parse hcldec result",
				Detail:   fmt.Sprintf("Sub-program hcldec produced an invalid result: %s.", err),
			})
			return cty.DynamicVal, diags
		}
		return val, diags
	}
}

func (r *Runner) prettyDirName(dir string) string {
	rel, err := filepath.Rel(r.baseDir, dir)
	if err != nil {
		return filepath.ToSlash(dir)
	}
	return filepath.ToSlash(rel)
}

func (r *Runner) prettyTestName(filename string) string {
	dir := filepath.Dir(filename)
	dirName := r.prettyDirName(dir)
	filename = filepath.Base(filename)
	testName := filename[:len(filename)-2]
	if dirName == "." {
		return testName
	}
	return fmt.Sprintf("%s/%s", dirName, testName)
}
