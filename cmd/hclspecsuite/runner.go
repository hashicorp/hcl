package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	ctyjson "github.com/zclconf/go-cty/cty/json"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type Runner struct {
	parser      *hclparse.Parser
	hcldecPath  string
	baseDir     string
	logBegin    LogBeginCallback
	logProblems LogProblemsCallback
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
		if r.logBegin != nil {
			r.logBegin(prettyName, nil)
		}
		if r.logProblems != nil {
			r.logProblems(prettyName, nil, diags)
			return nil // don't duplicate the diagnostics we already reported
		}
		return diags
	}

	if r.logBegin != nil {
		r.logBegin(prettyName, tf)
	}

	basePath := filename[:len(filename)-2]
	specFilename := basePath + ".hcldec"
	nativeFilename := basePath + ".hcl"
	jsonFilename := basePath + ".hcl.json"

	// We'll add the source code of the spec file to our own parser, even
	// though it'll actually be parsed by the hcldec child process, since that
	// way we can produce nice diagnostic messages if hcldec fails to process
	// the spec file.
	src, err := ioutil.ReadFile(specFilename)
	if err == nil {
		r.parser.AddFile(specFilename, &hcl.File{
			Bytes: src,
		})
	}

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

	if r.logProblems != nil {
		r.logProblems(prettyName, nil, diags)
		return nil // don't duplicate the diagnostics we already reported
	}

	return diags
}

