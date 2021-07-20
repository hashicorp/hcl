package gohcl

import (
	"fmt"
	"reflect"
	"sort"
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

	if ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
	}

	if ty.Kind() != reflect.Struct {
		panic(fmt.Sprintf("given value must be struct, not %T", val))
	}

	var attrSchemas []hcl.AttributeSchema
	var blockSchemas []hcl.BlockHeaderSchema

	tags := getFieldTags(ty)

	attrNames := make([]string, 0, len(tags.Attributes))
	for n := range tags.Attributes {
		attrNames = append(attrNames, n)
	}
	sort.Strings(attrNames)
	for _, n := range attrNames {
		for _, fieldInfo := range tags.Attributes[n] {
			optional := fieldInfo.optional
			field := fieldInfo.path.getTypeField(ty).Type

			var required bool

			switch {
			case field.AssignableTo(exprType):
				// If we're decoding to hcl.Expression then absence can be
				// indicated via a null value, so we don't specify that
				// the field is required during decoding.
				required = false
			case field.Kind() != reflect.Ptr && !optional:
				required = true
			default:
				required = false
			}

			attrSchemas = append(attrSchemas, hcl.AttributeSchema{
				Name:     n,
				Required: required,
			})
		}
	}

	blockNames := make([]string, 0, len(tags.Blocks))
	for n := range tags.Blocks {
		blockNames = append(blockNames, n)
	}
	sort.Strings(blockNames)
	for _, n := range blockNames {
		for _, block := range tags.Blocks[n] {
			field := block.getTypeField(ty)
			fty := field.Type
			if fty.Kind() == reflect.Slice {
				fty = fty.Elem()
			}
			if fty.Kind() == reflect.Ptr {
				fty = fty.Elem()
			}
			if fty.Kind() != reflect.Struct {
				panic(fmt.Sprintf(
					"hcl 'block' tag kind cannot be applied to %s field %s: struct required", fty.String(), field.Name,
				))
			}
			ftags := getFieldTags(fty)
			var labelNames []string
			for _, l := range ftags.Labels {
				labelNames = append(labelNames, l.Name)
			}

			blockSchemas = append(blockSchemas, hcl.BlockHeaderSchema{
				Type:       n,
				LabelNames: labelNames,
			})
		}
	}

	partial = tags.Remain != nil
	schema = &hcl.BodySchema{
		Attributes: attrSchemas,
		Blocks:     blockSchemas,
	}
	return schema, partial
}

type fieldTags struct {
	Attributes map[string]fieldInfoList
	Blocks     map[string][]fieldPath
	Labels     []labelField
	Remain     fieldPath
	Body       fieldPath
}

type labelField struct {
	FieldIndex fieldPath
	Name       string
}

type fieldInfoList []fieldInfo

func (fields fieldInfoList) IsOptional() bool {
	// If the same attribute is associated to many fields, we consider
	// it as non optional as soon as one of them is not optional.
	for _, info := range fields {
		if !info.optional {
			return false
		}
	}
	return true
}

type fieldInfo struct {
	optional bool
	path     fieldPath
}

// If the actual field comes from a squashed structure, the path is defined as the
// indexes of the field in all structures that are squashed into the main struct
type fieldPath []int

// This method retrive the actual field type by following the path from the original type
// up to the final field.
func (path fieldPath) getTypeField(ty reflect.Type) (result reflect.StructField) {
	for _, idx := range path {
		result = ty.Field(idx)
		ty = result.Type
	}
	return
}

// This method retrive the actual field value by following the path from the original type
// up to the final field.
func (path fieldPath) getValueField(val reflect.Value) reflect.Value {
	for _, idx := range path {
		val = val.Field(idx)
	}
	return val
}

func getFieldTags(ty reflect.Type) *fieldTags {
	return getFieldTagsInternal(&fieldTags{
		Attributes: map[string]fieldInfoList{},
		Blocks:     map[string][]fieldPath{},
	}, ty)
}

func getFieldTagsInternal(ret *fieldTags, ty reflect.Type, parents ...int) *fieldTags {
	// parents indicates that the supplied type is squashed within another types

	ct := ty.NumField()
	for i := 0; i < ct; i++ {
		field := ty.Field(i)
		tag := field.Tag.Get("hcl")
		if tag == "" {
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

		// Ensure that path to the attribute is a new array to avoid buffer overriding
		path := append([]int{}, append(parents, i)...)

		switch kind {
		case "optional", "attr":
			ret.Attributes[name] = append(ret.Attributes[name], fieldInfo{
				optional: kind == "optional",
				path:     path,
			})
		case "squash":
			getFieldTagsInternal(ret, ty.Field(i).Type, path...)
		case "block":
			ret.Blocks[name] = append(ret.Blocks[name], path)
		case "label":
			ret.Labels = append(ret.Labels, labelField{
				FieldIndex: path,
				Name:       name,
			})
		case "remain":
			if ret.Remain != nil {
				panic("only one 'remain' tag is permitted")
			}
			ret.Remain = path
		case "body":
			if ret.Body != nil {
				panic("only one 'body' tag is permitted")
			}
			ret.Body = path
		default:
			panic(fmt.Sprintf("invalid hcl field tag kind %q on %s %q", kind, field.Type.String(), field.Name))
		}
	}

	return ret
}
