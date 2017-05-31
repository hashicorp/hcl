package zclsyntax

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/apparentlymart/go-textseg/textseg"
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
			s, sDiags := p.decodeQuotedLit(tok)
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

// decodeQuotedLit processes the given TokenQuotedLit token as if it were
// a string literal appearing in quotes, returning the string resulting from
// resolving any escape sequences.
//
// If any error diagnostics are returned, the returned string may be incomplete
// or otherwise invalid.
func (p *parser) decodeQuotedLit(tok Token) (string, zcl.Diagnostics) {
	if tok.Type != TokenQuotedLit {
		panic("decodeQuotedLit can only be used with TokenQuotedLit tokens")
	}
	var diags zcl.Diagnostics

	ret := make([]byte, 0, len(tok.Bytes))

	sc := bufio.NewScanner(bytes.NewReader(tok.Bytes))
	sc.Split(textseg.ScanGraphemeClusters)

	escaping := rune(0)
	pos := tok.Range.Start
	for sc.Scan() {
		switch escaping {
		case '\\':
			escaping = 0
			ty := sc.Text()
			switch ty {
			case "n":
				ret = append(ret, 10)
			case "r":
				ret = append(ret, 13)
			case "t":
				ret = append(ret, 9)

			// TODO: numeric character escapes with \uXXXX

			default:
				diags = append(diags, &zcl.Diagnostic{
					Severity: zcl.DiagError,
					Summary:  "Invalid escape sequence",
					Detail:   fmt.Sprintf("The sequence \"\\%s\" is not a recognized escape sequence.", ty),
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
							Byte:   pos.Byte + len(ty),
						},
					},
				})
				ret = append(ret, sc.Bytes()...)
			}
		case '$', '!':
			bytes := sc.Bytes()
			if len(bytes) != 1 || bytes[0] == byte(escaping) {
				ret = append(ret, byte(escaping))
			}
			ret = append(ret, bytes...)
		default:
			switch sc.Text() {
			case "\\":
				escaping = '\\'
			case "$":
				escaping = '$'
			case "!":
				escaping = '!'
			default:
				ret = append(ret, sc.Bytes()...)
			}
		}

		// Literal newlines cannot appear in quoted literals, so it's safe
		// to just increment Column and Byte in our position.
		pos.Column++
		pos.Byte += len(sc.Bytes())
	}

	return string(ret), diags
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
