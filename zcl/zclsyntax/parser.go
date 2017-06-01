package zclsyntax

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/apparentlymart/go-textseg/textseg"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-zcl/zcl"
)

type parser struct {
	*peeker

	// set to true if any recovery is attempted. The parser can use this
	// to attempt to reduce error noise by suppressing "bad token" errors
	// in recovery mode, assuming that the recovery heuristics have failed
	// in this case and left the peeker in a wrong place.
	recovery bool
}

func (p *parser) ParseBody(end TokenType) (*Body, zcl.Diagnostics) {
	attrs := Attributes{}
	blocks := Blocks{}
	var diags zcl.Diagnostics

	startRange := p.PrevRange()
	var endRange zcl.Range

Token:
	for {
		next := p.Peek()
		if next.Type == end {
			endRange = p.NextRange()
			p.Read()
			break Token
		}

		switch next.Type {
		case TokenNewline:
			p.Read()
			continue
		case TokenIdent:
			item, itemDiags := p.ParseBodyItem()
			diags = append(diags, itemDiags...)
			switch titem := item.(type) {
			case *Block:
				blocks = append(blocks, titem)
			case *Attribute:
				if existing, exists := attrs[titem.Name]; exists {
					diags = append(diags, &zcl.Diagnostic{
						Severity: zcl.DiagError,
						Summary:  "Attribute redefined",
						Detail: fmt.Sprintf(
							"The attribute %q was already defined at %s. Each attribute may be defined only once.",
							titem.Name, existing.NameRange.String(),
						),
						Subject: &titem.NameRange,
					})
				} else {
					attrs[titem.Name] = titem
				}
			default:
				// This should never happen for valid input, but may if a
				// syntax error was detected in ParseBodyItem that prevented
				// it from even producing a partially-broken item. In that
				// case, it would've left at least one error in the diagnostics
				// slice we already dealt with above.
				//
				// We'll assume ParseBodyItem attempted recovery to leave
				// us in a reasonable position to try parsing the next item.
				continue
			}
		default:
			bad := p.Read()
			if !p.recovery {
				if bad.Type == TokenOQuote {
					diags = append(diags, &zcl.Diagnostic{
						Severity: zcl.DiagError,
						Summary:  "Invalid attribute name",
						Detail:   "Attribute names must not be quoted.",
						Subject:  &bad.Range,
					})
				} else {
					diags = append(diags, &zcl.Diagnostic{
						Severity: zcl.DiagError,
						Summary:  "Attribute or block definition required",
						Detail:   "An attribute or block definition is required here.",
						Subject:  &bad.Range,
					})
				}
			}
			endRange = p.PrevRange() // arbitrary, but somewhere inside the body means better diagnostics

			p.recover(end) // attempt to recover to the token after the end of this body
			break Token
		}
	}

	return &Body{
		Attributes: attrs,
		Blocks:     blocks,

		SrcRange: zcl.RangeBetween(startRange, endRange),
		EndRange: zcl.Range{
			Filename: endRange.Filename,
			Start:    endRange.End,
			End:      endRange.End,
		},
	}, diags
}

func (p *parser) ParseBodyItem() (Node, zcl.Diagnostics) {
	ident := p.Read()
	if ident.Type != TokenIdent {
		p.recoverAfterBodyItem()
		return nil, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Attribute or block definition required",
				Detail:   "An attribute or block definition is required here.",
				Subject:  &ident.Range,
			},
		}
	}

	next := p.Peek()

	switch next.Type {
	case TokenEqual:
		return p.finishParsingBodyAttribute(ident)
	case TokenOQuote, TokenOBrace:
		return p.finishParsingBodyBlock(ident)
	default:
		p.recoverAfterBodyItem()
		return nil, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Attribute or block definition required",
				Detail:   "An attribute or block definition is required here. To define an attribute, use the equals sign \"=\" to introduce the attribute value.",
				Subject:  &ident.Range,
			},
		}
	}

	return nil, nil
}

func (p *parser) finishParsingBodyAttribute(ident Token) (Node, zcl.Diagnostics) {
	panic("attribute parsing not yet implemented")
}

