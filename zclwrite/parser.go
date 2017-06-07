package zclwrite

import (
	"sort"

	"github.com/zclconf/go-zcl/zcl"
	"github.com/zclconf/go-zcl/zcl/zclsyntax"
)

// Our "parser" here is actually not doing any parsing of its own. Instead,
// it leans on the native parser in zclsyntax, and then uses the source ranges
// from the AST to partition the raw token sequence to match the raw tokens
// up to AST nodes.
//
// This strategy feels somewhat counter-intuitive, since most of the work the
// parser does is thrown away here, but this strategy is chosen because the
// normal parsing work done by zclsyntax is considered to be the "main case",
// while modifying and re-printing source is more of an edge case, used only
// in ancillary tools, and so it's good to keep all the main parsing logic
// with the main case but keep all of the extra complexity of token wrangling
// out of the main parser, which is already rather complex just serving the
// use-cases it already serves.
//
// If the parsing step produces any errors, the returned File is nil because
// we can't reliably extract tokens from the partial AST produced by an
// erroneous parse.
func parse(src []byte, filename string, start zcl.Pos) (*File, zcl.Diagnostics) {
	file, diags := zclsyntax.ParseConfig(src, filename, start)
	if diags.HasErrors() {
		return nil, diags
	}

	// To do our work here, we use the "native" tokens (those from zclsyntax)
	// to match against source ranges in the AST, but ultimately produce
	// slices from our sequence of "writer" tokens, which contain only
	// *relative* position information that is more appropriate for
	// transformation/writing use-cases.
	nativeTokens, diags := zclsyntax.LexConfig(src, filename, start)
	if diags.HasErrors() {
		// should never happen, since we would've caught these diags in
		// the first call above.
		return nil, diags
	}
	writerTokens := writerTokens(nativeTokens)

	from := inputTokens{
		nativeTokens: nativeTokens,
		writerTokens: writerTokens,
	}

	// we ignore "before" and "after" at the root, because the root body covers
	// the entire input.
	_, root, _ := parseBody(file.Body.(*zclsyntax.Body), from)

	return &File{
		Name:  filename,
		Bytes: src,

		Body: root,
	}, nil
}

type inputTokens struct {
	nativeTokens zclsyntax.Tokens
	writerTokens Tokens
}

func (it inputTokens) Partition(rng zcl.Range) (before, within, after inputTokens) {
	start, end := partitionTokens(it.nativeTokens, rng)
	before = it.Slice(0, start)
	within = it.Slice(start, end)
	after = it.Slice(end, len(it.nativeTokens))
	return
}

func (it inputTokens) Slice(start, end int) inputTokens {
	// When we slice, we create a new slice with no additional capacity because
	// we expect that these slices will be mutated in order to insert
	// new code into the AST, and we want to ensure that a new underlying
	// array gets allocated in that case, rather than writing into some
	// following slice and corrupting it.
	return inputTokens{
		nativeTokens: it.nativeTokens[start:end:end],
		writerTokens: it.writerTokens[start:end:end],
	}
}

func (it inputTokens) Len() int {
	return len(it.nativeTokens)
}

func (it inputTokens) Seq() *TokenSeq {
	return &TokenSeq{it.writerTokens}
}

// parseBody locates the given body within the given input tokens and returns
// the resulting *Body object as well as the tokens that appeared before and
// after it.
func parseBody(nativeBody *zclsyntax.Body, from inputTokens) (inputTokens, *Body, inputTokens) {
	before, within, after := from.Partition(nativeBody.SrcRange)

	// The main AST doesn't retain the original source ordering of the
	// body items, so we need to reconstruct that ordering by inspecting
	// their source ranges.
	nativeItems := make([]zclsyntax.Node, 0, len(nativeBody.Attributes)+len(nativeBody.Blocks))
	for _, nativeAttr := range nativeBody.Attributes {
		nativeItems = append(nativeItems, nativeAttr)
	}
	for _, nativeBlock := range nativeBody.Blocks {
		nativeItems = append(nativeItems, nativeBlock)
	}
	sort.Sort(nativeNodeSorter{nativeItems})

	body := &Body{
		IndentLevel: 0, // TODO: deal with this
	}

	// TODO: actually partition the native item tokens
	// For now we'll just return it all as one big Unstructured.
	unstructured := &Unstructured{
		AllTokens: within.Seq(),
	}
	body.AppendItem(unstructured)

	return before, body, after
}

