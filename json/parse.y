// This is the yacc input for creating the parser for HCL JSON.

%{
package json

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/token"
)

%}

%union {
	f        float64
	list     []ast.Node
	node     ast.Node
	num      int
	str      string
	obj      *ast.ObjectType
	objitem  *ast.ObjectItem
	objlist  *ast.ObjectList
}

%type	<f> float
%type	<list> array elements
%type   <node> number value
%type	<num> int
%type	<obj> object
%type	<objitem> pair
%type	<objlist> members
%type	<str> exp

%token  <f> FLOAT
%token  <num> NUMBER
%token  <str> COLON COMMA IDENTIFIER EQUAL NEWLINE STRING
%token  <str> LEFTBRACE RIGHTBRACE LEFTBRACKET RIGHTBRACKET
%token  <str> TRUE FALSE NULL MINUS PERIOD EPLUS EMINUS

%%

top:
	object
	{
		jsonResult = &ast.File{
			Node: $1.List,
		}
	}

object:
	LEFTBRACE members RIGHTBRACE
	{
		$$ = &ast.ObjectType{
			List: $2,
		}
	}
|	LEFTBRACE RIGHTBRACE
	{
		$$ = &ast.ObjectType{}
	}

members:
	pair
	{
		$$ = &ast.ObjectList{
			Items: []*ast.ObjectItem{$1},
		}
	}
|	members COMMA pair
	{
		$1.Items = append($1.Items, $3)
		$$ = $1
	}

pair:
	STRING COLON value
	{
		$$ = &ast.ObjectItem{
			Keys: []*ast.ObjectKey{
				&ast.ObjectKey{
					Token: token.Token{
						Type: token.IDENT,
						Text: $1,
					},
				},
			},

			Val: $3,
		}
	}

value:
	STRING
	{
		$$ = &ast.LiteralType{
			Token: token.Token{
				Type: token.STRING,
				Text: fmt.Sprintf(`"%s"`, $1),
			},
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
		$$ = &ast.ListType{
			List: $1,
		}
	}
|	TRUE
	{
		$$ = &ast.LiteralType{
			Token: token.Token{Type: token.BOOL, Text: "true"},
		}
	}
|	FALSE
	{
		$$ = &ast.LiteralType{
			Token: token.Token{Type: token.BOOL, Text: "false"},
		}
	}
|	NULL
	{
		$$ = &ast.LiteralType{
			Token: token.Token{Type: token.STRING, Text: ""},
		}
	}

array:
	LEFTBRACKET RIGHTBRACKET
	{
		$$ = nil
	}
|	LEFTBRACKET elements RIGHTBRACKET
	{
		$$ = $2
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
		$$ = &ast.LiteralType{
			Token: token.Token{
				Type: token.NUMBER,
				Text: fmt.Sprintf("%d", $1),
			},
		}
	}
|	float
	{
		$$ = &ast.LiteralType{
			Token: token.Token{
				Type: token.FLOAT,
				Text: fmt.Sprintf("%f", $1),
			},
		}
	}
|   int exp
    {
		fs := fmt.Sprintf("%d%s", $1, $2)
		$$ = &ast.LiteralType{
			Token: token.Token{
				Type: token.FLOAT,
				Text: fs,
			},
		}
    }
|   float exp
    {
		fs := fmt.Sprintf("%f%s", $1, $2)
		$$ = &ast.LiteralType{
			Token: token.Token{
				Type: token.FLOAT,
				Text: fs,
			},
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

float:
	 MINUS float
	{
		$$ = $2 * -1
	}
|	FLOAT
	{
		$$ = $1
	}

exp:
    EPLUS NUMBER
    {
        $$ = "e" + strconv.FormatInt(int64($2), 10)
    }
|   EMINUS NUMBER
    {
        $$ = "e-" + strconv.FormatInt(int64($2), 10)
    }

%%
