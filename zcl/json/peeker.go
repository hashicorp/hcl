package json

type peeker struct {
	tokens []token
	pos    int
}

func newPeeker(tokens []token) *peeker {
	return &peeker{
		tokens: tokens,
		pos:    0,
	}
}

func (p *peeker) Peek() token {
	return p.tokens[p.pos]
}

func (p *peeker) Read() token {
	if p.tokens[p.pos].Type != tokenEOF {
		p.pos++
	}
	return p.tokens[p.pos]
}
