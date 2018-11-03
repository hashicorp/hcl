package hclwrite_test

import (
	"fmt"

	"github.com/hashicorp/hcl2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func Example_generateFromScratch() {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	rootBody.SetAttributeValue("string", cty.StringVal("bar")) // this is overwritten later
	rootBody.AppendNewline()
	rootBody.SetAttributeValue("object", cty.ObjectVal(map[string]cty.Value{
		"foo": cty.StringVal("foo"),
		"bar": cty.NumberIntVal(5),
		"baz": cty.True,
	}))
	rootBody.SetAttributeValue("string", cty.StringVal("foo"))
	rootBody.SetAttributeValue("bool", cty.False)
	rootBody.AppendNewline()
	fooBlock := rootBody.AppendNewBlock("foo", nil)
	fooBody := fooBlock.Body()
	rootBody.AppendNewBlock("empty", nil)
	rootBody.AppendNewline()
	barBlock := rootBody.AppendNewBlock("bar", []string{"a", "b"})
	barBody := barBlock.Body()

	fooBody.SetAttributeValue("hello", cty.StringVal("world"))

	bazBlock := barBody.AppendNewBlock("baz", nil)
	bazBody := bazBlock.Body()
	bazBody.SetAttributeValue("foo", cty.NumberIntVal(10))
	bazBody.SetAttributeValue("beep", cty.StringVal("boop"))
	bazBody.SetAttributeValue("baz", cty.ListValEmpty(cty.String))

	fmt.Printf("%s", f.Bytes())
	// Output:
	// string = "foo"
	//
	// object = {bar = 5, baz = true, foo = "foo"}
	// bool   = false
	//
	// foo {
	//   hello = "world"
	// }
	// empty {
	// }
	//
	// bar "a" "b" {
	//   baz {
	//     foo  = 10
	//     beep = "boop"
	//     baz  = []
	//   }
	// }
}
