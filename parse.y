// This is the yacc input for creating the parser for HCL.

%{
package hcl

%}

%union {
	num int
	obj map[string]interface{}
	str string
}

%type   <obj> block object
%type   <str> blockId

%token  <num> NUMBER
%token  <str> IDENTIFIER EQUAL SEMICOLON STRING
%token  <str> LEFTBRACE RIGHTBRACE

%%

top:
	object
	{
		hclResult = $1
	}

object:
	object SEMICOLON
	{
		$$ = $1
	}
|	IDENTIFIER EQUAL NUMBER
	{
		$$ = map[string]interface{}{$1: []interface{}{$3}}
	}
|	IDENTIFIER EQUAL STRING
	{
		$$ = map[string]interface{}{$1: []interface{}{$3}}
	}
|	block
	{
		$$ = $1
	}

block:
	blockId LEFTBRACE object RIGHTBRACE
	{
		$$ = map[string]interface{}{$1: []interface{}{$3}}
	}
|	blockId block
	{
		$$ = map[string]interface{}{$1: []interface{}{$2}}
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

%%
