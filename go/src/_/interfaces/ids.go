package interfaces

type (
	Genre interface {
		GenreGetter
		Stringer
		IsConfig() bool
		IsNone() bool
		IsTag() bool
		IsType() bool
		IsZettel() bool
		GetGenreBitInt() byte
	}

	GenreGetter interface {
		GetGenre() Genre
	}

	ObjectId interface {
		GenreGetter
		Stringer
		IsEmpty() bool
	}

	ExternalObjectId interface {
		ObjectId
		ExternalObjectIdGetter
	}

	ExternalObjectIdGetter interface {
		GetExternalObjectId() ExternalObjectId
	}

	RepoId interface {
		Stringer
		GetRepoIdString() string
	}

	RepoIdGetter interface {
		GetRepoId() RepoId
	}

	Abbreviatable interface {
		Stringer
	}

	FuncExpandString     func(string) (string, error)
	FuncAbbreviateString func(Abbreviatable) (string, error)

	Abbreviator struct {
		Expand     FuncExpandString
		Abbreviate FuncAbbreviateString
	}
)
