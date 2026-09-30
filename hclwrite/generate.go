// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package hclwrite

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type HandlingStrategy = int

const (
	AsLiteral  = 0
	AsTemplate = 1
)

type Handling struct {
	String HandlingStrategy
}

type GenerateOptions struct {
	Handling Handling
}

type Generator struct{ GenerateOptions }

var defaultGenerator = Generator{
	GenerateOptions{
		Handling{
			String: AsLiteral,
		},
	},
}

// TokensForValue returns a sequence of tokens that represents the given
// constant value.
//
// This function only supports types that are used by HCL. In particular, it
// does not support capsule types and will panic if given one.
//
// It is not possible to express an unknown value in source code, so this
// function will panic if the given value is unknown or contains any unknown
// values. A caller can call the value's IsWhollyKnown method to verify that
// no unknown values are present before calling TokensForValue.
func TokensForValue(val cty.Value) Tokens {
	return defaultGenerator.TokensForValue(val)
}

// TokensForValue returns a sequence of tokens that represents the given
// constant value.
//
// This function only supports types that are used by HCL. In particular, it
// does not support capsule types and will panic if given one.
//
// It is not possible to express an unknown value in source code, so this
// function will panic if the given value is unknown or contains any unknown
// values. A caller can call the value's IsWhollyKnown method to verify that
// no unknown values are present before calling TokensForValue.
func (g *Generator) TokensForValue(val cty.Value) Tokens {
	toks := g.appendTokensForValue(val, nil)
	format(toks) // fiddle with the SpacesBefore field to get canonical spacing
	return toks
}

// TokensForTraversal returns a sequence of tokens that represents the given
// traversal.
//
// If the traversal is absolute then the result is a self-contained, valid
// reference expression. If the traversal is relative then the returned tokens
// could be appended to some other expression tokens to traverse into the
// represented expression.
func TokensForTraversal(traversal hcl.Traversal) Tokens {
	return defaultGenerator.TokensForTraversal(traversal)
}

// TokensForTraversal returns a sequence of tokens that represents the given
// traversal.
//
// If the traversal is absolute then the result is a self-contained, valid
// reference expression. If the traversal is relative then the returned tokens
// could be appended to some other expression tokens to traverse into the
// represented expression.
func (g *Generator) TokensForTraversal(traversal hcl.Traversal) Tokens {
	toks := g.appendTokensForTraversal(traversal, nil)
	format(toks) // fiddle with the SpacesBefore field to get canonical spacing
	return toks
}

// TokensForIdentifier returns a sequence of tokens representing just the
// given identifier.
//
// In practice this function can only ever generate exactly one token, because
// an identifier is always a leaf token in the syntax tree.
//
// This is similar to calling TokensForTraversal with a single-step absolute
// traversal, but avoids the need to construct a separate traversal object
// for this simple common case. If you need to generate a multi-step traversal,
// use TokensForTraversal instead.
func TokensForIdentifier(name string) Tokens {
	return defaultGenerator.TokensForIdentifier(name)
}

// TokensForIdentifier returns a sequence of tokens representing just the
// given identifier.
//
// In practice this function can only ever generate exactly one token, because
// an identifier is always a leaf token in the syntax tree.
//
// This is similar to calling TokensForTraversal with a single-step absolute
// traversal, but avoids the need to construct a separate traversal object
// for this simple common case. If you need to generate a multi-step traversal,
// use TokensForTraversal instead.
func (g *Generator) TokensForIdentifier(name string) Tokens {
	return Tokens{
		newIdentToken(name),
	}
}

// TokensForTuple returns a sequence of tokens that represents a tuple
// constructor, with element expressions populated from the given list
// of tokens.
//
// TokensForTuple includes the given elements verbatim into the element
// positions in the resulting tuple expression, without any validation to
// ensure that they represent valid expressions. Use TokensForValue or
// TokensForTraversal to generate valid leaf expression values, or use
// TokensForTuple, TokensForObject, and TokensForFunctionCall to
// generate other nested compound expressions.
func TokensForTuple(elems []Tokens) Tokens {
	return defaultGenerator.TokensForTuple(elems)
}

