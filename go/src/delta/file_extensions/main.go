package file_extensions

import "code.linenisgreat.com/dodder/go/src/alfa/interfaces"

var (
	_ interfaces.FileExtensions = TOMLV0{}
	_ interfaces.FileExtensions = TOMLV1{}
)

type Config struct {
	Zettel   string
	Organize string
	Type     string
	Tag      string
	Repo     string
	Config   string
}
