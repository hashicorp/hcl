package zclsyntax

import (
    "github.com/zclconf/go-zcl/zcl"
)

// This file is generated from scan_tokens.rl. DO NOT EDIT.
%%{
  # (except you are actually in scan_tokens.rl here, so edit away!)

  machine zcltok;
  write data;
}%%

func scanTokens(data []byte, filename string, start zcl.Pos) []Token {
    f := &tokenAccum{
        Filename: filename,
        Bytes:    data,
        Pos:      start,
    }

    %%{
        include UnicodeDerived "unicode_derived.rl";

        UTF8Cont = 0x80 .. 0xBF;
        AnyUTF8 = (
            0x00..0x7F |
            0xC0..0xDF . UTF8Cont |
            0xE0..0xEF . UTF8Cont . UTF8Cont |
            0xF0..0xF7 . UTF8Cont . UTF8Cont . UTF8Cont
        );
        BrokenUTF8 = any - AnyUTF8;

        NumberLit = digit (digit|'.'|('e'|'E') ('+'|'-')? digit)*;
        Ident = ID_Start ID_Continue*;

        # Symbols that just represent themselves are handled as a single rule.
        SelfToken = "[" | "]" | "(" | ")" | "." | "*" | "/" | "+" | "-" | "=" | "<" | ">" | "!" | "?" | ":" | "\n" | "&" | "|" | "~" | "^" | ";" | "`";

        NotEqual = "!=";
        GreaterThanEqual = ">=";
        LessThanEqual = "<=";
        LogicalAnd = "&&";
        LogicalOr = "||";

        Newline = '\r' ? '\n';

        BeginStringTmpl = '"';

        # Tabs are not valid, but we accept them in the scanner and mark them
        # as tokens so that we can produce diagnostics advising the user to
        # use spaces instead.
        Tabs = 0x09+;

        Spaces = ' '+;

        action beginStringTemplate {
            token(TokenOQuote);
            fcall stringTemplate;
        }

        action endStringTemplate {
            token(TokenCQuote);
            fret;
        }

        action beginTemplateInterp {
            token(TokenTemplateInterp);
            braces++;
            retBraces = append(retBraces, braces);
            fcall main;
        }

        action beginTemplateControl {
            token(TokenTemplateControl);
            braces++;
            retBraces = append(retBraces, braces);
            fcall main;
        }

        action openBrace {
            token(TokenOBrace);
            braces++;
        }

        action closeBrace {
            if len(retBraces) > 0 && retBraces[len(retBraces)-1] == braces {
                token(TokenTemplateSeqEnd);
                braces--;
                retBraces = retBraces[0:len(retBraces)-1]
                fret;
            } else {
                token(TokenCBrace);
                braces--;
            }
        }

        action closeTemplateSeqEatWhitespace {
            token(TokenTemplateSeqEnd);
            braces--;

            // Only consume from the retBraces stack and return if we are at
            // a suitable brace nesting level, otherwise things will get
            // confused. (Not entering this branch indicates a syntax error,
            // which we will catch in the parser.)
            if len(retBraces) > 0 && retBraces[len(retBraces)-1] == braces {
                retBraces = retBraces[0:len(retBraces)-1]
                fret;
            }
        }

        TemplateInterp = "${" ("~")?;
        TemplateControl = "!{" ("~")?;
        EndStringTmpl = '"';
        TemplateStringLiteral = (
            ('$' ^'{') |
            ('!' ^'{') |
            ('\\' AnyUTF8) |
            (AnyUTF8 - ("$" | "!" | '"'))
        )+;

        stringTemplate := |*
            TemplateInterp        => beginTemplateInterp;
            TemplateControl       => beginTemplateControl;
            EndStringTmpl         => endStringTemplate;
            TemplateStringLiteral => { token(TokenStringLit); };
            BrokenUTF8            => { token(TokenBadUTF8); };
        *|;

        main := |*
            Spaces           => {};
            NumberLit        => { token(TokenNumberLit) };
            Ident            => { token(TokenIdent) };

            NotEqual         => { token(TokenNotEqual); };
            GreaterThanEqual => { token(TokenGreaterThanEq); };
            LessThanEqual    => { token(TokenLessThanEq); };
            LogicalAnd       => { token(TokenAnd); };
            LogicalOr        => { token(TokenOr); };
            SelfToken        => { selfToken() };

            "{"              => openBrace;
            "}"              => closeBrace;

            "~}"             => closeTemplateSeqEatWhitespace;

            BeginStringTmpl  => beginStringTemplate;

            Tabs             => { token(TokenTabs) };
            AnyUTF8          => { token(TokenInvalid) };
            BrokenUTF8       => { token(TokenBadUTF8) };
        *|;

    }%%

    // Ragel state
	cs := 0 // Current State
	p := 0  // "Pointer" into data
	pe := len(data) // End-of-data "pointer"
    ts := 0
    te := 0
    act := 0
    eof := pe
    var stack []int
    var top int

    braces := 0
    var retBraces []int // stack of brace levels that cause us to use fret

    %%{
        prepush {
            stack = append(stack, 0);
        }
        postpop {
            stack = stack[:len(stack)-1];
        }
    }%%

    // Make Go compiler happy
    _ = ts
    _ = te
    _ = act
    _ = eof

    token := func (ty TokenType) {
        f.emitToken(ty, ts, te)
    }
    selfToken := func () {
        b := data[ts:te]
        if len(b) != 1 {
            // should never happen
            panic("selfToken only works for single-character tokens")
        }
        f.emitToken(TokenType(b[0]), ts, te)
    }

    %%{
        write init;
        write exec;
    }%%

    // If we fall out here without being in a final state then we've
    // encountered something that the scanner can't match, which we'll
    // deal with as an invalid.
    if cs < zcltok_first_final {
        f.emitToken(TokenInvalid, p, len(data))
    }

    // We always emit a synthetic EOF token at the end, since it gives the
    // parser position information for an "unexpected EOF" diagnostic.
    f.emitToken(TokenEOF, len(data), len(data))

    return f.Tokens
}
