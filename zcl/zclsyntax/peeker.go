package zclsyntax

type peeker struct {
	Tokens    Tokens
	NextIndex int

	IncludeComments      bool
	IncludeNewlinesStack []bool
}

func newPeeker(tokens Tokens, includeComments bool) *peeker {
	return &peeker{
		Tokens:          tokens,
		IncludeComments: includeComments,

		IncludeNewlinesStack: []bool{true},
	}
}

func (p *peeker) Peek() Token {
	ret, _ := p.nextToken()
	return ret
}

func (p *peeker) Read() Token {
	ret, nextIdx := p.nextToken()
	p.NextIndex = nextIdx
	return ret
}

func (p *peeker) nextToken() (Token, int) {
	for i := p.NextIndex; i < len(p.Tokens); i++ {
		tok := p.Tokens[i]
		switch tok.Type {
		case TokenComment:
			if !p.IncludeComments {
				continue
			}
		case TokenNewline:
			if !p.includingNewlines() {
				continue
			}
		}

		return tok, i + 1
	}

	// if we fall out here then we'll return the EOF token, and leave
	// our index pointed off the end of the array so we'll keep
	// returning EOF in future too.
	return p.Tokens[len(p.Tokens)-1], len(p.Tokens)
}

func (p *peeker) includingNewlines() bool {
	return p.IncludeNewlinesStack[len(p.IncludeNewlinesStack)-1]
}

func (p *peeker) PushIncludeNewlines(include bool) {
	p.IncludeNewlinesStack = append(p.IncludeNewlinesStack, include)
}

func (p *peeker) PopIncludeNewlines() bool {
	stack := p.IncludeNewlinesStack
	remain, ret := stack[:len(stack)-1], stack[len(stack)-1]
	p.IncludeNewlinesStack = remain
	return ret
}
