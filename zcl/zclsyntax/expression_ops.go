package zclsyntax

type Operation rune

const (
	OpNil Operation = 0 // Zero value of Operation. Not a valid Operation.

	OpLogicalOr          Operation = '∨'
	OpLogicalAnd         Operation = '∧'
	OpLogicalNot         Operation = '!'
	OpEqual              Operation = '='
	OpNotEqual           Operation = '≠'
	OpGreaterThan        Operation = '>'
	OpGreaterThanOrEqual Operation = '≥'
	OpLessThan           Operation = '<'
	OpLessThanOrEqual    Operation = '≤'
	OpAdd                Operation = '+'
	OpSubtract           Operation = '-'
	OpMultiply           Operation = '*'
	OpDivide             Operation = '/'
	OpModulo             Operation = '%'
	OpNegate             Operation = '∓'
)

var binaryOps []map[TokenType]Operation

func init() {
	// This operation table maps from the operator's token type
	// to the AST operation type. All expressions produced from
	// binary operators are BinaryOp nodes.
	//
	// Binary operator groups are listed in order of precedence, with
	// the *lowest* precedence first. Operators within the same group
	// have left-to-right associativity.
	binaryOps = []map[TokenType]Operation{
		{
			TokenOr: OpLogicalOr,
		},
		{
			TokenAnd: OpLogicalAnd,
		},
		{
			TokenEqual:    OpEqual,
			TokenNotEqual: OpNotEqual,
		},
		{
			TokenGreaterThan:   OpGreaterThan,
			TokenGreaterThanEq: OpGreaterThanOrEqual,
			TokenLessThan:      OpLessThan,
			TokenLessThanEq:    OpLessThanOrEqual,
		},
		{
			TokenPlus:  OpAdd,
			TokenMinus: OpSubtract,
		},
		{
			TokenStar:    OpMultiply,
			TokenSlash:   OpDivide,
			TokenPercent: OpModulo,
		},
	}
}