// TokensForTuple returns a sequence of tokens that represents a tuple
// constructor, with element expressions populated from the given list
// of tokens.
//
// TokensForTuple includes the given elements verbatim into the element
// positions in the resulting tuple expression, without any validation to
// ensure that they represent valid expressions. Use TokensForValue or
// TokensForTraversal to generate valid leaf expression values, or use
// TokensForTuple, TokensForObject, and TokensForFunctionCall to
// generate other nested compound expressions.
func (g *Generator) TokensForTuple(elems []Tokens) Tokens {
	var toks Tokens
	toks = append(toks, &Token{
		Type:  hclsyntax.TokenOBrack,
		Bytes: []byte{'['},
	})
	for index, elem := range elems {
		if index > 0 {
			toks = append(toks, &Token{
				Type:  hclsyntax.TokenComma,
				Bytes: []byte{','},
			})
		}
		toks = append(toks, elem...)
	}

	toks = append(toks, &Token{
		Type:  hclsyntax.TokenCBrack,
		Bytes: []byte{']'},
	})

	format(toks) // fiddle with the SpacesBefore field to get canonical spacing
	return toks
}

// TokensForObject returns a sequence of tokens that represents an object
// constructor, with attribute name/value pairs populated from the given
// list of attribute token objects.
//
// TokensForObject includes the given tokens verbatim into the name and
// value positions in the resulting object expression, without any validation
// to ensure that they represent valid expressions. Use TokensForValue or
// TokensForTraversal to generate valid leaf expression values, or use
// TokensForTuple, TokensForObject, and TokensForFunctionCall to
// generate other nested compound expressions.
//
// Note that HCL requires placing a traversal expression in parentheses if
// you intend to use it as an attribute name expression, because otherwise
// the parser will interpret it as a literal attribute name. TokensForObject
// does not handle that situation automatically, so a caller must add the
// necessary `TokenOParen` and TokenCParen` manually if needed.
func TokensForObject(attrs []ObjectAttrTokens) Tokens {
	return defaultGenerator.TokensForObject(attrs)
}

// TokensForObject returns a sequence of tokens that represents an object
// constructor, with attribute name/value pairs populated from the given
// list of attribute token objects.
//
// TokensForObject includes the given tokens verbatim into the name and
// value positions in the resulting object expression, without any validation
// to ensure that they represent valid expressions. Use TokensForValue or
// TokensForTraversal to generate valid leaf expression values, or use
// TokensForTuple, TokensForObject, and TokensForFunctionCall to
// generate other nested compound expressions.
//
// Note that HCL requires placing a traversal expression in parentheses if
// you intend to use it as an attribute name expression, because otherwise
// the parser will interpret it as a literal attribute name. TokensForObject
// does not handle that situation automatically, so a caller must add the
// necessary `TokenOParen` and TokenCParen` manually if needed.
func (g *Generator) TokensForObject(attrs []ObjectAttrTokens) Tokens {
	var toks Tokens
	toks = append(toks, &Token{
		Type:  hclsyntax.TokenOBrace,
		Bytes: []byte{'{'},
	})
	if len(attrs) > 0 {
		toks = append(toks, &Token{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte{'\n'},
		})
	}
	for _, attr := range attrs {
		toks = append(toks, attr.Name...)
		toks = append(toks, &Token{
			Type:  hclsyntax.TokenEqual,
			Bytes: []byte{'='},
		})
		toks = append(toks, attr.Value...)
		toks = append(toks, &Token{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte{'\n'},
		})
	}
	toks = append(toks, &Token{
		Type:  hclsyntax.TokenCBrace,
		Bytes: []byte{'}'},
	})

	format(toks) // fiddle with the SpacesBefore field to get canonical spacing
	return toks
}

// TokensForFunctionCall returns a sequence of tokens that represents call
// to the function with the given name, using the argument tokens to
// populate the argument expressions.
//
// TokensForFunctionCall includes the given argument tokens verbatim into the
// positions in the resulting call expression, without any validation
// to ensure that they represent valid expressions. Use TokensForValue or
// TokensForTraversal to generate valid leaf expression values, or use
// TokensForTuple, TokensForObject, and TokensForFunctionCall to
// generate other nested compound expressions.
//
// This function doesn't include an explicit way to generate the expansion
// symbol "..." on the final argument. Currently, generating that requires
// manually appending a TokenEllipsis with the bytes "..." to the tokens for
// the final argument.
func TokensForFunctionCall(funcName string, args ...Tokens) Tokens {
	return defaultGenerator.TokensForFunctionCall(funcName, args...)
}