func (p *parser) finishParsingBodyBlock(ident Token) (Node, zcl.Diagnostics) {
	var blockType = string(ident.Bytes)
	var diags zcl.Diagnostics
	var labels []string
	var labelRanges []zcl.Range

	var oBrace Token

Token:
	for {
		tok := p.Peek()

		switch tok.Type {

		case TokenOBrace:
			oBrace = p.Read()
			break Token

		case TokenOQuote:
			label, labelRange, labelDiags := p.parseQuotedStringLiteral()
			diags = append(diags, labelDiags...)
			labels = append(labels, label)
			labelRanges = append(labelRanges, labelRange)
			if labelDiags.HasErrors() {
				p.recoverAfterBodyItem()
				return &Block{
					Type:   blockType,
					Labels: labels,
					Body:   nil,

					TypeRange:       ident.Range,
					LabelRanges:     labelRanges,
					OpenBraceRange:  ident.Range, // placeholder
					CloseBraceRange: ident.Range, // placeholder
				}, diags
			}

		default:
			switch tok.Type {
			case TokenEqual:
				diags = append(diags, &zcl.Diagnostic{
					Severity: zcl.DiagError,
					Summary:  "Invalid block definition",
					Detail:   "The equals sign \"=\" indicates an attribute definition, and must not be used when defining a block.",
					Subject:  &tok.Range,
					Context:  zcl.RangeBetween(ident.Range, tok.Range).Ptr(),
				})
			case TokenNewline:
				diags = append(diags, &zcl.Diagnostic{
					Severity: zcl.DiagError,
					Summary:  "Invalid block definition",
					Detail:   "A block definition must have block content delimited by \"{\" and \"}\", starting on the same line as the block header.",
					Subject:  &tok.Range,
					Context:  zcl.RangeBetween(ident.Range, tok.Range).Ptr(),
				})
			default:
				if !p.recovery {
					diags = append(diags, &zcl.Diagnostic{
						Severity: zcl.DiagError,
						Summary:  "Invalid block definition",
						Detail:   "Either a quoted string block label or an opening brace (\"{\") is expected here.",
						Subject:  &tok.Range,
						Context:  zcl.RangeBetween(ident.Range, tok.Range).Ptr(),
					})
				}
			}

			p.recoverAfterBodyItem()

			return &Block{
				Type:   blockType,
				Labels: labels,
				Body:   nil,

				TypeRange:       ident.Range,
				LabelRanges:     labelRanges,
				OpenBraceRange:  ident.Range, // placeholder
				CloseBraceRange: ident.Range, // placeholder
			}, diags
		}
	}

	// Once we fall out here, the peeker is pointed just after our opening
	// brace, so we can begin our nested body parsing.
	body, bodyDiags := p.ParseBody(TokenCBrace)
	diags = append(diags, bodyDiags...)
	cBraceRange := p.PrevRange()

	return &Block{
		Type:   blockType,
		Labels: labels,
		Body:   body,

		TypeRange:       ident.Range,
		LabelRanges:     labelRanges,
		OpenBraceRange:  oBrace.Range,
		CloseBraceRange: cBraceRange,
	}, diags
}

func (p *parser) ParseExpression() (Expression, zcl.Diagnostics) {
	return p.parseTernaryConditional()
}

