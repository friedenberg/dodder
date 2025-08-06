package file_extensions

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

func GetFileExtensionForGenre(
	config interfaces.FileExtensions,
	getter interfaces.GenreGetter,
) string {
	genre := genres.Must(getter)

	switch genre {
	case genres.Zettel:
		return config.GetFileExtensionZettel()

	case genres.Type:
		return config.GetFileExtensionType()

	case genres.Tag:
		return config.GetFileExtensionTag()

	case genres.Repo:
		return config.GetFileExtensionRepo()

	case genres.Config:
		return config.GetFileExtensionConfig()

	default:
		return ""
	}
}
