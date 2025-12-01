package interfaces

type (
	ObjectId interface {
		GenreGetter
		Stringer
		Parts() [3]string
		IsEmpty() bool
	}

	RepoId interface {
		Stringer
		EqualsRepoId(RepoIdGetter) bool
		GetRepoIdString() string
	}

	RepoIdGetter interface {
		GetRepoId() RepoId
	}

	Genre interface {
		GenreGetter
		Stringer
		EqualsGenre(GenreGetter) bool
		GetGenreBitInt() byte
		GetGenreString() string
		GetGenreStringVersioned(StoreVersion) string
		GetGenreStringPlural(StoreVersion) string
	}

	GenreGetter interface {
		GetGenre() Genre
	}
)