func (p *parser) parseTernaryConditional() (Expression, zcl.Diagnostics) {
	// The ternary conditional operator (.. ? .. : ..) behaves somewhat
	// like a binary operator except that the "symbol" is itself
	// an expression enclosed in two punctuation characters.
	// The middle expression is parsed as if the ? and : symbols
	// were parentheses. The "rhs" (the "false expression") is then
	// treated right-associatively so it behaves similarly to the
	// middle in terms of precedence.

	startRange := p.NextRange()
	var condExpr, trueExpr, falseExpr Expression
	var diags zcl.Diagnostics

	condExpr, condDiags := p.parseBinaryOps(binaryOps)
	diags = append(diags, condDiags...)
	if p.recovery && condDiags.HasErrors() {
		return condExpr, diags
	}

	questionMark := p.Peek()
	if questionMark.Type != TokenQuestion {
		return condExpr, diags
	}

	p.Read() // eat question mark

	trueExpr, trueDiags := p.ParseExpression()
	diags = append(diags, trueDiags...)
	if p.recovery && trueDiags.HasErrors() {
		return condExpr, diags
	}

	colon := p.Peek()
	if colon.Type != TokenColon {
		diags = append(diags, &zcl.Diagnostic{
			Severity: zcl.DiagError,
			Summary:  "Missing false expression in conditional",
			Detail:   "The conditional operator (...?...:...) requires a false expression, delimited by a colon.",
			Subject:  &colon.Range,
			Context:  zcl.RangeBetween(startRange, colon.Range).Ptr(),
		})
		return condExpr, diags
	}

	p.Read() // eat colon

	falseExpr, falseDiags := p.ParseExpression()
	diags = append(diags, falseDiags...)
	if p.recovery && falseDiags.HasErrors() {
		return condExpr, diags
	}

	return &ConditionalExpr{
		Condition:   condExpr,
		TrueResult:  trueExpr,
		FalseResult: falseExpr,

		SrcRange: zcl.RangeBetween(startRange, falseExpr.Range()),
	}, diags
}

// parseBinaryOps calls itself recursively to work through all of the
// operator precedence groups, and then eventually calls parseExpressionTerm
// for each operand.
func (p *parser) parseBinaryOps(ops []map[TokenType]Operation) (Expression, zcl.Diagnostics) {
	if len(ops) == 0 {
		// We've run out of operators, so now we'll just try to parse a term.
		return p.parseExpressionTerm()
	}

	thisLevel := ops[0]
	remaining := ops[1:]

	var lhs, rhs Expression
	operation := OpNil
	var diags zcl.Diagnostics

	// Parse a term that might be the first operand of a binary
	// operation or it might just be a standalone term.
	// We won't know until we've parsed it and can look ahead
	// to see if there's an operator token for this level.
	lhs, lhsDiags := p.parseBinaryOps(remaining)
	diags = append(diags, lhsDiags...)
	if p.recovery && lhsDiags.HasErrors() {
		return lhs, diags
	}

	// We'll keep eating up operators until we run out, so that operators
	// with the same precedence will combine in a left-associative manner:
	// a+b+c => (a+b)+c, not a+(b+c)
	//
	// Should we later want to have right-associative operators, a way
	// to achieve that would be to call back up to ParseExpression here
	// instead of iteratively parsing only the remaining operators.
	for {
		next := p.Peek()
		var newOp Operation
		var ok bool
		if newOp, ok = thisLevel[next.Type]; !ok {
			break
		}

		// Are we extending an expression started on the previous iteration?
		if operation != OpNil {
			lhs = &BinaryOpExpr{
				LHS: lhs,
				Op:  operation,
				RHS: rhs,

				SrcRange: zcl.RangeBetween(lhs.Range(), rhs.Range()),
			}
		}

		operation = newOp
		p.Read() // eat operator token
		var rhsDiags zcl.Diagnostics
		rhs, rhsDiags = p.parseBinaryOps(remaining)
		diags = append(diags, rhsDiags...)
		if p.recovery && rhsDiags.HasErrors() {
			return lhs, diags
		}
	}

	if operation == OpNil {
		return lhs, diags
	}

	return &BinaryOpExpr{
		LHS: lhs,
		Op:  operation,
		RHS: rhs,

		SrcRange: zcl.RangeBetween(lhs.Range(), rhs.Range()),
	}, diags
}

