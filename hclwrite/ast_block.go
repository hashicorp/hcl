package hclwrite

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type Block struct {
	inTree

	leadComments *node
	typeName     *node
	labels       nodeSet
	open         *node
	body         *node
	close        *node
}

func newBlock() *Block {
	return &Block{
		inTree: newInTree(),
		labels: newNodeSet(),
	}
}

// NewBlock constructs a new, empty block with the given type name and labels.
func NewBlock(typeName string, labels []string) *Block {
	block := newBlock()
	block.init(typeName, labels)
	return block
}

func (b *Block) init(typeName string, labels []string) {
	nameTok := newIdentToken(typeName)
	nameObj := newIdentifier(nameTok)
	b.leadComments = b.children.Append(newComments(nil))
	b.typeName = b.children.Append(nameObj)
	for _, label := range labels {
		labelToks := TokensForValue(cty.StringVal(label))
		labelObj := newQuoted(labelToks)
		labelNode := b.children.Append(labelObj)
		b.labels.Add(labelNode)
	}
	b.open = b.children.AppendUnstructuredTokens(Tokens{
		{
			Type:  hclsyntax.TokenOBrace,
			Bytes: []byte{'{'},
		},
		{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte{'\n'},
		},
	})
	body := newBody() // initially totally empty; caller can append to it subsequently
	b.body = b.children.Append(body)
	b.close = b.children.AppendUnstructuredTokens(Tokens{
		{
			Type:  hclsyntax.TokenCBrace,
			Bytes: []byte{'}'},
		},
		{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte{'\n'},
		},
	})
}

// Body returns the body that represents the content of the receiving block.
//
// Appending to or otherwise modifying this body will make changes to the
// tokens that are generated between the blocks open and close braces.
func (b *Block) Body() *Body {
	return b.body.content.(*Body)
}

// Type returns the type name of the block.
func (b *Block) Type() string {
	typeNameObj := b.typeName.content.(*identifier)
	return string(typeNameObj.token.Bytes)
}

// SetType updates the type name of the block to a given name.
func (b *Block) SetType(typeName string) {
	nameTok := newIdentToken(typeName)
	nameObj := newIdentifier(nameTok)
	b.typeName.ReplaceWith(nameObj)
}

// Labels returns the labels of the block.
func (b *Block) Labels() []string {
	labelNames := make([]string, 0, len(b.labels))
	list := b.labels.List()

	for _, label := range list {
		switch labelObj := label.content.(type) {
		case *identifier:
			if labelObj.token.Type == hclsyntax.TokenIdent {
				labelString := string(labelObj.token.Bytes)
				labelNames = append(labelNames, labelString)
			}

		case *quoted:
			tokens := labelObj.tokens
			if len(tokens) == 3 &&
				tokens[0].Type == hclsyntax.TokenOQuote &&
				tokens[1].Type == hclsyntax.TokenQuotedLit &&
				tokens[2].Type == hclsyntax.TokenCQuote {
				// Note that TokenQuotedLit may contain escape sequences.
				labelString, diags := hclsyntax.ParseStringLiteralToken(tokens[1].asHCLSyntax())

				// If parsing the string literal returns error diagnostics
				// then we can just assume the label doesn't match, because it's invalid in some way.
				if !diags.HasErrors() {
					labelNames = append(labelNames, labelString)
				}
			}

		default:
			// If neither of the previous cases are true (should be impossible)
			// then we can just ignore it, because it's invalid too.
		}
	}

	return labelNames
}

// SetLabels updates the labels of the block to given labels.
// Since we cannot assume that old and new labels are equal in length,
// remove old labels and insert new ones before TokenOBrace.
func (b *Block) SetLabels(labels []string) {
	// Remove old labels
	for oldLabel := range b.labels {
		oldLabel.Detach()
		b.labels.Remove(oldLabel)
	}

	// Insert new labels before TokenOBrace.
	for _, label := range labels {
		labelToks := TokensForValue(cty.StringVal(label))
		// Force a new label to use the quoted form even if the old one is unquoted.
		// The unquoted form is supported in HCL 2 only for compatibility with some
		// historical use in HCL 1.
		labelObj := newQuoted(labelToks)
		labelNode := b.children.Insert(b.open, labelObj)
		b.labels.Add(labelNode)
	}
}
