// This is the yacc input for creating the parser for HCL JSON.

%{
package json

import (
	"github.com/hashicorp/hcl/ast"
)

%}

%union {
	array    ast.ListNode
	assign   ast.AssignmentNode
	item     ast.Node
	list     []ast.Node
	num      int
	str      string
	obj      ast.ObjectNode
}

%type	<array> array
%type	<assign> pair
%type	<item> value
%type	<list> elements members
%type	<obj> object

%token  <num> NUMBER
%token  <str> COLON COMMA IDENTIFIER EQUAL NEWLINE STRING
%token  <str> LEFTBRACE RIGHTBRACE LEFTBRACKET RIGHTBRACKET
%token  <str> TRUE FALSE NULL

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
		$$ = []ast.Node{$1}
	}
|	pair COMMA members
	{
		$$ = append($3, $1)
	}

pair:
	STRING COLON value
	{
		$$ = ast.AssignmentNode{
			Key:   $1,
			Value: $3,
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
|	NUMBER
	{
		$$ = ast.LiteralNode{
			Type:  ast.ValueTypeInt,
			Value: $1,
		}
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
|	value COMMA elements
	{
		$$ = append($3, $1)
	}

%%
