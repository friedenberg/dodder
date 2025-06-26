package repo

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

//go:generate stringer -type=RemoteServeType
type RemoteServeType int

const (
	RemoteServeTypeUnspecified = RemoteServeType(iota)
	RemoteServeTypeSocketUnix
	RemoteServeTypePortHTTP
	RemoteServeTypeStdioLocal
)

func (t *RemoteServeType) Set(v string) (err error) {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "", "none", "unspecified":
		*t = RemoteServeTypeUnspecified

	case "socket-unix":
		*t = RemoteServeTypeSocketUnix

	case "port-http":
		*t = RemoteServeTypePortHTTP

	case "stdio-local":
		*t = RemoteServeTypeStdioLocal

	default:
		err = errors.ErrorWithStackf("unsupported remote type: %q", v)
		return
	}

	return
}
