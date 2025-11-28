package doddish

type Token struct {
	Contents []byte
	TokenType
}

func (token Token) String() string {
	return string(token.Contents)
}

func (token Token) Clone() (dst Token) {
	dst = token
	dst.Contents = make([]byte, len(token.Contents))
	copy(dst.Contents, token.Contents)
	return dst
}