func (p *parser) parseExpressionTerm() (Expression, zcl.Diagnostics) {
	start := p.Peek()

	switch start.Type {
	case TokenOParen:
		p.Read() // eat open paren
		expr, diags := p.ParseExpression()
		if diags.HasErrors() {
			// attempt to place the peeker after our closing paren
			// before we return, so that the next parser has some
			// chance of finding a valid expression.
			p.recover(TokenCParen)
			return expr, diags
		}

		close := p.Peek()
		if close.Type != TokenCParen {
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Unbalanced parentheses",
				Detail:   "Expected a closing parenthesis to terminate the expression.",
				Subject:  &close.Range,
				Context:  zcl.RangeBetween(start.Range, close.Range).Ptr(),
			})
			p.setRecovery()
		}

		p.Read() // eat closing paren

		return expr, diags

	case TokenNumberLit:
		tok := p.Read() // eat number token

		// We'll lean on the cty converter to do the conversion, to ensure that
		// the behavior is the same as what would happen if converting a
		// non-literal string to a number.
		numStrVal := cty.StringVal(string(tok.Bytes))
		numVal, err := convert.Convert(numStrVal, cty.Number)
		if err != nil {
			ret := &LiteralValueExpr{
				Val:      cty.UnknownVal(cty.Number),
				SrcRange: tok.Range,
			}
			return ret, zcl.Diagnostics{
				{
					Severity: zcl.DiagError,
					Summary:  "Invalid number literal",
					// FIXME: not a very good error message, but convert only
					// gives us "a number is required", so not much help either.
					Detail:  "Failed to recognize the value of this number literal.",
					Subject: &ret.SrcRange,
				},
			}
		}

		return &LiteralValueExpr{
			Val:      numVal,
			SrcRange: tok.Range,
		}, nil

	case TokenIdent:
		tok := p.Read() // eat identifier token

		name := string(tok.Bytes)
		switch name {
		case "true":
			return &LiteralValueExpr{
				Val:      cty.True,
				SrcRange: tok.Range,
			}, nil
		case "false":
			return &LiteralValueExpr{
				Val:      cty.False,
				SrcRange: tok.Range,
			}, nil
		case "null":
			return &LiteralValueExpr{
				Val:      cty.NullVal(cty.DynamicPseudoType),
				SrcRange: tok.Range,
			}, nil
		default:
			return &ScopeTraversalExpr{
				Traversal: zcl.Traversal{
					zcl.TraverseRoot{
						Name:     name,
						SrcRange: tok.Range,
					},
				},
				SrcRange: tok.Range,
			}, nil
		}

	case TokenOQuote, TokenOHeredoc:
		open := p.Read() // eat opening marker
		closer := p.oppositeBracket(open.Type)
		return p.ParseTemplate(closer)

	case TokenMinus:
		tok := p.Read() // eat minus token

		// Important to use parseExpressionTerm rather than parseExpression
		// here, otherwise we can capture a following binary expression into
		// our negation.
		// e.g. -46+5 should parse as (-46)+5, not -(46+5)
		operand, diags := p.parseExpressionTerm()
		return &UnaryOpExpr{
			Op:  OpNegate,
			Val: operand,

			SrcRange:    zcl.RangeBetween(tok.Range, operand.Range()),
			SymbolRange: tok.Range,
		}, diags

	case TokenBang:
		tok := p.Read() // eat bang token

		// Important to use parseExpressionTerm rather than parseExpression
		// here, otherwise we can capture a following binary expression into
		// our negation.
		operand, diags := p.parseExpressionTerm()
		return &UnaryOpExpr{
			Op:  OpLogicalNot,
			Val: operand,

			SrcRange:    zcl.RangeBetween(tok.Range, operand.Range()),
			SymbolRange: tok.Range,
		}, diags

	default:
		var diags zcl.Diagnostics
		if !p.recovery {
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Invalid expression",
				Detail:   "Expected the start of an expression, but found an invalid expression token.",
				Subject:  &start.Range,
			})
		}
		p.setRecovery()

		// Return a placeholder so that the AST is still structurally sound
		// even in the presence of parse errors.
		return &LiteralValueExpr{
			Val:      cty.DynamicVal,
			SrcRange: start.Range,
		}, diags
	}
}

