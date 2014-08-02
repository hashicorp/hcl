// This is the yacc input for creating the parser for HCL.

%{
package hcl

import (
	"github.com/hashicorp/hcl/ast"
)

%}

%union {
	list     []ast.Node
	klist    []ast.KeyedNode
	kitem    ast.KeyedNode
	listitem ast.Node
	num      int
	obj      ast.ObjectNode
	str      string
}

%type   <list> list
%type   <klist> objectlist
%type   <kitem> objectitem
%type   <listitem> listitem
%type   <obj> block object
%type   <str> blockId

%token  <num> NUMBER
%token  <str> COMMA IDENTIFIER EQUAL NEWLINE STRING
%token  <str> LEFTBRACE RIGHTBRACE LEFTBRACKET RIGHTBRACKET

%%

top:
	objectlist
	{
		hclResult = &ast.ObjectNode{
			K:    "",
			Elem: $1,
		}
	}

objectlist:
	objectitem
	{
		$$ = []ast.KeyedNode{$1}
	}
|	objectitem objectlist
	{
		$$ = append($2, $1)
	}

object:
	LEFTBRACE objectlist RIGHTBRACE
	{
		$$ = ast.ObjectNode{Elem: $2}
	}
|	LEFTBRACE RIGHTBRACE
	{
		$$ = ast.ObjectNode{}
	}

objectitem:
	IDENTIFIER EQUAL NUMBER
	{
		$$ = ast.AssignmentNode{
			K:     $1,
			Value: ast.LiteralNode{
				Type:  ast.ValueTypeInt,
				Value: $3,
			},
		}
	}
|	IDENTIFIER EQUAL STRING
	{
		$$ = ast.AssignmentNode{
			K:     $1,
			Value: ast.LiteralNode{
				Type:  ast.ValueTypeString,
				Value: $3,
			},
		}
	}
|	IDENTIFIER EQUAL object
	{
		$$ = ast.AssignmentNode{
			K:     $1,
			Value: $3,
		}
	}
|	IDENTIFIER EQUAL LEFTBRACKET list RIGHTBRACKET
	{
		$$ = ast.AssignmentNode{
			K:     $1,
			Value: ast.ListNode{Elem: $4},
		}
	}
|	block
	{
		$$ = $1
	}

block:
	blockId object
	{
		$$ = $2
		$$.K = $1
	}
|	blockId block
	{
		$$ = ast.ObjectNode{
			K:    $1,
			Elem: []ast.KeyedNode{$2},
		}
	}

blockId:
	IDENTIFIER
	{
		$$ = $1
	}
|	STRING
	{
		$$ = $1
	}

list:
	listitem
	{
		$$ = []ast.Node{$1}
	}
|	list COMMA listitem
	{
		$$ = append($1, $3)
	}

listitem:
	NUMBER
	{
		$$ = ast.LiteralNode{
			Type:  ast.ValueTypeInt,
			Value: $1,
		}
	}
|	STRING
	{
		$$ = ast.LiteralNode{
			Type:  ast.ValueTypeString,
			Value: $1,
		}
	}

%%
