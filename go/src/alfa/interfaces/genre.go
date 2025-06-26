package interfaces

type ObjectId interface {
	GenreGetter
	Stringer
	Parts() [3]string
}

type RepoId interface {
	Stringer
	EqualsRepoId(RepoIdGetter) bool
	GetRepoIdString() string
}

type RepoIdGetter interface {
	GetRepoId() RepoId
}

type Genre interface {
	GenreGetter
	Stringer
	EqualsGenre(GenreGetter) bool
	GetGenreBitInt() byte
	GetGenreString() string
	GetGenreStringVersioned(StoreVersion) string
	GetGenreStringPlural(StoreVersion) string
}

type GenreGetter interface {
	GetGenre() Genre
}
