package blech32

import "errors"

var (
	ErrEmptyHRP         = errors.New("empty HRP")
	ErrSeparatorMissing = errors.New("separator (`-`) missing")
)
