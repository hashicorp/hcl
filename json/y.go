//line parse.y:3
package json

import __yyfmt__ "fmt"

//line parse.y:5
import (
	"github.com/hashicorp/hcl/ast"
)

//line parse.y:12
type jsonSymType struct {
	yys    int
	array  ast.ListNode
	assign ast.AssignmentNode
	item   ast.Node
	list   []ast.Node
	num    int
	str    string
	obj    ast.ObjectNode
}

const NUMBER = 57346
const COLON = 57347
const COMMA = 57348
const IDENTIFIER = 57349
const EQUAL = 57350
const NEWLINE = 57351
const STRING = 57352
const LEFTBRACE = 57353
const RIGHTBRACE = 57354
const LEFTBRACKET = 57355
const RIGHTBRACKET = 57356
const TRUE = 57357
const FALSE = 57358
const NULL = 57359

var jsonToknames = []string{
	"NUMBER",
	"COLON",
	"COMMA",
	"IDENTIFIER",
	"EQUAL",
	"NEWLINE",
	"STRING",
	"LEFTBRACE",
	"RIGHTBRACE",
	"LEFTBRACKET",
	"RIGHTBRACKET",
	"TRUE",
	"FALSE",
	"NULL",
}
var jsonStatenames = []string{}

const jsonEofCode = 1
const jsonErrCode = 2
const jsonMaxDepth = 200

//line parse.y:136

//line yacctab:1
var jsonExca = []int{
	-1, 1,
	1, -1,
	-2, 0,
}

const jsonNprod = 18
const jsonPrivate = 57344

var jsonTokenNames []string
var jsonStates []string

const jsonLast = 35

var jsonAct = []int{

	22, 14, 24, 7, 8, 5, 3, 13, 3, 14,
	20, 21, 17, 18, 19, 13, 3, 23, 20, 7,
	17, 18, 19, 4, 25, 9, 26, 10, 12, 15,
	2, 1, 6, 11, 16,
}
var jsonPact = []int{

	-5, -1000, -1000, -7, -8, -1000, 19, 22, -1000, 9,
	5, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-3, -1000, -12, 18, -1000, 5, -1000,
}
var jsonPgo = []int{

	0, 34, 32, 17, 0, 23, 29, 31,
}
var jsonR1 = []int{

	0, 7, 6, 6, 5, 5, 2, 3, 3, 3,
	3, 3, 3, 3, 1, 1, 4, 4,
}
var jsonR2 = []int{

	0, 1, 3, 2, 1, 3, 3, 1, 1, 1,
	1, 1, 1, 1, 2, 3, 1, 3,
}
var jsonChk = []int{

	-1000, -7, -6, 11, -5, 12, -2, 10, 12, 6,
	5, -5, -3, 10, 4, -6, -1, 15, 16, 17,
	13, 14, -4, -3, 14, 6, -4,
}
var jsonDef = []int{

	0, -2, 1, 0, 0, 3, 4, 0, 2, 0,
	0, 5, 6, 7, 8, 9, 10, 11, 12, 13,
	0, 14, 0, 16, 15, 0, 17,
}
var jsonTok1 = []int{

	1,
}
var jsonTok2 = []int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17,
}
var jsonTok3 = []int{
	0,
}

//line yaccpar:1

/*	parser for yacc output	*/

var jsonDebug = 0

type jsonLexer interface {
	Lex(lval *jsonSymType) int
	Error(s string)
}

const jsonFlag = -1000

