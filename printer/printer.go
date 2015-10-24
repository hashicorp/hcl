package printer

import (
	"bytes"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/fatih/hcl/ast"
)

type printer struct {
	cfg  Config
	node ast.Node
}

func (p *printer) output() []byte {
	var buf bytes.Buffer
	fmt.Println("STARTING OUTPUT")

	ast.Walk(p.node, func(n ast.Node) bool {
		fmt.Printf("n = %+v\n", n)
		return true
	})

	return buf.Bytes()
}

// A Mode value is a set of flags (or 0). They control printing.
type Mode uint

const (
	RawFormat Mode = 1 << iota // do not use a tabwriter; if set, UseSpaces is ignored
	TabIndent                  // use tabs for indentation independent of UseSpaces
	UseSpaces                  // use spaces instead of tabs for alignment
)

// A Config node controls the output of Fprint.
type Config struct {
	Mode     Mode // default: 0
	Tabwidth int  // default: 8
	Indent   int  // default: 0 (all code is indented at least by this much)
}

func (c *Config) fprint(output io.Writer, node ast.Node) error {
	p := &printer{
		cfg:  *c,
		node: node,
	}

	// TODO(arslan): implement this
	// redirect output through a trimmer to eliminate trailing whitespace
	// (Input to a tabwriter must be untrimmed since trailing tabs provide
	// formatting information. The tabwriter could provide trimming
	// functionality but no tabwriter is used when RawFormat is set.)
	// output = &trimmer{output: output}

	// redirect output through a tabwriter if necessary
	if c.Mode&RawFormat == 0 {
		minwidth := c.Tabwidth

		padchar := byte('\t')
		if c.Mode&UseSpaces != 0 {
			padchar = ' '
		}

		twmode := tabwriter.DiscardEmptyColumns
		if c.Mode&TabIndent != 0 {
			minwidth = 0
			twmode |= tabwriter.TabIndent
		}

		output = tabwriter.NewWriter(output, minwidth, c.Tabwidth, 1, padchar, twmode)
	}

	// write printer result via tabwriter/trimmer to output
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
	return (&Config{Tabwidth: 8}).Fprint(output, node)
}
