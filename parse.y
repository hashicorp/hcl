// This is the yacc input for creating the parser for HCL.

%{
package hcl

%}

%union {
	list []Node
	listitem Node
	num int
	obj ObjectNode
	str string
}

%type   <list> list
%type   <listitem> listitem
%type   <obj> block object objectlist
%type   <str> blockId

%token  <num> NUMBER
%token  <str> COMMA IDENTIFIER EQUAL NEWLINE STRING
%token  <str> LEFTBRACE RIGHTBRACE LEFTBRACKET RIGHTBRACKET

%%

top:
	objectlist
	{
		hclResult = &ObjectNode{
			Elem: $1.Elem,
		}
	}

objectlist:
	object
	{
		$$ = $1
	}
|	object objectlist
	{
		$$ = $1
		for k, v := range $2.Elem {
			if _, ok := $$.Elem[k]; ok {
				$$.Elem[k] = append($$.Elem[k], v...)
			} else {
				$$.Elem[k] = v
			}
		}
	}

object:
	IDENTIFIER EQUAL NUMBER
	{
		$$ = ObjectNode{
			Elem: map[string][]Node{
				$1: []Node{
					ValueNode{
						Type:  ValueTypeInt,
						Value: $3,
					},
				},
			},
		}
	}
|	IDENTIFIER EQUAL STRING
	{
		$$ = ObjectNode{
			Elem: map[string][]Node{
				$1: []Node{
					ValueNode{
						Type:  ValueTypeString,
						Value: $3,
					},
				},
			},
		}
	}
|	IDENTIFIER EQUAL LEFTBRACKET list RIGHTBRACKET
	{
		$$ = ObjectNode{
			Elem: map[string][]Node{
				$1: $4,
			},
		}
	}
|	block
	{
		$$ = $1
	}

block:
	blockId LEFTBRACE objectlist RIGHTBRACE
	{
		$$ = ObjectNode{
			Elem: map[string][]Node{
				$1: []Node{$3},
			},
		}
	}
|	blockId block
	{
		$$ = ObjectNode{
			Elem: map[string][]Node{
				$1: []Node{$2},
			},
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
		$$ = []Node{$1}
	}
|	list COMMA listitem
	{
		$$ = append($1, $3)
	}

listitem:
	object
	{
		$$ = $1
	}
|	NUMBER
	{
		$$ = $1
	}
|	STRING
	{
		$$ = $1
	}

%%
