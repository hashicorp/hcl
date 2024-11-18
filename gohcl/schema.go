// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gohcl

import (
	"cmp"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

// ImpliedBodySchema produces a hcl.BodySchema derived from the type of the
// given value, which must be a struct value or a pointer to one. If an
// inappropriate value is passed, this function will panic.
//
// The second return argument indicates whether the given struct includes
// a "remain" field, and thus the returned schema is non-exhaustive.
//
// This uses the tags on the fields of the struct to discover how each
// field's value should be expressed within configuration. If an invalid
// mapping is attempted, this function will panic.
func ImpliedBodySchema(val interface{}) (schema *hcl.BodySchema, partial bool) {
	ty := reflect.TypeOf(val)
	if nty, ok := val.(reflect.Value); ok {
		// Recursion through embedded structs
		ty = nty.Type()
	}

	if ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
	}

	if ty.Kind() != reflect.Struct {
		panic(fmt.Sprintf("given value must be struct, not %T", val))
	}

	var attrSchemas []hcl.AttributeSchema
	var blockSchemas []hcl.BlockHeaderSchema

	tags := getFieldTags(ty)

	for n := range tags.Attributes {
		idx := tags.Attributes[n]
		optional := tags.Optional[n]
		field := ty.Field(idx)

		var required bool

		switch {
		case field.Type.AssignableTo(exprType):
			// If we're decoding to hcl.Expression then absense can be
			// indicated via a null value, so we don't specify that
			// the field is required during decoding.
			required = false
		case field.Type.Kind() != reflect.Ptr && !optional:
			required = true
		default:
			required = false
		}

		attrSchemas = append(attrSchemas, hcl.AttributeSchema{
			Name:     n,
			Required: required,
		})
	}

	for n := range tags.Blocks {
		idx := tags.Blocks[n]
		field := ty.Field(idx)
		fty := field.Type
		if fty.Kind() == reflect.Slice {
			fty = fty.Elem()
		}
		if fty.Kind() == reflect.Ptr {
			fty = fty.Elem()
		}
		if fty.Kind() != reflect.Struct {
			panic(fmt.Sprintf(
				"hcl 'block' tag kind cannot be applied to %s field %s: struct required", field.Type.String(), field.Name,
			))
		}
		ftags := getFieldTags(fty)
		var labelNames []string
		if len(ftags.Labels) > 0 {
			labelNames = make([]string, len(ftags.Labels))
			for i, l := range ftags.Labels {
				labelNames[i] = l.Name
			}
		}

		blockSchemas = append(blockSchemas, hcl.BlockHeaderSchema{
			Type:       n,
			LabelNames: labelNames,
		})
	}

	partial = tags.Remain != nil

	for _, embedded := range tags.Embedded {
		nested, npartial := ImpliedBodySchema(reflect.New(embedded.Type))
		if npartial && partial {
			panic("only one 'remain' tag is permitted (nested)")
		}

		attrSchemas = append(attrSchemas, nested.Attributes...)
		blockSchemas = append(blockSchemas, nested.Blocks...)
		println(fmt.Sprintf("%v: %#v\n", embedded.Type, nested))
	}

	slices.SortStableFunc(attrSchemas, func(a, b hcl.AttributeSchema) int {
		return cmp.Compare(a.Name, b.Name)
	})
	slices.SortStableFunc(blockSchemas, func(a, b hcl.BlockHeaderSchema) int {
		return cmp.Compare(a.Type, b.Type)
	})

	schema = &hcl.BodySchema{
		Attributes: attrSchemas,
		Blocks:     blockSchemas,
	}
	return schema, partial
}

type fieldTags struct {
	Attributes map[string]int
	Blocks     map[string]int
	Labels     []labelField
	Remain     *int
	Body       *int
	Optional   map[string]bool
	Embedded   []embeddedField
}

type embeddedField struct {
	FieldIndex int
	Type       reflect.Type
}

type labelField struct {
	FieldIndex int
	Name       string
}

func getFieldTags(ty reflect.Type) *fieldTags {
	ret := &fieldTags{
		Attributes: map[string]int{},
		Blocks:     map[string]int{},
		Optional:   map[string]bool{},
	}

	ct := ty.NumField()
	for i := 0; i < ct; i++ {
		field := ty.Field(i)

		tag := field.Tag.Get("hcl")
		if tag == "" {
			if field.Type.Kind() == reflect.Struct && field.Anonymous {
				ret.Embedded = append(ret.Embedded, embeddedField{
					FieldIndex: i,
					Type:       field.Type,
				})
			}
			continue
		}

		comma := strings.Index(tag, ",")
		var name, kind string
		if comma != -1 {
			name = tag[:comma]
			kind = tag[comma+1:]
		} else {
			name = tag
			kind = "attr"
		}

		switch kind {
		case "attr":
			ret.Attributes[name] = i
		case "block":
			ret.Blocks[name] = i
		case "label":
			ret.Labels = append(ret.Labels, labelField{
				FieldIndex: i,
				Name:       name,
			})
		case "remain":
			if ret.Remain != nil {
				panic("only one 'remain' tag is permitted")
			}
			idx := i // copy, because this loop will continue assigning to i
			ret.Remain = &idx
		case "body":
			if ret.Body != nil {
				panic("only one 'body' tag is permitted")
			}
			idx := i // copy, because this loop will continue assigning to i
			ret.Body = &idx
		case "optional":
			ret.Attributes[name] = i
			ret.Optional[name] = true
		default:
			panic(fmt.Sprintf("invalid hcl field tag kind %q on %s %q", kind, field.Type.String(), field.Name))
		}
	}

	return ret
}