func (p *parser) ParseTemplate(end TokenType) (Expression, zcl.Diagnostics) {
	var parts []Expression
	var diags zcl.Diagnostics

	startRange := p.NextRange()

Token:
	for {
		next := p.Read()
		if next.Type == end {
			// all done!
			break
		}

		switch next.Type {
		case TokenStringLit, TokenQuotedLit:
			str, strDiags := p.decodeStringLit(next)
			diags = append(diags, strDiags...)
			parts = append(parts, &LiteralValueExpr{
				Val:      cty.StringVal(str),
				SrcRange: next.Range,
			})
		case TokenTemplateInterp:
			// TODO: if opener has ~ mark, eat trailing spaces in the previous
			// literal.
			expr, exprDiags := p.ParseExpression()
			diags = append(diags, exprDiags...)
			close := p.Peek()
			if close.Type != TokenTemplateSeqEnd {
				p.recover(TokenTemplateSeqEnd)
			} else {
				p.Read() // eat closing brace
				// TODO: if closer has ~ mark, remember to eat leading spaces
				// in the following literal.
			}
			parts = append(parts, expr)
		case TokenTemplateControl:
			panic("template control sequences not yet supported")

		default:
			if !p.recovery {
				diags = append(diags, &zcl.Diagnostic{
					Severity: zcl.DiagError,
					Summary:  "Unterminated template string",
					Detail:   "No closing marker was found for the string.",
					Subject:  &next.Range,
					Context:  zcl.RangeBetween(startRange, next.Range).Ptr(),
				})
			}
			p.recover(end)
			break Token
		}
	}

	if len(parts) == 0 {
		// If a sequence has no content, we'll treat it as if it had an
		// empty string in it because that's what the user probably means
		// if they write "" in configuration.
		return &LiteralValueExpr{
			Val: cty.StringVal(""),
			SrcRange: zcl.Range{
				Filename: startRange.Filename,
				Start:    startRange.Start,
				End:      startRange.Start,
			},
		}, diags
	}

	if len(parts) == 1 {
		// If a sequence only has one part then as a special case we return
		// that part alone. This allows the use of single-part templates to
		// represent general expressions in syntaxes such as JSON where
		// un-quoted expressions are not possible.
		return parts[0], diags
	}

	return &TemplateExpr{
		Parts: parts,

		SrcRange: zcl.RangeBetween(parts[0].Range(), parts[len(parts)-1].Range()),
	}, diags
}

// parseQuotedStringLiteral is a helper for parsing quoted strings that
// aren't allowed to contain any interpolations, such as block labels.
func (p *parser) parseQuotedStringLiteral() (string, zcl.Range, zcl.Diagnostics) {
	oQuote := p.Read()
	if oQuote.Type != TokenOQuote {
		return "", oQuote.Range, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Invalid string literal",
				Detail:   "A quoted string is required here.",
				Subject:  &oQuote.Range,
			},
		}
	}

	var diags zcl.Diagnostics
	ret := &bytes.Buffer{}
	var cQuote Token

Token:
	for {
		tok := p.Read()
		switch tok.Type {

		case TokenCQuote:
			cQuote = tok
			break Token

		case TokenQuotedLit:
			s, sDiags := p.decodeStringLit(tok)
			diags = append(diags, sDiags...)
			ret.WriteString(s)

		case TokenTemplateControl, TokenTemplateInterp:
			which := "$"
			if tok.Type == TokenTemplateControl {
				which = "!"
			}

			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Invalid string literal",
				Detail: fmt.Sprintf(
					"Template sequences are not allowed in this string. To include a literal %q, double it (as \"%s%s\") to escape it.",
					which, which, which,
				),
				Subject: &tok.Range,
				Context: zcl.RangeBetween(oQuote.Range, tok.Range).Ptr(),
			})
			p.recover(TokenTemplateSeqEnd)

		case TokenEOF:
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Unterminated string literal",
				Detail:   "Unable to find the closing quote mark before the end of the file.",
				Subject:  &tok.Range,
				Context:  zcl.RangeBetween(oQuote.Range, tok.Range).Ptr(),
			})
			break Token

		default:
			// Should never happen, as long as the scanner is behaving itself
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Invalid string literal",
				Detail:   "This item is not valid in a string literal.",
				Subject:  &tok.Range,
				Context:  zcl.RangeBetween(oQuote.Range, tok.Range).Ptr(),
			})
			p.recover(TokenOQuote)
			break Token

		}

	}

	return ret.String(), zcl.RangeBetween(oQuote.Range, cQuote.Range), diags
}

