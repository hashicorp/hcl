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
    offset := 0

    f := &tokenAccum{
        Filename: filename,
        Bytes:    data,
        Start:    start,
    }

    %%{
        action start {
            offset = p
            fgoto token;
        }

        action EmitInvalid {
            f.emitToken(TokenInvalid, offset, p+1)
        }

        action EmitBadUTF8 {
            f.emitToken(TokenBadUTF8, offset, p+1)
        }

        action EmitEOF {
            f.emitToken(TokenEOF, offset, offset)
        }

        UTF8Cont = 0x80 .. 0xBF;
        AnyUTF8 = (
            0x00..0x7F |
            0xC0..0xDF . UTF8Cont |
            0xE0..0xEF . UTF8Cont . UTF8Cont |
            0xF0..0xF7 . UTF8Cont . UTF8Cont . UTF8Cont
        );

        AnyUTF8Tok = AnyUTF8 >start;
        BrokenUTF8 = any - AnyUTF8;
        EmptyTok = "";

        # Tabs are not valid, but we accept them in the scanner and mark them
        # as tokens so that we can produce diagnostics advising the user to
        # use spaces instead.
        TabTok = 0x09 >start;

        token := |*
            AnyUTF8    => EmitInvalid;
            BrokenUTF8 => EmitBadUTF8;
            EmptyTok   => EmitEOF;
        *|;

        Spaces = ' '*;

        main := Spaces @start;

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

    return f.Tokens
}
