package printer

import (
	"io"
	"text/tabwriter"

	"github.com/fatih/hcl/ast"
)

type printer struct {
	out  []byte // raw printer result
	cfg  Config
	node ast.Node
}

func (p *printer) output() []byte {
	return p.printNode(p.node)
}

// A Config node controls the output of Fprint.
type Config struct {
	SpaceWidth int // if set, it will use spaces instead of tabs for alignment
}

func (c *Config) fprint(output io.Writer, node ast.Node) error {
	p := &printer{
		cfg:  *c,
		node: node,
	}

	if _, err := output.Write(p.output()); err != nil {
		return err
	}

	// flush tabwriter, if any
	var err error
	if tw, _ := output.(*tabwriter.Writer); tw != nil {
		err = tw.Flush()
	}

	return err
}

func (c *Config) Fprint(output io.Writer, node ast.Node) error {
	return c.fprint(output, node)
}

// Fprint "pretty-prints" an HCL node to output
// It calls Config.Fprint with default settings.
func Fprint(output io.Writer, node ast.Node) error {
	return (&Config{}).Fprint(output, node)
}
