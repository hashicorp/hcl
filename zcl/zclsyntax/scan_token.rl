package zclsyntax

import (
    "github.com/zclconf/go-zcl/zcl"
)

// This file is generated from scan_token.rl. DO NOT EDIT.
%%{
  # (except you are actually in scan_token.rl here, so edit away!)

  machine zcltok;
  write data;
}%%

func nextToken(data []byte, filename string, start zcl.Pos) (Token, []byte) {
    offset := 0

    f := tokenFactory{
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
            return f.makeToken(TokenInvalid, offset, p+1)
        }

        action EmitBadUTF8 {
            return f.makeToken(TokenBadUTF8, offset, p+1)
        }

        action EmitEOF {
            return f.makeToken(TokenEOF, offset, offset)
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

    // If we fall out here then we'll just classify the remainder of the
    // file as invalid.
    return f.makeToken(TokenInvalid, 0, len(data))
}
