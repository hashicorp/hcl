//line parse.y:4
package hcl

import __yyfmt__ "fmt"

//line parse.y:4
import (
	"github.com/hashicorp/hcl/ast"
)

//line parse.y:12
type hclSymType struct {
	yys      int
	list     []ast.Node
	klist    []ast.KeyedNode
	kitem    ast.KeyedNode
	listitem ast.Node
	num      int
	obj      ast.ObjectNode
	str      string
}

const NUMBER = 57346
const COMMA = 57347
const IDENTIFIER = 57348
const EQUAL = 57349
const NEWLINE = 57350
const STRING = 57351
const LEFTBRACE = 57352
const RIGHTBRACE = 57353
const LEFTBRACKET = 57354
const RIGHTBRACKET = 57355

var hclToknames = []string{
	"NUMBER",
	"COMMA",
	"IDENTIFIER",
	"EQUAL",
	"NEWLINE",
	"STRING",
	"LEFTBRACE",
	"RIGHTBRACE",
	"LEFTBRACKET",
	"RIGHTBRACKET",
}
var hclStatenames = []string{}

const hclEofCode = 1
const hclErrCode = 2
const hclMaxDepth = 200

//line parse.y:165

//line yacctab:1
var hclExca = []int{
	-1, 1,
	1, -1,
	-2, 0,
}

const hclNprod = 19
const hclPrivate = 57344

var hclTokenNames []string
var hclStates []string

const hclLast = 33

var hclAct = []int{

	21, 14, 24, 26, 2, 1, 15, 12, 8, 17,
	4, 25, 10, 7, 9, 19, 13, 18, 22, 7,
	12, 4, 16, 23, 7, 5, 6, 27, 3, 20,
	0, 0, 11,
}
var hclPact = []int{

	15, -1000, -1000, 15, 7, -1000, 10, -1000, -1000, -3,
	-1000, -1000, 4, -1000, -1000, -1000, -1000, 14, -9, -1000,
	-2, -1000, -1000, -1000, -1000, -1000, 14, -1000,
}
var hclPgo = []int{

	0, 29, 4, 28, 25, 0, 12, 26, 5,
}
var hclR1 = []int{

	0, 8, 2, 2, 6, 6, 3, 3, 3, 3,
	3, 4, 4, 7, 7, 1, 1, 5, 5,
}
var hclR2 = []int{

	0, 1, 1, 2, 3, 2, 3, 3, 3, 5,
	1, 2, 2, 1, 1, 1, 3, 1, 1,
}
var hclChk = []int{

	-1000, -8, -2, -3, 6, -4, -7, 9, -2, 7,
	-6, -4, 10, 6, 4, 9, -6, 12, -2, 11,
	-1, -5, 4, 9, 11, 13, 5, -5,
}
var hclDef = []int{

	0, -2, 1, 2, 13, 10, 0, 14, 3, 0,
	11, 12, 0, 13, 6, 7, 8, 0, 0, 5,
	0, 15, 17, 18, 4, 9, 0, 16,
}
var hclTok1 = []int{

	1,
}
var hclTok2 = []int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13,
}
var hclTok3 = []int{
	0,
}

//line yaccpar:1

/*	parser for yacc output	*/

var hclDebug = 0

type hclLexer interface {
	Lex(lval *hclSymType) int
	Error(s string)
}

const hclFlag = -1000

