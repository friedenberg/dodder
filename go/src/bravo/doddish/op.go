package doddish

import "unicode"

type operatorType byte

const (
	operatorTypeUnknown                        = iota
	operatorTypeSoloSeq                        // {,}
	operatorTypeMixedSeq                       // {one,/,two}
	operatorTypePrefixSeqOrSeparateTokenWithin // {id,=,value} or {=,tag}
	operatorTypePrefixSeqOrInlineIdentifier    // {-,tag} or {tag-with-stuff}
)

func (opType operatorType) isSoloOrPrefix() bool {
	switch opType {
	default:
		return false

	case operatorTypeSoloSeq, operatorTypePrefixSeqOrSeparateTokenWithin:
		return true
	}
}

func (opType operatorType) isMixedOrPrefix() bool {
	switch opType {
	default:
		return false

	case operatorTypeMixedSeq, operatorTypePrefixSeqOrSeparateTokenWithin:
		return true
	}
}

// TODO switch to byte
type Op rune

const (
	OpOr            = Op(',')
	OpAnd           = Op(' ')
	OpGroupOpen     = Op('[')
	OpGroupClose    = Op(']')
	OpNegation      = Op('^')
	OpExact         = Op('=')
	OpNewline       = Op('\n')
	OpSigilLatest   = Op(':')
	OpSigilHistory  = Op('+')
	OpSigilExternal = Op('.')
	OpSigilHidden   = Op('?')
	OpPathSeparator = Op('/')
	OpType          = Op('!')
	OpVirtual       = Op('%')
	OpBlob          = Op('@')
)

func MakeOp(char rune) (Op, operatorType) {
	op := Op(char)
	return op, op.GetType()
}

func (op Op) ToRune() rune {
	return rune(op)
}

func (op Op) ToByte() byte {
	return byte(op)
}

func (op Op) ToBytes() []byte {
	return []byte{op.ToByte()}
}

func (op Op) GetType() operatorType {
	switch op {
	default:
		return operatorTypeUnknown

	case
		OpAnd,
		OpGroupClose,
		OpGroupOpen,
		OpNegation,
		OpNewline,
		OpOr,
		OpSigilHidden,
		OpSigilHistory,
		OpSigilLatest:
		return operatorTypeSoloSeq

	case
		OpBlob,
		OpPathSeparator,
		OpType,
		OpVirtual:
		return operatorTypeMixedSeq

	case OpExact, OpSigilExternal:
		return operatorTypePrefixSeqOrSeparateTokenWithin

		// case OpTagSeparator:
		// 	return operatorTypeUnknown
		// return operatorTypePrefixSeq
		// return operatorTypeMixedSeq
	}
}

// Is this an operator, and should it take up an entire seq?
func (op Op) isSoloSeqOp(dotAllowed bool) bool {
	if dotAllowed && op == OpSigilExternal {
		return false
	}

	return op.GetType().isSoloOrPrefix()
}

// Is this an operator, and can it appear within a sequence with other tokens?
func (op Op) isMixedSeqOp() bool {
	return op.GetType().isMixedOrPrefix()
}

func (op Op) isSpace() bool {
	return unicode.IsSpace(op.ToRune())
}
