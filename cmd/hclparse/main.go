// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

const versionStr = "0.0.1-dev"

var (
	exprMode     = flag.Bool("e", false, "parse as expression")
	templateMode = flag.Bool("t", false, "parse as template")
	showVersion  = flag.Bool("version", false, "show the version number and immediately exit")
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Println(versionStr)
		return
	}

	if flag.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "only one file or content can be specified\n")
		os.Exit(2)
	}

	if flag.NArg() == 0 {
		os.Exit(processFile("<stdin>", os.Stdin))
	}

	var fn string
	var in io.Reader
	if *exprMode || *templateMode {
		fn = "<input>"
		in = strings.NewReader(flag.Arg(0))
	} else {
		fn = flag.Arg(0)
		in = nil
	}
	os.Exit(processFile(fn, in))
}

func processFile(fn string, in io.Reader) int {
	var err error
	if in == nil {
		in, err = os.Open(fn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open %s: %s\n", fn, err)
			return 1
		}
	}

	inSrc, err := io.ReadAll(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read %s: %s\n", fn, err)
		return 1
	}

	var node hclsyntax.Node
	var diags hcl.Diagnostics
	if *exprMode {
		node, diags = hclsyntax.ParseExpression(inSrc, fn, hcl.InitialPos)
	} else if *templateMode {
		node, diags = hclsyntax.ParseTemplate(inSrc, fn, hcl.InitialPos)
	} else {
		file, d := hclsyntax.ParseConfig(inSrc, fn, hcl.InitialPos)
		node = file.Body.(*hclsyntax.Body)
		diags = d
	}

	if diags.HasErrors() {
		fmt.Fprintf(os.Stderr, "failed to parse. %d diagnostic(s):\n\n", len(diags))
		for _, diag := range diags {
			fmt.Fprintf(os.Stderr, "%s: %s\n", diag.Summary, diag.Detail)
		}
		return 1
	}

	diags = hclsyntax.Walk(node, &walker{file: inSrc})
	if diags.HasErrors() {
		fmt.Fprintf(os.Stderr, "failed to walk. %d diagnostic(s):\n\n", len(diags))
		for _, diag := range diags {
			fmt.Fprintf(os.Stderr, "%s: %s\n", diag.Summary, diag.Detail)
		}
		return 1
	}

	return 0
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: hclparse [options] [file or content]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

type walker struct {
	indent int
	leaf   bool
	file   []byte
}

var _ hclsyntax.Walker = (*walker)(nil)

func (w *walker) Enter(node hclsyntax.Node) hcl.Diagnostics {
	if w.leaf {
		panic("leaf node should not have children")
	}

	fmt.Print(strings.Repeat(" ", w.indent))

	switch node := node.(type) {
	case *hclsyntax.Attribute:
		fmt.Printf(`(%T "%s"`, node, node.Name)
	case *hclsyntax.Block:
		fmt.Printf(`(%T "%s" %s`, node, node.Type, node.Labels)
	case *hclsyntax.LiteralValueExpr:
		fmt.Printf(`(%T "%s")`, node, node.SrcRange.SliceBytes(w.file))
		w.leaf = true
	case *hclsyntax.ScopeTraversalExpr:
		fmt.Printf(`(%T "%s")`, node, node.SrcRange.SliceBytes(w.file))
		w.leaf = true
	case *hclsyntax.RelativeTraversalExpr:
		fmt.Printf(`(%T "%s"`, node, node.Traversal.SourceRange().SliceBytes(w.file))
	case *hclsyntax.FunctionCallExpr:
		fmt.Printf(`(%T "%s"`, node, node.Name)
	case *hclsyntax.ForExpr:
		fmt.Printf(`(%T`, node)
		if node.KeyVar != "" {
			fmt.Printf(` key="%s"`, node.KeyVar)
		}
		if node.ValVar != "" {
			fmt.Printf(` val="%s"`, node.ValVar)
		}
	case *hclsyntax.AnonSymbolExpr:
		fmt.Printf(`(%T)`, node)
		w.leaf = true
	case *hclsyntax.BinaryOpExpr:
		fmt.Printf(`(%T "%s"`, node, opAsString(node.Op))
	case *hclsyntax.UnaryOpExpr:
		fmt.Printf(`(%T "%s"`, node, opAsString(node.Op))
	default:
		fmt.Printf("(%T", node)
	}

	fmt.Print("\n")
	w.indent += 2
	return nil
}

func (w *walker) Exit(node hclsyntax.Node) hcl.Diagnostics {
	w.indent -= 2

	if w.leaf {
		w.leaf = false
		return nil
	}

	fmt.Print(strings.Repeat(" ", w.indent))
	fmt.Printf(")\n")
	return nil
}

func opAsString(op *hclsyntax.Operation) string {
	switch op {
	case hclsyntax.OpLogicalOr:
		return "||"
	case hclsyntax.OpLogicalAnd:
		return "&&"
	case hclsyntax.OpLogicalNot:
		return "!"
	case hclsyntax.OpEqual:
		return "=="
	case hclsyntax.OpNotEqual:
		return "!="
	case hclsyntax.OpGreaterThan:
		return ">"
	case hclsyntax.OpGreaterThanOrEqual:
		return ">="
	case hclsyntax.OpLessThan:
		return "<"
	case hclsyntax.OpLessThanOrEqual:
		return "<="
	case hclsyntax.OpAdd:
		return "+"
	case hclsyntax.OpSubtract:
		return "-"
	case hclsyntax.OpMultiply:
		return "*"
	case hclsyntax.OpDivide:
		return "/"
	case hclsyntax.OpModulo:
		return "%"
	case hclsyntax.OpNegate:
		return "-"
	default:
		panic(fmt.Sprintf("unknown operation type: %T", op))
	}
}
