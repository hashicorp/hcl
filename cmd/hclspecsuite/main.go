package main

import (
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

func main() {
	os.Exit(realMain(os.Args[1:]))
}

func realMain(args []string) int {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: hclspecsuite <tests-dir> <hcldec-file>\n")
		return 2
	}

	testsDir := args[0]
	hcldecPath := args[1]

	hcldecPath, err := exec.LookPath(hcldecPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 2
	}

	parser := hclparse.NewParser()

	color := terminal.IsTerminal(int(os.Stderr.Fd()))
	w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		w = 80
	}
	diagWr := hcl.NewDiagnosticTextWriter(os.Stderr, parser.Files(), uint(w), color)
	var diagCount int

	runner := &Runner{
		parser:     parser,
		hcldecPath: hcldecPath,
		baseDir:    testsDir,
		logBegin: func(name string, file *TestFile) {
			fmt.Printf("- %s\n", name)
		},
		logProblems: func(name string, file *TestFile, diags hcl.Diagnostics) {
			if len(diags) != 0 {
				os.Stderr.WriteString("\n")
				diagWr.WriteDiagnostics(diags)
				diagCount += len(diags)
			}
			fmt.Printf("- %s\n", name)
		},
	}
	diags := runner.Run()

	if len(diags) != 0 {
		os.Stderr.WriteString("\n\n\n== Test harness problems:\n\n")
		diagWr.WriteDiagnostics(diags)
		diagCount += len(diags)
	}

	if diagCount > 0 {
		return 2
	}
	return 0
}
