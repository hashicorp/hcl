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

%type   <list> list objectlist
%type   <listitem> listitem objectitem
%type   <obj> block object
%type   <str> blockId

%token  <num> NUMBER
%token  <str> COMMA IDENTIFIER EQUAL NEWLINE STRING
%token  <str> LEFTBRACE RIGHTBRACE LEFTBRACKET RIGHTBRACKET

%%

top:
	objectlist
	{
		hclResult = &ObjectNode{
			Key:  "",
			Elem: $1,
		}
	}

objectlist:
	objectitem
	{
		$$ = []Node{$1}
	}
|	objectitem objectlist
	{
		$$ = append($2, $1)
	}

object:
	LEFTBRACE objectlist RIGHTBRACE
	{
		$$ = ObjectNode{Elem: $2}
	}
|	LEFTBRACE RIGHTBRACE
	{
		$$ = ObjectNode{}
	}

objectitem:
	IDENTIFIER EQUAL NUMBER
	{
		$$ = AssignmentNode{
			Key:   $1,
			Value: LiteralNode{
				Type:  ValueTypeInt,
				Value: $3,
			},
		}
	}
|	IDENTIFIER EQUAL STRING
	{
		$$ = AssignmentNode{
			Key:   $1,
			Value: LiteralNode{
				Type:  ValueTypeString,
				Value: $3,
			},
		}
	}
|	IDENTIFIER EQUAL object
	{
		$$ = AssignmentNode{
			Key:   $1,
			Value: $3,
		}
	}
|	IDENTIFIER EQUAL LEFTBRACKET list RIGHTBRACKET
	{
		$$ = AssignmentNode{
			Key:   $1,
			Value: ListNode{Elem: $4},
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
		$$.Key = $1
	}
|	blockId block
	{
		$$ = ObjectNode{
			Key:  $1,
			Elem: []Node{$2},
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
	NUMBER
	{
		$$ = LiteralNode{
			Type:  ValueTypeInt,
			Value: $1,
		}
	}
|	STRING
	{
		$$ = LiteralNode{
			Type:  ValueTypeString,
			Value: $1,
		}
	}

%%
