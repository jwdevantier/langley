package tokenizer

// TokenType interface
type TokenType interface {
	Type() string
	CountNewLines() bool
}

type tokenType struct {
	typ string
}

// Ensure tokenType implements TokenType interface
// (see Go FAQ)
var _ TokenType = tokenType{}

// var _ TokenType = (*tokenType)(nil) to check ptr

var (
	// Error signals error during tokenization
	Error = tokenType{"ERR"}
	// EOF signals the end of the token stream
	EOF = tokenType{"EOF"}
)

func (tt tokenType) Type() string {
	return tt.typ
}

func (tt tokenType) CountNewLines() bool {
	return false
}
