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

func (tipe *RemoteConnectionType) GetCLICompletion() map[string]string {
	return map[string]string{
		"native-dotenv-xdg": "",
		"none":              "",
		"socket-unix":       "",
		"stdio-local":       "",
		"stdio-ssh":         "",
		"unspecified":       "",
		"url":               "",
	}
}

func (tipe *RemoteConnectionType) Set(value string) (err error) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "none", "unspecified":
		*tipe = RemoteConnectionTypeUnspecified

	case "native-dotenv-xdg":
		*tipe = RemoteConnectionTypeNativeDotenvXDG

	case "socket-unix":
		*tipe = RemoteConnectionTypeSocketUnix

	case "url":
		*tipe = RemoteConnectionTypeUrl

	case "stdio-local":
		*tipe = RemoteConnectionTypeStdioLocal

	case "stdio-ssh":
		*tipe = RemoteConnectionTypeStdioSSH

	default:
		err = errors.ErrorWithStackf("unsupported remote type: %q", value)
		return err
	}

	return err
}
