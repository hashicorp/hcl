package main

import (
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
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

	runner := &Runner{
		parser:     parser,
		hcldecPath: hcldecPath,
		baseDir:    testsDir,
		log: func(name string, file *TestFile) {
			fmt.Printf("- %s\n", name)
		},
	}
	diags := runner.Run()

	if len(diags) != 0 {
		os.Stderr.WriteString("\n")
		color := terminal.IsTerminal(int(os.Stderr.Fd()))
		w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			w = 80
		}
		diagWr := hcl.NewDiagnosticTextWriter(os.Stderr, parser.Files(), uint(w), color)
		diagWr.WriteDiagnostics(diags)
		return 2
	}

	return 0
}
