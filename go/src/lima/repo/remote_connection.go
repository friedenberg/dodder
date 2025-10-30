package repo

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

//go:generate stringer -type=RemoteConnection
type RemoteConnection int

const (
	RemoteConnectionUnspecified = RemoteConnection(iota)
	RemoteConnectionNative
	RemoteConnectionNativeLocalOverridePath
	RemoteConnectionSocketUnix
	RemoteConnectionUrl
	RemoteConnectionStdioLocal
	RemoteConnectionStdioSSH
	_RemoteConnectionMax
)

func GetAllRemoteConnectionTypes() []RemoteConnection {
	types := make([]RemoteConnection, 0)

	for i := RemoteConnectionUnspecified + 1; i < _RemoteConnectionMax; i++ {
		types = append(types, RemoteConnection(i))
	}

	return types
}

func (tipe *RemoteConnection) GetCLICompletion() map[string]string {
	return map[string]string{
		"native": "",
		// TODO rename
		"native-dotenv-xdg": "",
		"none":              "",
		"socket-unix":       "",
		"stdio-local":       "",
		"stdio-ssh":         "",
		"unspecified":       "",
		"url":               "",
	}
}

func (tipe *RemoteConnection) Set(value string) (err error) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "none", "unspecified":
		*tipe = RemoteConnectionUnspecified

	case "native":
		*tipe = RemoteConnectionNative

		// TODO rename
	case "native-dotenv-xdg":
		*tipe = RemoteConnectionNativeLocalOverridePath

	case "socket-unix":
		*tipe = RemoteConnectionSocketUnix

	case "url":
		*tipe = RemoteConnectionUrl

	case "stdio-local":
		*tipe = RemoteConnectionStdioLocal

	case "stdio-ssh":
		*tipe = RemoteConnectionStdioSSH

	default:
		err = errors.ErrorWithStackf("unsupported remote type: %q", value)
		return err
	}

	return err
}
