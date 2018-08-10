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

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/zclconf/go-cty/cty"
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
		filename = filepath.Join(r.baseDir, filename)
		testDiags := r.runTest(filename)
		diags = append(diags, testDiags...)
	}

	for _, dirName := range subDirs {
		dir := filepath.Join(r.baseDir, dirName)
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
	//jsonFilename := basePath + ".hcl.json"

	if _, err := os.Stat(specFilename); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing .hcldec file",
			Detail:   fmt.Sprintf("No specification file for test %s: %s.", prettyName, err),
		})
		return diags
	}

	if _, err := os.Stat(nativeFilename); err == nil {

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
			"--spec=" + specFile,
			"--diags=json",
			inputFile,
		},
		Stdout: &outBuffer,
		Stderr: &errBuffer,
	}
	err := cmd.Run()
	if _, isExit := err.(*exec.ExitError); !isExit {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to run hcldec",
			Detail:   fmt.Sprintf("Sub-program hcldec failed to start: %s.", err),
		})
		return cty.DynamicVal, diags
	}

	if err != nil {
		// If we exited unsuccessfully then we'll expect diagnostics on stderr
		// TODO: implement that
	} else {
		// Otherwise, we expect a JSON result value on stdout
		// TODO: implement that
	}

	return cty.DynamicVal, diags
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