// TokensForFunctionCall returns a sequence of tokens that represents call
// to the function with the given name, using the argument tokens to
// populate the argument expressions.
//
// TokensForFunctionCall includes the given argument tokens verbatim into the
// positions in the resulting call expression, without any validation
// to ensure that they represent valid expressions. Use TokensForValue or
// TokensForTraversal to generate valid leaf expression values, or use
// TokensForTuple, TokensForObject, and TokensForFunctionCall to
// generate other nested compound expressions.
//
// This function doesn't include an explicit way to generate the expansion
// symbol "..." on the final argument. Currently, generating that requires
// manually appending a TokenEllipsis with the bytes "..." to the tokens for
// the final argument.
func (g *Generator) TokensForFunctionCall(funcName string, args ...Tokens) Tokens {
	var toks Tokens
	toks = append(toks, TokensForIdentifier(funcName)...)
	toks = append(toks, &Token{
		Type:  hclsyntax.TokenOParen,
		Bytes: []byte{'('},
	})
	for index, arg := range args {
		if index > 0 {
			toks = append(toks, &Token{
				Type:  hclsyntax.TokenComma,
				Bytes: []byte{','},
			})
		}
		toks = append(toks, arg...)
	}
	toks = append(toks, &Token{
		Type:  hclsyntax.TokenCParen,
		Bytes: []byte{')'},
	})

	format(toks) // fiddle with the SpacesBefore field to get canonical spacing
	return toks
}

func (g *Generator) appendTokensForValue(val cty.Value, toks Tokens) Tokens {
	switch {

	case !val.IsKnown():
		panic("cannot produce tokens for unknown value")

	case val.IsNull():
		toks = append(toks, &Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte(`null`),
		})

	case val.Type() == cty.Bool:
		var src []byte
		if val.True() {
			src = []byte(`true`)
		} else {
			src = []byte(`false`)
		}
		toks = append(toks, &Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: src,
		})

	case val.Type() == cty.Number:
		bf := val.AsBigFloat()
		srcStr := bf.Text('f', -1)
		toks = append(toks, &Token{
			Type:  hclsyntax.TokenNumberLit,
			Bytes: []byte(srcStr),
		})

	case val.Type() == cty.String:
		s := val.AsString()
		var content Tokens
		if len(s) == 0 {
			content = Tokens{}
		} else {
			switch strategy := g.GenerateOptions.Handling.String; strategy {
			case AsLiteral:
				src := escapeQuotedStringLit(s)
				content = Tokens{&Token{
					Type:  hclsyntax.TokenQuotedLit,
					Bytes: src,
				}}
			case AsTemplate:
				syntaxTokens, diag := hclsyntax.LexTemplate([]byte(s), s, hcl.Pos{})

				if len(diag.Errs()) > 0 {
					panic(fmt.Sprintf(`cannot lex "%s" as template`, s))
				}
				content = make(Tokens, 0, len(syntaxTokens)-1)
				var n int
				for i, v := range syntaxTokens {
					if v.Type == hclsyntax.TokenEOF {
						break
					}
					n = i
					content = append(content, &Token{Type: v.Type, Bytes: v.Bytes})
				}
				if n > 0 && content[n].Type == hclsyntax.TokenTemplateSeqEnd {
					findTemplateSequence := func(start int) int {
						for i, v := range content[start:] {
							if v.Type == hclsyntax.TokenTemplateControl || v.Type == hclsyntax.TokenTemplateInterp {
								return i
							}
						}
						return -1
					}
					if findTemplateSequence(0) == 0 && findTemplateSequence(1) < 0 {
						// i.e., the entire string was a single template control or interpolation
						return append(toks, content[1:n]...)
					}
				}
			default:
				panic(fmt.Sprintf("Unknown string handling strategy %d", strategy))
			}
		}
		// TODO: If it's a multi-line string ending in a newline, format
		// it as a HEREDOC instead.
		toks = append(toks, &Token{
			Type:  hclsyntax.TokenOQuote,
			Bytes: []byte{'"'},
		})
		toks = append(toks, content...)
		toks = append(toks, &Token{
			Type:  hclsyntax.TokenCQuote,
			Bytes: []byte{'"'},
		})

	case val.Type().IsListType() || val.Type().IsSetType() || val.Type().IsTupleType():
		toks = append(toks, &Token{
			Type:  hclsyntax.TokenOBrack,
			Bytes: []byte{'['},
		})

		i := 0
		for it := val.ElementIterator(); it.Next(); {
			if i > 0 {
				toks = append(toks, &Token{
					Type:  hclsyntax.TokenComma,
					Bytes: []byte{','},
				})
			}
			_, eVal := it.Element()
			toks = g.appendTokensForValue(eVal, toks)
			i++
		}

		toks = append(toks, &Token{
			Type:  hclsyntax.TokenCBrack,
			Bytes: []byte{']'},
		})

	case val.Type().IsMapType() || val.Type().IsObjectType():
		toks = append(toks, &Token{
			Type:  hclsyntax.TokenOBrace,
			Bytes: []byte{'{'},
		})
		if val.LengthInt() > 0 {
			toks = append(toks, &Token{
				Type:  hclsyntax.TokenNewline,
				Bytes: []byte{'\n'},
			})
		}

		i := 0
		for it := val.ElementIterator(); it.Next(); {
			eKey, eVal := it.Element()
			if hclsyntax.ValidIdentifier(eKey.AsString()) {
				toks = append(toks, &Token{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte(eKey.AsString()),
				})
			} else {
				toks = g.appendTokensForValue(eKey, toks)
			}
			toks = append(toks, &Token{
				Type:  hclsyntax.TokenEqual,
				Bytes: []byte{'='},
			})
			toks = g.appendTokensForValue(eVal, toks)
			toks = append(toks, &Token{
				Type:  hclsyntax.TokenNewline,
				Bytes: []byte{'\n'},
			})
			i++
		}

		toks = append(toks, &Token{
			Type:  hclsyntax.TokenCBrace,
			Bytes: []byte{'}'},
		})

	default:
		panic(fmt.Sprintf("cannot produce tokens for %#v", val))
	}

	return toks
}