func jsonTokname(c int) string {
	// 4 is TOKSTART above
	if c >= 4 && c-4 < len(jsonToknames) {
		if jsonToknames[c-4] != "" {
			return jsonToknames[c-4]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func jsonStatname(s int) string {
	if s >= 0 && s < len(jsonStatenames) {
		if jsonStatenames[s] != "" {
			return jsonStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func jsonlex1(lex jsonLexer, lval *jsonSymType) int {
	c := 0
	char := lex.Lex(lval)
	if char <= 0 {
		c = jsonTok1[0]
		goto out
	}
	if char < len(jsonTok1) {
		c = jsonTok1[char]
		goto out
	}
	if char >= jsonPrivate {
		if char < jsonPrivate+len(jsonTok2) {
			c = jsonTok2[char-jsonPrivate]
			goto out
		}
	}
	for i := 0; i < len(jsonTok3); i += 2 {
		c = jsonTok3[i+0]
		if c == char {
			c = jsonTok3[i+1]
			goto out
		}
	}

out:
	if c == 0 {
		c = jsonTok2[1] /* unknown char */
	}
	if jsonDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", jsonTokname(c), uint(char))
	}
	return c
}

func jsonParse(jsonlex jsonLexer) int {
	var jsonn int
	var jsonlval jsonSymType
	var jsonVAL jsonSymType
	jsonS := make([]jsonSymType, jsonMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	jsonstate := 0
	jsonchar := -1
	jsonp := -1
	goto jsonstack

ret0:
	return 0

ret1:
	return 1

jsonstack:
	/* put a state and value onto the stack */
	if jsonDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", jsonTokname(jsonchar), jsonStatname(jsonstate))
	}

	jsonp++
	if jsonp >= len(jsonS) {
		nyys := make([]jsonSymType, len(jsonS)*2)
		copy(nyys, jsonS)
		jsonS = nyys
	}
	jsonS[jsonp] = jsonVAL
	jsonS[jsonp].yys = jsonstate

jsonnewstate:
	jsonn = jsonPact[jsonstate]
	if jsonn <= jsonFlag {
		goto jsondefault /* simple state */
	}
	if jsonchar < 0 {
		jsonchar = jsonlex1(jsonlex, &jsonlval)
	}
	jsonn += jsonchar
	if jsonn < 0 || jsonn >= jsonLast {
		goto jsondefault
	}
	jsonn = jsonAct[jsonn]
	if jsonChk[jsonn] == jsonchar { /* valid shift */
		jsonchar = -1
		jsonVAL = jsonlval
		jsonstate = jsonn
		if Errflag > 0 {
			Errflag--
		}
		goto jsonstack
	}

jsondefault:
	/* default state action */
	jsonn = jsonDef[jsonstate]
	if jsonn == -2 {
		if jsonchar < 0 {
			jsonchar = jsonlex1(jsonlex, &jsonlval)
		}

		/* look through exception table */
		xi := 0
		for {
			if jsonExca[xi+0] == -1 && jsonExca[xi+1] == jsonstate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			jsonn = jsonExca[xi+0]
			if jsonn < 0 || jsonn == jsonchar {
				break
			}
		}
		jsonn = jsonExca[xi+1]
		if jsonn < 0 {
			goto ret0
		}
	}
	if jsonn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			jsonlex.Error("syntax error")
			Nerrs++
			if jsonDebug >= 1 {
				__yyfmt__.Printf("%s", jsonStatname(jsonstate))
				__yyfmt__.Printf(" saw %s\n", jsonTokname(jsonchar))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for jsonp >= 0 {
				jsonn = jsonPact[jsonS[jsonp].yys] + jsonErrCode
				if jsonn >= 0 && jsonn < jsonLast {
					jsonstate = jsonAct[jsonn] /* simulate a shift of "error" */
					if jsonChk[jsonstate] == jsonErrCode {
						goto jsonstack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if jsonDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", jsonS[jsonp].yys)
				}
				jsonp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if jsonDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", jsonTokname(jsonchar))
			}
			if jsonchar == jsonEofCode {
				goto ret1
			}
			jsonchar = -1
			goto jsonnewstate /* try again in the same state */
		}
	}

	/* reduction by production jsonn */
	if jsonDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", jsonn, jsonStatname(jsonstate))
	}

	jsonnt := jsonn
	jsonpt := jsonp
	_ = jsonpt // guard against "declared and not used"

	jsonp -= jsonR2[jsonn]
	jsonVAL = jsonS[jsonp+1]

	/* consult goto table to find next state */
	jsonn = jsonR1[jsonn]
	jsong := jsonPgo[jsonn]
	jsonj := jsong + jsonS[jsonp].yys + 1

	if jsonj >= jsonLast {
		jsonstate = jsonAct[jsong]
	} else {
		jsonstate = jsonAct[jsonj]
		if jsonChk[jsonstate] != -jsonn {
			jsonstate = jsonAct[jsong]
		}
	}
	// dummy call; replaced with literal code
	switch jsonnt {

	case 1:
		//line parse.y:37
		{
			obj := jsonS[jsonpt-0].obj
			jsonResult = &obj
		}
	case 2:
		//line parse.y:44
		{
			jsonVAL.obj = ast.ObjectNode{Elem: jsonS[jsonpt-1].list}
		}
	case 3:
		//line parse.y:48
		{
			jsonVAL.obj = ast.ObjectNode{}
		}
	case 4:
		//line parse.y:54
		{
			jsonVAL.list = []ast.Node{jsonS[jsonpt-0].assign}
		}
	case 5:
		//line parse.y:58
		{
			jsonVAL.list = append(jsonS[jsonpt-0].list, jsonS[jsonpt-2].assign)
		}
	case 6:
		//line parse.y:64
		{
			jsonVAL.assign = ast.AssignmentNode{
				Key:   jsonS[jsonpt-2].str,
				Value: jsonS[jsonpt-0].item,
			}
		}
	case 7:
		//line parse.y:73
		{
			jsonVAL.item = ast.LiteralNode{
				Type:  ast.ValueTypeString,
				Value: jsonS[jsonpt-0].str,
			}
		}
	case 8:
		//line parse.y:80
		{
			jsonVAL.item = ast.LiteralNode{
				Type:  ast.ValueTypeInt,
				Value: jsonS[jsonpt-0].num,
			}
		}
	case 9:
		//line parse.y:87
		{
			jsonVAL.item = jsonS[jsonpt-0].obj
		}
	case 10:
		//line parse.y:91
		{
			jsonVAL.item = jsonS[jsonpt-0].array
		}
	case 11:
		//line parse.y:95
		{
			jsonVAL.item = ast.LiteralNode{
				Type:  ast.ValueTypeBool,
				Value: true,
			}
		}
	case 12:
		//line parse.y:102
		{
			jsonVAL.item = ast.LiteralNode{
				Type:  ast.ValueTypeBool,
				Value: false,
			}
		}
	case 13:
		//line parse.y:109
		{
			jsonVAL.item = ast.LiteralNode{
				Type:  ast.ValueTypeNil,
				Value: nil,
			}
		}
	case 14:
		//line parse.y:118
		{
			jsonVAL.array = ast.ListNode{}
		}
	case 15:
		//line parse.y:122
		{
			jsonVAL.array = ast.ListNode{Elem: jsonS[jsonpt-1].list}
		}
	case 16:
		//line parse.y:128
		{
			jsonVAL.list = []ast.Node{jsonS[jsonpt-0].item}
		}
	case 17:
		//line parse.y:132
		{
			jsonVAL.list = append(jsonS[jsonpt-0].list, jsonS[jsonpt-2].item)
		}
	}
	goto jsonstack /* stack new state and value */
}
