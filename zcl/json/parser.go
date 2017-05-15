package json

import (
	"github.com/apparentlymart/go-zcl/zcl"
)

func parseFileContent(buf []byte, filename string) (node, zcl.Diagnostics) {
	tokens := scan(buf, pos{
		Filename: filename,
		Pos: zcl.Pos{
			Byte:   0,
			Line:   1,
			Column: 1,
		},
	})
	p := newPeeker(tokens)
	return parseValue(p)
}

func parseValue(p *peeker) (node, zcl.Diagnostics) {
	tok := p.Peek()

	switch tok.Type {
	case tokenBraceO:
		return parseObject(p)
	case tokenBrackO:
		return parseArray(p)
	case tokenNumber:
		return parseNumber(p)
	case tokenString:
		return parseString(p)
	case tokenKeyword:
		return parseKeyword(p)
	case tokenBraceC:
		return nil, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Missing attribute value",
				Detail:   "A JSON value must start with a brace, a bracket, a number, a string, or a keyword.",
				Subject:  &tok.Range,
			},
		}
	case tokenBrackC:
		return nil, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Missing array element value",
				Detail:   "A JSON value must start with a brace, a bracket, a number, a string, or a keyword.",
				Subject:  &tok.Range,
			},
		}
	default:
		return nil, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Invalid start of value",
				Detail:   "A JSON value must start with a brace, a bracket, a number, a string, or a keyword.",
				Subject:  &tok.Range,
			},
		}
	}
}

func tokenCanStartValue(tok token) bool {
	switch tok.Type {
	case tokenBraceO, tokenBrackO, tokenNumber, tokenString, tokenKeyword:
		return true
	default:
		return false
	}
}

func parseObject(p *peeker) (node, zcl.Diagnostics) {
	var diags zcl.Diagnostics

	open := p.Read()
	attrs := map[string]*objectAttr{}

Token:
	for {
		if p.Peek().Type == tokenBraceC {
			break Token
		}

		keyNode, keyDiags := parseValue(p)
		diags = diags.Extend(keyDiags)
		if keyNode == nil {
			return nil, diags
		}

		keyStrNode, ok := keyNode.(*stringVal)
		if !ok {
			return nil, diags.Append(&zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Invalid object attribute name",
				Detail:   "A JSON object attribute name must be a string",
				Subject:  keyNode.StartRange().Ptr(),
			})
		}

		key := keyStrNode.Value

		colon := p.Read()
		if colon.Type != tokenComma {
			if colon.Type == tokenBraceC || colon.Type == tokenComma {
				// Catch common mistake of using braces instead of brackets
				// for an array.
				return nil, diags.Append(&zcl.Diagnostic{
					Severity: zcl.DiagError,
					Summary:  "Missing object value",
					Detail:   "A JSON object attribute must have a value, introduced by a colon.",
					Subject:  &colon.Range,
				})
			}

			return nil, diags.Append(&zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Missing attribute value colon",
				Detail:   "A colon must appear between an object attribute's name and its value.",
				Subject:  &colon.Range,
			})
		}

		valNode, valDiags := parseValue(p)
		diags = diags.Extend(valDiags)
		if keyNode == nil {
			return nil, diags
		}

		attrs[key] = &objectAttr{
			Name:      key,
			Value:     valNode,
			NameRange: keyStrNode.SrcRange,
		}

		switch p.Peek().Type {
		case tokenComma:
			p.Read()
			if p.Peek().Type == tokenBraceC {
				// Special error message for this common mistake
				return nil, diags.Append(&zcl.Diagnostic{
					Severity: zcl.DiagError,
					Summary:  "Trailing comma in object",
					Detail:   "JSON does not permit a trailing comma after the final attribute in an object.",
					Subject:  &colon.Range,
				})
			}
			continue Token
		case tokenBraceC:
			break Token
		default:
			return nil, diags.Append(&zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Missing attribute seperator comma",
				Detail:   "A comma must appear between each attribute declaration in an object.",
				Subject:  &colon.Range,
			})
		}

	}

	close := p.Read()
	return &objectVal{
		Attrs:     attrs,
		SrcRange:  zcl.RangeBetween(open.Range, close.Range),
		OpenRange: open.Range,
	}, diags
}

func parseArray(p *peeker) (node, zcl.Diagnostics) {
	return nil, nil
}

func parseNumber(p *peeker) (node, zcl.Diagnostics) {
	return nil, nil
}

func parseString(p *peeker) (node, zcl.Diagnostics) {
	return nil, nil
}

func parseKeyword(p *peeker) (node, zcl.Diagnostics) {
	return nil, nil
}