func (g *Generator) appendTokensForTraversal(traversal hcl.Traversal, toks Tokens) Tokens {
	for _, step := range traversal {
		toks = g.appendTokensForTraversalStep(step, toks)
	}
	return toks
}

func (g *Generator) appendTokensForTraversalStep(step hcl.Traverser, toks Tokens) Tokens {
	switch ts := step.(type) {
	case hcl.TraverseRoot:
		toks = append(toks, &Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte(ts.Name),
		})
	case hcl.TraverseAttr:
		toks = append(
			toks,
			&Token{
				Type:  hclsyntax.TokenDot,
				Bytes: []byte{'.'},
			},
			&Token{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(ts.Name),
			},
		)
	case hcl.TraverseIndex:
		toks = append(toks, &Token{
			Type:  hclsyntax.TokenOBrack,
			Bytes: []byte{'['},
		})
		toks = g.appendTokensForValue(ts.Key, toks)
		toks = append(toks, &Token{
			Type:  hclsyntax.TokenCBrack,
			Bytes: []byte{']'},
		})
	default:
		panic(fmt.Sprintf("unsupported traversal step type %T", step))
	}

	return toks
}

func escapeQuotedStringLit(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	buf := make([]byte, 0, len(s))
	for i, r := range s {
		switch r {
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		case '"':
			buf = append(buf, '\\', '"')
		case '\\':
			buf = append(buf, '\\', '\\')
		case '$', '%':
			buf = appendRune(buf, r)
			remain := s[i+1:]
			if len(remain) > 0 && remain[0] == '{' {
				// Double up our template introducer symbol to escape it.
				buf = appendRune(buf, r)
			}
		default:
			if !unicode.IsPrint(r) {
				var fmted string
				if r < 65536 {
					fmted = fmt.Sprintf("\\u%04x", r)
				} else {
					fmted = fmt.Sprintf("\\U%08x", r)
				}
				buf = append(buf, fmted...)
			} else {
				buf = appendRune(buf, r)
			}
		}
	}
	return buf
}

func appendRune(b []byte, r rune) []byte {
	l := utf8.RuneLen(r)
	for i := 0; i < l; i++ {
		b = append(b, 0) // make room at the end of our buffer
	}
	ch := b[len(b)-l:]
	utf8.EncodeRune(ch, r)
	return b
}