// writerTokens takes a sequence of tokens as produced by the main zclsyntax
// package and transforms it into an equivalent sequence of tokens using
// this package's own token model.
//
// The resulting list contains the same number of tokens and uses the same
// indices as the input, allowing the two sets of tokens to be correlated
// by index.
func writerTokens(nativeTokens zclsyntax.Tokens) Tokens {
	// Ultimately we want a slice of token _pointers_, but since we can
	// predict how much memory we're going to devote to tokens we'll allocate
	// it all as a single flat buffer and thus give the GC less work to do.
	tokBuf := make([]Token, len(nativeTokens))
	var lastByteOffset int
	for i, mainToken := range nativeTokens {
		// Create a copy of the bytes so that we can mutate without
		// corrupting the original token stream.
		bytes := make([]byte, len(mainToken.Bytes))
		copy(bytes, mainToken.Bytes)

		tokBuf[i] = Token{
			Type:  mainToken.Type,
			Bytes: bytes,

			// We assume here that spaces are always ASCII spaces, since
			// that's what the scanner also assumes, and thus the number
			// of bytes skipped is also the number of space characters.
			SpacesBefore: mainToken.Range.Start.Byte - lastByteOffset,
		}

		lastByteOffset = mainToken.Range.End.Byte
	}

	// Now make a slice of pointers into the previous slice.
	ret := make(Tokens, len(tokBuf))
	for i := range ret {
		ret[i] = &tokBuf[i]
	}

	return ret
}

// partitionTokens takes a sequence of tokens and a zcl.Range and returns
// two indices within the token sequence that correspond with the range
// boundaries, such that the slice operator could be used to produce
// three token sequences for before, within, and after respectively:
//
//     start, end := partitionTokens(toks, rng)
//     before := toks[:start]
//     within := toks[start:end]
//     after := toks[end:]
//
// This works best when the range is aligned with token boundaries (e.g.
// because it was produced in terms of the scanner's result) but if that isn't
// true then it will make a best effort that may produce strange results at
// the boundaries.
//
// Native zclsyntax tokens are used here, because they contain the necessary
// absolute position information. However, since writerTokens produces a
// correlatable sequence of writer tokens, the resulting indices can be
// used also to index into its result, allowing the partitioning of writer
// tokens to be driven by the partitioning of native tokens.
//
// The tokens are assumed to be in source order and non-overlapping, which
// will be true if the token sequence from the scanner is used directly.
func partitionTokens(toks zclsyntax.Tokens, rng zcl.Range) (start, end int) {
	// We us a linear search here because we assume tha in most cases our
	// target range is close to the beginning of the sequence, and the seqences
	// are generally small for most reasonable files anyway.
	for i := 0; ; i++ {
		if i >= len(toks) {
			// No tokens for the given range at all!
			return len(toks), len(toks)
		}

		if toks[i].Range.ContainsOffset(rng.Start.Byte) {
			start = i
			break
		}

		if toks[i].Range.Start.Byte < rng.Start.Byte && toks[i].Range.End.Byte < rng.Start.Byte {
			// We overshot, it seems.
			start = i
			break
		}
	}

	for i := start; ; i++ {
		if i >= len(toks) {
			// The range "hangs off" the end of the token sequence
			return start, len(toks)
		}

		if toks[i].Range.End.Byte >= rng.End.Byte {
			end = i + 1 // end marker is exclusive
			break
		}
	}

	return start, end
}

// lexConfig uses the zclsyntax scanner to get a token stream and then
// rewrites it into this package's token model.
//
// Any errors produced during scanning are ignored, so the results of this
// function should be used with care.
func lexConfig(src []byte) Tokens {
	mainTokens, _ := zclsyntax.LexConfig(src, "", zcl.Pos{Byte: 0, Line: 1, Column: 1})
	return writerTokens(mainTokens)
}
