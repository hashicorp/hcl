package zclsyntax

import (
	"fmt"

	"github.com/zclconf/go-zcl/zcl"
)

type parser struct {
	*peeker
}

func (p *parser) ParseBody(end TokenType) (*Body, zcl.Diagnostics) {
	attrs := Attributes{}
	blocks := Blocks{}
	var diags zcl.Diagnostics

	startRange := p.NextRange()
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
		case TokenIdent, TokenOQuote:
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
			diags = append(diags, &zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Attribute or block definition required",
				Detail:   "An attribute or block definition is required here.",
				Subject:  &bad.Range,
			})
			endRange = p.NextRange() // arbitrary, but somewhere inside the body means better diagnostics

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
	return nil, nil
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

	nest := 0
	for {
		tok := p.Read()
		switch tok.Type {
		case start:
			nest++
		case end:
			if nest < 1 {
				return
			}

			nest--
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

	default:
		return TokenNil
	}
}
