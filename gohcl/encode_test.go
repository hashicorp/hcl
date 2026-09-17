// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package gohcl_test

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"testing"
)

func ExampleEncodeIntoBody() {
	type Service struct {
		Name string   `hcl:"name,label"`
		Exe  []string `hcl:"executable"`
	}
	type Constraints struct {
		OS   string `hcl:"os"`
		Arch string `hcl:"arch"`
	}
	type App struct {
		Name        string       `hcl:"name"`
		Desc        string       `hcl:"description"`
		Constraints *Constraints `hcl:"constraints,block"`
		Services    []Service    `hcl:"service,block"`
	}

	app := App{
		Name: "awesome-app",
		Desc: "Such an awesome application",
		Constraints: &Constraints{
			OS:   "linux",
			Arch: "amd64",
		},
		Services: []Service{
			{
				Name: "web",
				Exe:  []string{"./web", "--listen=:8080"},
			},
			{
				Name: "worker",
				Exe:  []string{"./worker"},
			},
		},
	}

	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(&app, f.Body())
	fmt.Printf("%s", f.Bytes())

	// Output:
	// name        = "awesome-app"
	// description = "Such an awesome application"
	//
	// constraints {
	//   os   = "linux"
	//   arch = "amd64"
	// }
	//
	// service "web" {
	//   executable = ["./web", "--listen=:8080"]
	// }
	// service "worker" {
	//   executable = ["./worker"]
	// }
}

// The following tests define a type alias and struct that implement ExpressionMarshaler and BlockMarshaler

// rawString is a string that will be written literally to HCL without any escaping
type rawString string

func (r rawString) MarshalExpression() (*hclwrite.Expression, error) {
	return hclwrite.NewExpressionRaw(hclwrite.Tokens{
		{
			Bytes: []byte(r),
		},
	}), nil
}

// customBlock implements BlockMarshalelr to hardcode a label and customize attribute name
type customBlock struct {
	Name string
}

func (c customBlock) MarshalBlock(blockType string) *hclwrite.Block {
	block := hclwrite.NewBlock(blockType, []string{"hardcoded_label"})
	body := block.Body()
	body.SetAttributeValue(fmt.Sprintf("name_%s", c.Name), cty.StringVal(c.Name))
	return block
}

type testBlock struct {
	Label       string      `hcl:",label"`
	Title       string      `hcl:"title"`
	RawExpr     rawString   `hcl:"raw"`
	CustomBlock customBlock `hcl:"custom,block"`
}

func TestMarshalInterfaces(t *testing.T) {
	inBlock := testBlock{
		Label:   "label1",
		Title:   "title",
		RawExpr: "foo.bar[0]",
		CustomBlock: customBlock{
			Name: "Foobar",
		},
	}
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(&inBlock, f.Body())

	want := `title = "title"
raw   = foo.bar[0]

custom "hardcoded_label" {
  name_Foobar = "Foobar"
}
`
	got := fmt.Sprintf("%s", hclwrite.Format(f.Bytes()))
	if got != want {
		t.Errorf("got: %q, wanted, %q", got, want)
	}
}

// These tests are in place to ensure the interface checks for BlockMarshalers work with ptrs and concrete types and panic on nil and zero values
type Fruit interface {
	isFruit()
	gohcl.BlockMarshaler
}

type Apple struct {
	Name string `hcl:"name"`
}

func (a Apple) isFruit() {}

func (a Apple) MarshalBlock(blockType string) *hclwrite.Block {
	type customBlock struct {
		FruitName string `hcl:"fruit_name"`
	}
	return gohcl.EncodeAsBlock(customBlock{FruitName: "apple"}, blockType)
}

func TestInterfaceBlocks(t *testing.T) {
	type Banana struct {
		Name  string `hcl:"name"`
		Fruit Fruit  `hcl:"fruit,block"`
	}
	testCases := []struct {
		name      string
		banana    Banana
		wantHcl   string
		wantPanic bool
	}{
		{
			name: "with concrete type",
			banana: Banana{
				Name:  "my banana",
				Fruit: Apple{Name: "my-apple"},
			},
			wantHcl: `banana {
  name = "my banana"

  fruit {
    fruit_name = "apple"
  }
}
`,
		},
		{
			name: "with ptr to struct",
			banana: Banana{
				Name:  "my banana",
				Fruit: &Apple{Name: "my-apple-ptr"},
			},
			wantHcl: `banana {
  name = "my banana"

  fruit {
    fruit_name = "apple"
  }
}
`,
		},
		{
			name: "with nil",
			banana: Banana{
				Name:  "my banana",
				Fruit: nil,
			},
			wantPanic: true,
		},
		{
			name: "with empty",
			banana: Banana{
				Name: "my banana",
			},
			wantPanic: true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.wantPanic {
					t.Errorf("got panic, did not want one")
				}
			}()
			block := gohcl.EncodeAsBlock(tt.banana, "banana")
			if block != nil {
				f := hclwrite.NewEmptyFile()
				f.Body().AppendBlock(block)
				gotHcl := string(hclwrite.Format(f.Bytes()))
				if gotHcl != tt.wantHcl {
					t.Errorf("got: %q, want: %q", gotHcl, tt.wantHcl)
				}
			}
		})
	}
}
