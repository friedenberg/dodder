package box

type TokenMatcher interface {
	Match(Token) bool
}

type TokenMatcherOp byte

func (tokenMatcher TokenMatcherOp) Match(token Token) bool {
	if token.TokenType != TokenTypeOperator {
		return false
	}

	if token.Contents[0] != byte(tokenMatcher) {
		return false
	}

	return true
}

func TokenMatcherOr(tm ...TokenMatcher) tokenMatcherOr {
	return tokenMatcherOr(tm)
}

type tokenMatcherOr []TokenMatcher

func (tokenMatcher tokenMatcherOr) Match(token Token) bool {
	for _, t := range tokenMatcher {
		if t.Match(token) {
			return true
		}
	}

	return false
}