func (r *Runner) runTestInput(specFilename, inputFilename string, tf *TestFile) hcl.Diagnostics {
	// We'll add the source code of the input file to our own parser, even
	// though it'll actually be parsed by the hcldec child process, since that
	// way we can produce nice diagnostic messages if hcldec fails to process
	// the input file.
	src, err := ioutil.ReadFile(inputFilename)
	if err == nil {
		r.parser.AddFile(inputFilename, &hcl.File{
			Bytes: src,
		})
	}

	var diags hcl.Diagnostics

	if tf.ChecksTraversals {
		gotTraversals, moreDiags := r.hcldecVariables(specFilename, inputFilename)
		diags = append(diags, moreDiags...)
		if !moreDiags.HasErrors() {
			expected := tf.ExpectedTraversals
			for _, got := range gotTraversals {
				e := findTraversalSpec(got, expected)
				rng := got.SourceRange()
				if e == nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unexpected traversal",
						Detail:   "Detected traversal that is not indicated as expected in the test file.",
						Subject:  &rng,
					})
				} else {
					moreDiags := checkTraversalsMatch(got, inputFilename, e)
					diags = append(diags, moreDiags...)
				}
			}

			// Look for any traversals that didn't show up at all.
			for _, e := range expected {
				if t := findTraversalForSpec(e, gotTraversals); t == nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Missing expected traversal",
						Detail:   "This expected traversal was not detected.",
						Subject:  e.Traversal.SourceRange().Ptr(),
					})
				}
			}
		}

	}

	val, transformDiags := r.hcldecTransform(specFilename, inputFilename)
	if len(tf.ExpectedDiags) == 0 {
		diags = append(diags, transformDiags...)
		if transformDiags.HasErrors() {
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
							"Input file %s produced %#v, but was expecting %#v.\n\n%s",
							inputFilename, val, tf.Result,
							ctydebug.DiffValues(tf.Result, val),
						),
					})
				}
			}
		}
	} else {
		// We're expecting diagnostics, and so we'll need to correlate the
		// severities and source ranges of our actual diagnostics against
		// what we were expecting.
		type DiagnosticEntry struct {
			Severity hcl.DiagnosticSeverity
			Range    hcl.Range
		}
		got := make(map[DiagnosticEntry]*hcl.Diagnostic)
		want := make(map[DiagnosticEntry]hcl.Range)
		for _, diag := range transformDiags {
			if diag.Subject == nil {
				// Sourceless diagnostics can never be expected, so we'll just
				// pass these through as-is and assume they are hcldec
				// operational errors.
				diags = append(diags, diag)
				continue
			}
			if diag.Subject.Filename != inputFilename {
				// If the problem is for something other than the input file
				// then it can't be expected.
				diags = append(diags, diag)
				continue
			}
			entry := DiagnosticEntry{
				Severity: diag.Severity,
				Range:    *diag.Subject,
			}
			got[entry] = diag
		}
		for _, e := range tf.ExpectedDiags {
			e.Range.Filename = inputFilename // assumed here, since we don't allow any other filename to be expected
			entry := DiagnosticEntry{
				Severity: e.Severity,
				Range:    e.Range,
			}
			want[entry] = e.DeclRange
		}

		for gotEntry, diag := range got {
			if _, wanted := want[gotEntry]; !wanted {
				// Pass through the diagnostic itself so the user can see what happened
				diags = append(diags, diag)
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unexpected diagnostic",
					Detail: fmt.Sprintf(
						"No %s diagnostic was expected %s. The unexpected diagnostic was shown above.",
						severityString(gotEntry.Severity), rangeString(gotEntry.Range),
					),
					Subject: gotEntry.Range.Ptr(),
				})
			}
		}

		for wantEntry, declRange := range want {
			if _, gotted := got[wantEntry]; !gotted {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Missing expected diagnostic",
					Detail: fmt.Sprintf(
						"No %s diagnostic was generated %s.",
						severityString(wantEntry.Severity), rangeString(wantEntry.Range),
					),
					Subject: declRange.Ptr(),
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
			"--keep-nulls",
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

func (r *Runner) hcldecVariables(specFile, inputFile string) ([]hcl.Traversal, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	cmd := &exec.Cmd{
		Path: r.hcldecPath,
		Args: []string{
			r.hcldecPath,
			"--spec=" + specFile,
			"--diags=json",
			"--var-refs",
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
				Detail:   fmt.Sprintf("Sub-program hcldec (evaluating input) failed to start: %s.", err),
			})
			return nil, diags
		}

		// If we exited unsuccessfully then we'll expect diagnostics on stderr
		moreDiags := decodeJSONDiagnostics(errBuffer.Bytes())
		diags = append(diags, moreDiags...)
		return nil, diags
	} else {
		// Otherwise, we expect a JSON description of the traversals on stdout.
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
			Steps []StepJSON `json:"steps"`
		}

		var raw []TraversalJSON
		err := json.Unmarshal(outBuffer.Bytes(), &raw)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Failed to parse hcldec result",
				Detail:   fmt.Sprintf("Sub-program hcldec (with --var-refs) produced an invalid result: %s.", err),
			})
			return nil, diags
		}

		var ret []hcl.Traversal
		if len(raw) == 0 {
			return ret, diags
		}

		ret = make([]hcl.Traversal, 0, len(raw))
		for _, rawT := range raw {
			traversal := make(hcl.Traversal, 0, len(rawT.Steps))
			for _, rawS := range rawT.Steps {
				rng := hcl.Range{
					Filename: rawS.Range.Filename,
					Start: hcl.Pos{
						Line:   rawS.Range.Start.Line,
						Column: rawS.Range.Start.Column,
						Byte:   rawS.Range.Start.Byte,
					},
					End: hcl.Pos{
						Line:   rawS.Range.End.Line,
						Column: rawS.Range.End.Column,
						Byte:   rawS.Range.End.Byte,
					},
				}

				switch rawS.Kind {

				case "root":
					traversal = append(traversal, hcl.TraverseRoot{
						Name:     rawS.Name,
						SrcRange: rng,
					})

				case "attr":
					traversal = append(traversal, hcl.TraverseAttr{
						Name:     rawS.Name,
						SrcRange: rng,
					})

				case "index":
					ty, err := ctyjson.ImpliedType([]byte(rawS.Key))
					if err != nil {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Failed to parse hcldec result",
							Detail:   fmt.Sprintf("Sub-program hcldec (with --var-refs) produced an invalid result: traversal step has invalid index key %s.", rawS.Key),
						})
						return nil, diags
					}
					keyVal, err := ctyjson.Unmarshal([]byte(rawS.Key), ty)
					if err != nil {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Failed to parse hcldec result",
							Detail:   fmt.Sprintf("Sub-program hcldec (with --var-refs) produced a result with an invalid index key %s: %s.", rawS.Key, err),
						})
						return nil, diags
					}

					traversal = append(traversal, hcl.TraverseIndex{
						Key:      keyVal,
						SrcRange: rng,
					})

				default:
					// Should never happen since the above cases are exhaustive,
					// but we'll catch it gracefully since this is coming from
					// a possibly-buggy hcldec implementation that we're testing.
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Failed to parse hcldec result",
						Detail:   fmt.Sprintf("Sub-program hcldec (with --var-refs) produced an invalid result: traversal step of unsupported kind %q.", rawS.Kind),
					})
					return nil, diags
				}
			}

			ret = append(ret, traversal)
		}
		return ret, diags
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
