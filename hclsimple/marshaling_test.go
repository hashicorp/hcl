package hclsimple_test

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
)

func ExampleMarshal_simpleSquash() {
	type A struct {
		Name   string `hcl:"name,label"`
		AField string `hcl:"a"`
	}

	type B struct {
		A      `hcl:",squash"`
		BField string `hcl:"b,optional"`
	}

	type data struct {
		Title string `hcl:"title"`
		Array []B    `hcl:"array,block"`
	}

	content := `
	title = "test"

	array "a1" {
		a = "foo1"
		b = "bar1"
	}

	array "a2" {
		a = "foo2"
	}
	`

	// Fill the structure with the content.
	var got data
	if err := hclsimple.Unmarshal([]byte(content), &got); err != nil {
		fmt.Println(err)
	}

	// Then render the structure as hcl.
	output, _ := hclsimple.Marshal(got)
	fmt.Println(string(output))

	// Output:
	// title = "test"
	//
	// array "a1" {
	//   a = "foo1"
	//   b = "bar1"
	// }
	// array "a2" {
	//   a = "foo2"
	// }
}

func ExampleMarshal_withDoubleEmbedded() {
	type A struct {
		AName  string `hcl:"a_name,label"`
		AField string `hcl:"a"`
	}

	type B struct {
		A      `hcl:",squash"`
		BName  string `hcl:"b_name,label"`
		BField string `hcl:"b"` // This field will be populated twice
	}

	type C struct {
		B      `hcl:",squash"`
		CField string `hcl:"c,optional"`
	}

	type data struct {
		Title string `hcl:"title"`
		Array []C    `hcl:"array,block"`
	}

	content := `
	array "a2" "b2" {
		b = "bar2"
		a = "foo2"
	}
	array "a1" "b1" {
		c = "That's all folks!"
		b = "bar1"
		a = "foo1"
	}
	title = "test"
	`

	// Fill the structure with the content.
	var got data
	if err := hclsimple.Unmarshal([]byte(content), &got); err != nil {
		fmt.Println(err)
	}

	// Then render the structure as hcl.
	// Note that the fields are reordered, but the array order remains the same.
	if result, err := hclsimple.Marshal(got); err == nil {
		fmt.Println(string(result))
	} else {
		fmt.Println(err)
	}

	// Output:
	// title = "test"
	//
	// array "a2" "b2" {
	//   a = "foo2"
	//   b = "bar2"
	// }
	// array "a1" "b1" {
	//   a = "foo1"
	//   b = "bar1"
	//   c = "That's all folks!"
	// }
}

func ExampleMarshal_complex() {
	type A struct {
		AName   string `hcl:"a_name,label"`
		AField  string `hcl:"a"`
		SecondB string `hcl:"b"`          // This field will be populated twice
		SecondC string `hcl:"c,optional"` // This field will be populated three times
	}

	type B struct {
		A `hcl:",squash"`

		BName  string `hcl:"b_name,label"`
		BField string `hcl:"b"` // This field will be populated twice
	}

	// Private type
	type c struct {
		CField string `hcl:"c,optional"` // This field will be populated three times
	}

	type D struct {
		B `hcl:",squash"` // Anonymous embedded structure can be squashed

		DField int    `hcl:"d,optional"`
		Bool1  bool   `hcl:"bool_1"`
		Bool2  bool   `hcl:"bool_2,optional"`
		ThirdC string `hcl:"c,optional"`           // This field will be populated three times
		C      *c     `hcl:"optional_block,block"` // A block can be made optional using a pointer
		C2     c      `hcl:",squash"`              // private embedded struct can also be squashed but need a public name
	}

	type data struct {
		Title string `hcl:"title"`
		Array []D    `hcl:"array,block"`
	}

	content := `
	array "a2" "b2" {
		b      = "bar2"
		a      = "foo2"
		d      = 123
		bool_1 = true
		bool_2 = false
	}

	array "a1" "b1" {
		optional_block { c = "hello" }
		c      = "That's all folks!"
		b      = "bar1"
		a      = "foo1"
		bool_1 = false
		bool_2 = true
	}

	title = "test"
	`

	// Fill the structure with the content.
	var got data
	if err := hclsimple.Unmarshal([]byte(content), &got); err != nil {
		for _, err := range err.(hcl.Diagnostics) {
			fmt.Println(err)
		}
	}

	// As expected, the b value filled both BField and SecondB.
	a0 := got.Array[0]
	fmt.Printf("BField='%s' SecondB='%s'\n", a0.BField, a0.SecondB)

	// As expected, the c value filled CField, SecondC and ThirdC.
	a1 := got.Array[1]
	fmt.Printf("CField='%s' all equal=%t\n\n", a1.C2.CField, a1.C2.CField == a1.ThirdC && a1.C2.CField == a1.SecondC)

	// Then render the structure as hcl.
	// Note that the fields are reordered, but the array order is preserved.
	if result, err := hclsimple.Marshal(got); err == nil {
		fmt.Println(string(result))
	} else {
		fmt.Println(err)
	}

	// Output:
	// BField='bar2' SecondB='bar2'
	// CField='That's all folks!' all equal=true
	//
	// title = "test"
	//
	// array "a2" "b2" {
	//   a      = "foo2"
	//   b      = "bar2"
	//   d      = 123
	//   bool_1 = true
	// }
	// array "a1" "b1" {
	//   a      = "foo1"
	//   b      = "bar1"
	//   c      = "That's all folks!"
	//   bool_1 = false
	//   bool_2 = true
	//
	//   optional_block {
	//     c = "hello"
	//   }
	// }
}
