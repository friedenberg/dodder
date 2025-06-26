package repo

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

//go:generate stringer -type=RemoteType
type RemoteType int

// TODO rename to ConnectionType

const (
	RemoteTypeUnspecified = RemoteType(iota)
	RemoteTypeNativeDotenvXDG
	RemoteTypeSocketUnix
	RemoteTypeUrl
	RemoteTypeStdioLocal
	RemoteTypeStdioSSH
	_RemoteTypeMax
)

func GetAllRemoteTypes() []RemoteType {
	types := make([]RemoteType, 0)

	for i := RemoteTypeUnspecified + 1; i < _RemoteTypeMax; i++ {
		types = append(types, RemoteType(i))
	}

	return types
}

func (t *RemoteType) Set(v string) (err error) {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "", "none", "unspecified":
		*t = RemoteTypeUnspecified

	case "native-dotenv-xdg":
		*t = RemoteTypeNativeDotenvXDG

	case "socket-unix":
		*t = RemoteTypeSocketUnix

	case "url":
		*t = RemoteTypeUrl

	case "stdio-local":
		*t = RemoteTypeStdioLocal

	case "stdio-ssh":
		*t = RemoteTypeStdioSSH

	default:
		err = errors.ErrorWithStackf("unsupported remote type: %q", v)
		return
	}

	return
}
