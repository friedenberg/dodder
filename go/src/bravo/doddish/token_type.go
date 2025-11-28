package doddish

//go:generate stringer -type=TokenType
type TokenType int

const (
	TokenTypeIncomplete = TokenType(iota)
	TokenTypeOperator   // " =,.:+?^[]"
	TokenTypeIdentifier // ["one", "uno", "tag", "one", "type", "/browser/bookmark-1", "sha"...]
	TokenTypeLiteral    // ["\"some text\"", "\"some text \\\" with escape\""]
)

func (expected TokenType) Match(actual Token) bool {
	return actual.TokenType == expected
}
