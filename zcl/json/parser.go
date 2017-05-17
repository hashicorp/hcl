package json

import (
	"encoding/json"
	"fmt"
	"math/big"

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
	node, diags := parseValue(p)
	if diags.HasErrors() {
		// Don't return a node if there were errors during parsing.
		return nil, diags
	}
	return node, diags
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
		if colon.Type != tokenColon {
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

		if existing := attrs[key]; existing != nil {
			// Generate a diagnostic for the duplicate key, but continue parsing
			// anyway since this is a semantic error we can recover from.
			diags = diags.Append(&zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Duplicate object attribute",
				Detail: fmt.Sprintf(
					"An attribute named %q was previously introduced at %s",
					key, existing.NameRange.String(),
				),
				Subject: &colon.Range,
			})
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
		case tokenEOF:
			return nil, diags.Append(&zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Unclosed object",
				Detail:   "No closing brace was found for this JSON object.",
				Subject:  &open.Range,
			})
		case tokenBrackC:
			return nil, diags.Append(&zcl.Diagnostic{
				Severity: zcl.DiagError,
				Summary:  "Mismatched braces",
				Detail:   "A JSON object must be closed with a brace, not a bracket.",
				Subject:  p.Peek().Range.Ptr(),
			})
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
	tok := p.Read()

	// Use encoding/json to validate the number syntax.
	// TODO: Do this more directly to produce better diagnostics.
	var num json.Number
	err := json.Unmarshal(tok.Bytes, &num)
	if err != nil {
		return nil, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Invalid JSON number",
				Detail:   fmt.Sprintf("There is a syntax error in the given JSON number."),
				Subject:  &tok.Range,
			},
		}
	}

	f, _, err := (&big.Float{}).Parse(string(num), 10)
	if err != nil {
		// Should never happen if above passed, since JSON numbers are a subset
		// of what big.Float can parse...
		return nil, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Invalid JSON number",
				Detail:   fmt.Sprintf("There is a syntax error in the given JSON number."),
				Subject:  &tok.Range,
			},
		}
	}

	return &numberVal{
		Value:    f,
		SrcRange: tok.Range,
	}, nil
}

func parseString(p *peeker) (node, zcl.Diagnostics) {
	tok := p.Read()
	var str string
	err := json.Unmarshal(tok.Bytes, &str)

	if err != nil {
		var errRange zcl.Range
		if serr, ok := err.(*json.SyntaxError); ok {
			errOfs := serr.Offset
			errPos := tok.Range.Start
			errPos.Byte += int(errOfs)

			// TODO: Use the byte offset to properly count unicode
			// characters for the column, and mark the whole of the
			// character that was wrong as part of our range.
			errPos.Column += int(errOfs)

			errEndPos := errPos
			errEndPos.Byte++
			errEndPos.Column++

			errRange = zcl.Range{
				Filename: tok.Range.Filename,
				Start:    errPos,
				End:      errEndPos,
			}
		} else {
			errRange = tok.Range
		}

		var contextRange *zcl.Range
		if errRange != tok.Range {
			contextRange = &tok.Range
		}

		// FIXME: Eventually we should parse strings directly here so
		// we can produce a more useful error message in the face fo things
		// such as invalid escapes, etc.
		return nil, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Invalid JSON string",
				Detail:   fmt.Sprintf("There is a syntax error in the given JSON string."),
				Subject:  &errRange,
				Context:  contextRange,
			},
		}
	}

	return &stringVal{
		Value:    str,
		SrcRange: tok.Range,
	}, nil
}

func parseKeyword(p *peeker) (node, zcl.Diagnostics) {
	tok := p.Read()
	s := string(tok.Bytes)

	switch s {
	case "true":
		return &booleanVal{
			Value:    true,
			SrcRange: tok.Range,
		}, nil
	case "false":
		return &booleanVal{
			Value:    false,
			SrcRange: tok.Range,
		}, nil
	case "null":
		return &nullVal{
			SrcRange: tok.Range,
		}, nil
	case "undefined", "NaN", "Infinity":
		return nil, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Invalid JSON keyword",
				Detail:   fmt.Sprintf("The JavaScript identifier %q cannot be used in JSON.", s),
				Subject:  &tok.Range,
			},
		}
	default:
		var dym string
		if suggest := keywordSuggestion(s); suggest != "" {
			dym = fmt.Sprintf(" Did you mean %q?", suggest)
		}

		return nil, zcl.Diagnostics{
			{
				Severity: zcl.DiagError,
				Summary:  "Invalid JSON keyword",
				Detail:   fmt.Sprintf("%q is not a valid JSON keyword.%s", s, dym),
				Subject:  &tok.Range,
			},
		}
	}
}
