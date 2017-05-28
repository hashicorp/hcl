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

        UTF8Cont = 0x80 .. 0xBF;
        AnyUTF8 = (
            0x00..0x7F |
            0xC0..0xDF . UTF8Cont |
            0xE0..0xEF . UTF8Cont . UTF8Cont |
            0xF0..0xF7 . UTF8Cont . UTF8Cont . UTF8Cont
        );
        BrokenUTF8 = any - AnyUTF8;

        NumberLit = digit (digit|'.'|('e'|'E') ('+'|'-')? digit)*;

        # Tabs are not valid, but we accept them in the scanner and mark them
        # as tokens so that we can produce diagnostics advising the user to
        # use spaces instead.
        Tabs = 0x09+;

        Spaces = ' '+;

        main := |*
            Spaces           => {};
            NumberLit        => { token(TokenNumberLit) };
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

    // Make Go compiler happy
    _ = ts
    _ = te
    _ = act
    _ = eof

    token := func (ty TokenType) {
        f.emitToken(ty, ts, te)
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
