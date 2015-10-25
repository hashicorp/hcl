// Package printer implements printing of AST nodes to HCL format.
package printer

import (
	"io"
	"text/tabwriter"

	"github.com/fatih/hcl/ast"
)

type printer struct {
	cfg Config
}

// A Config node controls the output of Fprint.
type Config struct {
	SpacesWidth int // if set, it will use spaces instead of tabs for alignment
}

func (c *Config) Fprint(output io.Writer, node ast.Node) error {
	p := &printer{
		cfg: *c,
	}

	if _, err := output.Write(p.output(node)); err != nil {
		return err
	}

	// flush tabwriter, if any
	var err error
	if tw, _ := output.(*tabwriter.Writer); tw != nil {
		err = tw.Flush()
	}

	return err
}

// Fprint "pretty-prints" an HCL node to output
// It calls Config.Fprint with default settings.
func Fprint(output io.Writer, node ast.Node) error {
	return (&Config{}).Fprint(output, node)
}