// decodeStringLit processes the given token, which must be either a
// TokenQuotedLit or a TokenStringLit, returning the string resulting from
// resolving any escape sequences.
//
// If any error diagnostics are returned, the returned string may be incomplete
// or otherwise invalid.
func (p *parser) decodeStringLit(tok Token) (string, zcl.Diagnostics) {
	var quoted bool
	switch tok.Type {
	case TokenQuotedLit:
		quoted = true
	case TokenStringLit:
		quoted = false
	default:
		panic("decodeQuotedLit can only be used with TokenStringLit and TokenQuotedLit tokens")
	}
	var diags zcl.Diagnostics

	ret := make([]byte, 0, len(tok.Bytes))
	var esc []byte

	sc := bufio.NewScanner(bytes.NewReader(tok.Bytes))
	sc.Split(textseg.ScanGraphemeClusters)

	pos := tok.Range.Start
	newPos := pos
Character:
	for sc.Scan() {
		pos = newPos
		ch := sc.Bytes()

		// Adjust position based on our new character.
		// \r\n is considered to be a single character in text segmentation,
		if (len(ch) == 1 && ch[0] == '\n') || (len(ch) == 2 && ch[1] == '\n') {
			newPos.Line++
			newPos.Column = 0
		} else {
			newPos.Column++
		}
		newPos.Byte += len(ch)

		if len(esc) > 0 {
			switch esc[0] {
			case '\\':
				if len(ch) == 1 {
					switch ch[0] {

					// TODO: numeric character escapes with \uXXXX

					case 'n':
						ret = append(ret, '\n')
						esc = esc[:0]
						continue Character
					case 'r':
						ret = append(ret, '\r')
						esc = esc[:0]
						continue Character
					case 't':
						ret = append(ret, '\t')
						esc = esc[:0]
						continue Character
					case '"':
						ret = append(ret, '"')
						esc = esc[:0]
						continue Character
					case '\\':
						ret = append(ret, '\\')
						esc = esc[:0]
						continue Character
					}
				}

				var detail string
				switch {
				case len(ch) == 1 && (ch[0] == '$' || ch[0] == '!'):
					detail = fmt.Sprintf(
						"The characters \"\\%s\" do not form a recognized escape sequence. To escape a \"%s{\" template sequence, use \"%s%s{\".",
						ch, ch, ch, ch,
					)
				default:
					detail = fmt.Sprintf("The characters \"\\%s\" do not form a recognized escape sequence.", ch)
				}

				diags = append(diags, &zcl.Diagnostic{
					Severity: zcl.DiagError,
					Summary:  "Invalid escape sequence",
					Detail:   detail,
					Subject: &zcl.Range{
						Filename: tok.Range.Filename,
						Start: zcl.Pos{
							Line:   pos.Line,
							Column: pos.Column - 1, // safe because we know the previous character must be a backslash
							Byte:   pos.Byte - 1,
						},
						End: zcl.Pos{
							Line:   pos.Line,
							Column: pos.Column + 1, // safe because we know the previous character must be a backslash
							Byte:   pos.Byte + len(ch),
						},
					},
				})
				ret = append(ret, ch...)
				esc = esc[:0]
				continue Character

			case '$', '!':
				switch len(esc) {
				case 1:
					if len(ch) == 1 && ch[0] == esc[0] {
						esc = append(esc, ch[0])
						continue Character
					}

					// Any other character means this wasn't an escape sequence
					// after all.
					ret = append(ret, esc...)
					ret = append(ret, ch...)
					esc = esc[:0]
				case 2:
					if len(ch) == 1 && ch[0] == '{' {
						// successful escape sequence
						ret = append(ret, esc[0])
					} else {
						// not an escape sequence, so just output literal
						ret = append(ret, esc...)
					}
					ret = append(ret, ch...)
					esc = esc[:0]
				default:
					// should never happen
					panic("have invalid escape sequence >2 characters")
				}

			}
		} else {
			if len(ch) == 1 {
				switch ch[0] {
				case '\\':
					if quoted { // ignore backslashes in unquoted mode
						esc = append(esc, '\\')
						continue Character
					}
				case '$':
					esc = append(esc, '$')
					continue Character
				case '!':
					esc = append(esc, '!')
					continue Character
				}
			}
			ret = append(ret, ch...)
		}
	}

	return string(ret), diags
}

// setRecovery turns on recovery mode without actually doing any recovery.
// This can be used when a parser knowingly leaves the peeker in a useless
// place and wants to suppress errors that might result from that decision.
func (p *parser) setRecovery() {
	p.recovery = true
}