func hclTokname(c int) string {
	// 4 is TOKSTART above
	if c >= 4 && c-4 < len(hclToknames) {
		if hclToknames[c-4] != "" {
			return hclToknames[c-4]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func hclStatname(s int) string {
	if s >= 0 && s < len(hclStatenames) {
		if hclStatenames[s] != "" {
			return hclStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func hcllex1(lex hclLexer, lval *hclSymType) int {
	c := 0
	char := lex.Lex(lval)
	if char <= 0 {
		c = hclTok1[0]
		goto out
	}
	if char < len(hclTok1) {
		c = hclTok1[char]
		goto out
	}
	if char >= hclPrivate {
		if char < hclPrivate+len(hclTok2) {
			c = hclTok2[char-hclPrivate]
			goto out
		}
	}
	for i := 0; i < len(hclTok3); i += 2 {
		c = hclTok3[i+0]
		if c == char {
			c = hclTok3[i+1]
			goto out
		}
	}

out:
	if c == 0 {
		c = hclTok2[1] /* unknown char */
	}
	if hclDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", hclTokname(c), uint(char))
	}
	return c
}

func hclParse(hcllex hclLexer) int {
	var hcln int
	var hcllval hclSymType
	var hclVAL hclSymType
	hclS := make([]hclSymType, hclMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	hclstate := 0
	hclchar := -1
	hclp := -1
	goto hclstack

ret0:
	return 0

ret1:
	return 1

hclstack:
	/* put a state and value onto the stack */
	if hclDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", hclTokname(hclchar), hclStatname(hclstate))
	}

	hclp++
	if hclp >= len(hclS) {
		nyys := make([]hclSymType, len(hclS)*2)
		copy(nyys, hclS)
		hclS = nyys
	}
	hclS[hclp] = hclVAL
	hclS[hclp].yys = hclstate

hclnewstate:
	hcln = hclPact[hclstate]
	if hcln <= hclFlag {
		goto hcldefault /* simple state */
	}
	if hclchar < 0 {
		hclchar = hcllex1(hcllex, &hcllval)
	}
	hcln += hclchar
	if hcln < 0 || hcln >= hclLast {
		goto hcldefault
	}
	hcln = hclAct[hcln]
	if hclChk[hcln] == hclchar { /* valid shift */
		hclchar = -1
		hclVAL = hcllval
		hclstate = hcln
		if Errflag > 0 {
			Errflag--
		}
		goto hclstack
	}

hcldefault:
	/* default state action */
	hcln = hclDef[hclstate]
	if hcln == -2 {
		if hclchar < 0 {
			hclchar = hcllex1(hcllex, &hcllval)
		}

		/* look through exception table */
		xi := 0
		for {
			if hclExca[xi+0] == -1 && hclExca[xi+1] == hclstate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			hcln = hclExca[xi+0]
			if hcln < 0 || hcln == hclchar {
				break
			}
		}
		hcln = hclExca[xi+1]
		if hcln < 0 {
			goto ret0
		}
	}
	if hcln == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			hcllex.Error("syntax error")
			Nerrs++
			if hclDebug >= 1 {
				__yyfmt__.Printf("%s", hclStatname(hclstate))
				__yyfmt__.Printf(" saw %s\n", hclTokname(hclchar))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for hclp >= 0 {
				hcln = hclPact[hclS[hclp].yys] + hclErrCode
				if hcln >= 0 && hcln < hclLast {
					hclstate = hclAct[hcln] /* simulate a shift of "error" */
					if hclChk[hclstate] == hclErrCode {
						goto hclstack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if hclDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", hclS[hclp].yys)
				}
				hclp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if hclDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", hclTokname(hclchar))
			}
			if hclchar == hclEofCode {
				goto ret1
			}
			hclchar = -1
			goto hclnewstate /* try again in the same state */
		}
	}

	/* reduction by production hcln */
	if hclDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", hcln, hclStatname(hclstate))
	}

	hclnt := hcln
	hclpt := hclp
	_ = hclpt // guard against "declared and not used"

	hclp -= hclR2[hcln]
	hclVAL = hclS[hclp+1]

	/* consult goto table to find next state */
	hcln = hclR1[hcln]
	hclg := hclPgo[hcln]
	hclj := hclg + hclS[hclp].yys + 1

	if hclj >= hclLast {
		hclstate = hclAct[hclg]
	} else {
		hclstate = hclAct[hclj]
		if hclChk[hclstate] != -hcln {
			hclstate = hclAct[hclg]
		}
	}
	// dummy call; replaced with literal code
	switch hclnt {

	case 1:
		//line parse.y:37
		{
			hclResult = &ast.ObjectNode{
				K:    "",
				Elem: hclS[hclpt-0].klist,
			}
		}
	case 2:
		//line parse.y:46
		{
			hclVAL.klist = []ast.KeyedNode{hclS[hclpt-0].kitem}
		}
	case 3:
		//line parse.y:50
		{
			hclVAL.klist = append(hclS[hclpt-0].klist, hclS[hclpt-1].kitem)
		}
	case 4:
		//line parse.y:56
		{
			hclVAL.obj = ast.ObjectNode{Elem: hclS[hclpt-1].klist}
		}
	case 5:
		//line parse.y:60
		{
			hclVAL.obj = ast.ObjectNode{}
		}
	case 6:
		//line parse.y:66
		{
			hclVAL.kitem = ast.AssignmentNode{
				K: hclS[hclpt-2].str,
				Value: ast.LiteralNode{
					Type:  ast.ValueTypeInt,
					Value: hclS[hclpt-0].num,
				},
			}
		}
	case 7:
		//line parse.y:76
		{
			hclVAL.kitem = ast.AssignmentNode{
				K: hclS[hclpt-2].str,
				Value: ast.LiteralNode{
					Type:  ast.ValueTypeString,
					Value: hclS[hclpt-0].str,
				},
			}
		}
	case 8:
		//line parse.y:86
		{
			hclVAL.kitem = ast.AssignmentNode{
				K:     hclS[hclpt-2].str,
				Value: hclS[hclpt-0].obj,
			}
		}
	case 9:
		//line parse.y:93
		{
			hclVAL.kitem = ast.AssignmentNode{
				K:     hclS[hclpt-4].str,
				Value: ast.ListNode{Elem: hclS[hclpt-1].list},
			}
		}
	case 10:
		//line parse.y:100
		{
			hclVAL.kitem = hclS[hclpt-0].kitem
		}
	case 11:
		//line parse.y:106
		{
			hclVAL.kitem = ast.AssignmentNode{
				K: hclS[hclpt-1].str,
				Value: ast.ListNode{
					Elem: []ast.Node{hclS[hclpt-0].obj},
				},
			}
		}
	case 12:
		//line parse.y:115
		{
			obj := ast.ObjectNode{
				K:    hclS[hclpt-0].kitem.Key(),
				Elem: []ast.KeyedNode{hclS[hclpt-0].kitem},
			}

			hclVAL.kitem = ast.AssignmentNode{
				K: hclS[hclpt-1].str,
				Value: ast.ListNode{
					Elem: []ast.Node{obj},
				},
			}
		}
	case 13:
		//line parse.y:131
		{
			hclVAL.str = hclS[hclpt-0].str
		}
	case 14:
		//line parse.y:135
		{
			hclVAL.str = hclS[hclpt-0].str
		}
	case 15:
		//line parse.y:141
		{
			hclVAL.list = []ast.Node{hclS[hclpt-0].listitem}
		}
	case 16:
		//line parse.y:145
		{
			hclVAL.list = append(hclS[hclpt-2].list, hclS[hclpt-0].listitem)
		}
	case 17:
		//line parse.y:151
		{
			hclVAL.listitem = ast.LiteralNode{
				Type:  ast.ValueTypeInt,
				Value: hclS[hclpt-0].num,
			}
		}
	case 18:
		//line parse.y:158
		{
			hclVAL.listitem = ast.LiteralNode{
				Type:  ast.ValueTypeString,
				Value: hclS[hclpt-0].str,
			}
		}
	}
	goto hclstack /* stack new state and value */
}
