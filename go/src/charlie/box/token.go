package box

//go:generate stringer -type=TokenType
type TokenType int

const (
	TokenTypeIncomplete = TokenType(iota)
	TokenTypeOperator   // " =,.:+?^[]"
	TokenTypeIdentifier // ["one", "uno", "tag", "one", "type", "/browser/bookmark-1", "@sha"...]
	TokenTypeLiteral    // ["\"some text\"", "\"some text \\\" with escape\""]
)

func (expected TokenType) Match(actual Token) bool {
	return actual.TokenType == expected
}

type Token struct {
	Contents []byte
	TokenType
}

func (token Token) String() string {
	return string(token.Contents)
}

func (src Token) Clone() (dst Token) {
	dst = src
	dst.Contents = make([]byte, len(src.Contents))
	copy(dst.Contents, src.Contents)
	return
}
