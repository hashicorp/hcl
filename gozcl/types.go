package gozcl

import (
	"reflect"

	"github.com/hashicorp/hcl2/zcl"
)

var victimExpr zcl.Expression
var victimBody zcl.Body

var exprType = reflect.TypeOf(&victimExpr).Elem()
var bodyType = reflect.TypeOf(&victimBody).Elem()
var blockType = reflect.TypeOf((*zcl.Block)(nil))
var attrType = reflect.TypeOf((*zcl.Attribute)(nil))
var attrsType = reflect.TypeOf(zcl.Attributes(nil))
