// This is the yacc input for creating the parser for HCL.

%{
package hcl

%}

%union {
	num int
	obj map[string]interface{}
	str string
}

%type   <obj> object

%token  <num> NUMBER
%token  <str> IDENTIFIER EQUAL SEMICOLON STRING
%token  <str> LEFTBRACE RIGHTBRACE

%%

top:
	object
	{
		exprResult = []map[string]interface{}{$1}
	}

object:
	IDENTIFIER EQUAL STRING
	{
		$$ = map[string]interface{}{$1: $3}
	}

%%