// recover seeks forward in the token stream until it finds TokenType "end",
// then returns with the peeker pointed at the following token.
//
// If the given token type is a bracketer, this function will additionally
// count nested instances of the brackets to try to leave the peeker at
// the end of the _current_ instance of that bracketer, skipping over any
// nested instances. This is a best-effort operation and may have
// unpredictable results on input with bad bracketer nesting.
func (p *parser) recover(end TokenType) {
	start := p.oppositeBracket(end)
	p.recovery = true

	nest := 0
	for {
		tok := p.Read()
		ty := tok.Type
		if end == TokenTemplateSeqEnd && ty == TokenTemplateControl {
			// normalize so that our matching behavior can work, since
			// TokenTemplateControl/TokenTemplateInterp are asymmetrical
			// with TokenTemplateSeqEnd and thus we need to count both
			// openers if that's the closer we're looking for.
			ty = TokenTemplateInterp
		}

		switch ty {
		case start:
			nest++
		case end:
			if nest < 1 {
				return
			}

			nest--
		case TokenEOF:
			return
		}
	}
}

// recoverOver seeks forward in the token stream until it finds a block
// starting with TokenType "start", then finds the corresponding end token,
// leaving the peeker pointed at the token after that end token.
//
// The given token type _must_ be a bracketer. For example, if the given
// start token is TokenOBrace then the parser will be left at the _end_ of
// the next brace-delimited block encountered, or at EOF if no such block
// is found or it is unclosed.
func (p *parser) recoverOver(start TokenType) {
	end := p.oppositeBracket(start)

	// find the opening bracket first
Token:
	for {
		tok := p.Read()
		switch tok.Type {
		case start, TokenEOF:
			break Token
		}
	}

	// Now use our existing recover function to locate the _end_ of the
	// container we've found.
	p.recover(end)
}

func (p *parser) recoverAfterBodyItem() {
	p.recovery = true
	var open []TokenType

Token:
	for {
		tok := p.Read()

		switch tok.Type {

		case TokenNewline:
			if len(open) == 0 {
				break Token
			}

		case TokenEOF:
			break Token

		case TokenOBrace, TokenOBrack, TokenOParen, TokenOQuote, TokenOHeredoc, TokenTemplateInterp, TokenTemplateControl:
			open = append(open, tok.Type)

		case TokenCBrace, TokenCBrack, TokenCParen, TokenCQuote, TokenCHeredoc:
			opener := p.oppositeBracket(tok.Type)
			for len(open) > 0 && open[len(open)-1] != opener {
				open = open[:len(open)-1]
			}
			if len(open) > 0 {
				open = open[:len(open)-1]
			}

		case TokenTemplateSeqEnd:
			for len(open) > 0 && open[len(open)-1] != TokenTemplateInterp && open[len(open)-1] != TokenTemplateControl {
				open = open[:len(open)-1]
			}
			if len(open) > 0 {
				open = open[:len(open)-1]
			}

		}
	}
}

// oppositeBracket finds the bracket that opposes the given bracketer, or
// NilToken if the given token isn't a bracketer.
//
// "Bracketer", for the sake of this function, is one end of a matching
// open/close set of tokens that establish a bracketing context.
func (p *parser) oppositeBracket(ty TokenType) TokenType {
	switch ty {

	case TokenOBrace:
		return TokenCBrace
	case TokenOBrack:
		return TokenCBrack
	case TokenOParen:
		return TokenCParen
	case TokenOQuote:
		return TokenCQuote
	case TokenOHeredoc:
		return TokenCHeredoc

	case TokenCBrace:
		return TokenOBrace
	case TokenCBrack:
		return TokenOBrack
	case TokenCParen:
		return TokenOParen
	case TokenCQuote:
		return TokenOQuote
	case TokenCHeredoc:
		return TokenOHeredoc

	case TokenTemplateControl:
		return TokenTemplateSeqEnd
	case TokenTemplateInterp:
		return TokenTemplateSeqEnd
	case TokenTemplateSeqEnd:
		// This is ambigous, but we return Interp here because that's
		// what's assumed by the "recover" method.
		return TokenTemplateInterp

	default:
		return TokenNil
	}
}
