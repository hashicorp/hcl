// This is the yacc input for creating the parser for HCL JSON.

%{
package json

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/hcl/ast"
)

%}

%union {
	array    ast.ListNode
	assign   ast.AssignmentNode
	item     ast.Node
	klist    []ast.AssignmentNode
	list     []ast.Node
	num      int
	str      string
	obj      ast.ObjectNode
}

%type	<array> array
%type	<assign> pair
%type	<item> value number
%type	<klist> members
%type	<list> elements
%type	<num> int
%type	<obj> object
%type	<str> frac

%token  <num> NUMBER
%token  <str> COLON COMMA IDENTIFIER EQUAL NEWLINE STRING
%token  <str> LEFTBRACE RIGHTBRACE LEFTBRACKET RIGHTBRACKET
%token  <str> TRUE FALSE NULL MINUS PERIOD

%%

top:
	object
	{
		obj := $1
		jsonResult = &obj
	}

object:
	LEFTBRACE members RIGHTBRACE
	{
		$$ = ast.ObjectNode{Elem: $2}
	}
|	LEFTBRACE RIGHTBRACE
	{
		$$ = ast.ObjectNode{}
	}

members:
	pair
	{
		$$ = []ast.AssignmentNode{$1}
	}
|	members COMMA pair
	{
		$$ = append($1, $3)
	}

pair:
	STRING COLON value
	{
		value := $3
		if obj, ok := value.(ast.ObjectNode); ok {
			obj.K = $1
			value = obj
		}

		$$ = ast.AssignmentNode{
			K:     $1,
			Value: value,
		}
	}

value:
	STRING
	{
		$$ = ast.LiteralNode{
			Type:  ast.ValueTypeString,
			Value: $1,
		}
	}
|	number
	{
		$$ = $1
	}
|	object
	{
		$$ = $1
	}
|	array
	{
		$$ = $1
	}
|	TRUE
	{
		$$ = ast.LiteralNode{
			Type:  ast.ValueTypeBool,
			Value: true,
		}
	}
|	FALSE
	{
		$$ = ast.LiteralNode{
			Type:  ast.ValueTypeBool,
			Value: false,
		}
	}
|	NULL
	{
		$$ = ast.LiteralNode{
			Type:  ast.ValueTypeNil,
			Value: nil,
		}
	}

array:
	LEFTBRACKET RIGHTBRACKET
	{
		$$ = ast.ListNode{}
	}
|	LEFTBRACKET elements RIGHTBRACKET
	{
		$$ = ast.ListNode{Elem: $2}
	}

elements:
	value
	{
		$$ = []ast.Node{$1}
	}
|	elements COMMA value
	{
		$$ = append($1, $3)
	}

number:
	int
	{
		$$ = ast.LiteralNode{
			Type:  ast.ValueTypeInt,
			Value: $1,
		}
	}
|	int frac
	{
		fs := fmt.Sprintf("%d.%s", $1, $2)
		f, err := strconv.ParseFloat(fs, 64)
		if err != nil {
			panic(err)
		}

		$$ = ast.LiteralNode{
			Type:  ast.ValueTypeFloat,
			Value: f,
		}
	}

int:
	MINUS int
	{
		$$ = $2 * -1
	}
|	NUMBER
	{
		$$ = $1
	}

frac:
	PERIOD NUMBER
	{
		$$ = strconv.FormatInt(int64($2), 10)
	}

%%
