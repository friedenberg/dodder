package repo

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

//go:generate stringer -type=RemoteConnectionType
type RemoteConnectionType int

const (
	RemoteConnectionTypeUnspecified = RemoteConnectionType(iota)
	RemoteConnectionTypeNativeDotenvXDG
	RemoteConnectionTypeSocketUnix
	RemoteConnectionTypeUrl
	RemoteConnectionTypeStdioLocal
	RemoteConnectionTypeStdioSSH
	_RemoteConnectionTypeMax
)

func GetAllRemoteConnectionTypes() []RemoteConnectionType {
	types := make([]RemoteConnectionType, 0)

	for i := RemoteConnectionTypeUnspecified + 1; i < _RemoteConnectionTypeMax; i++ {
		types = append(types, RemoteConnectionType(i))
	}

	return types
}

func (t *RemoteConnectionType) Set(v string) (err error) {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "", "none", "unspecified":
		*t = RemoteConnectionTypeUnspecified

	case "native-dotenv-xdg":
		*t = RemoteConnectionTypeNativeDotenvXDG

	case "socket-unix":
		*t = RemoteConnectionTypeSocketUnix

	case "url":
		*t = RemoteConnectionTypeUrl

	case "stdio-local":
		*t = RemoteConnectionTypeStdioLocal

	case "stdio-ssh":
		*t = RemoteConnectionTypeStdioSSH

	default:
		err = errors.ErrorWithStackf("unsupported remote type: %q", v)
		return
	}

	return
}
