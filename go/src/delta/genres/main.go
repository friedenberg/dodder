package genres

import (
	"fmt"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
)

type Genre byte

// Do not change this order, various serialization formats rely on the
// underlying integer values.
const (
	None = Genre(iota)
	Blob
	Type
	_ // Bezeichnung
	Tag
	_ // Hinweis
	_ // Transaktion
	Zettel
	Config
	_ // Kennung
	InventoryList
	_ // AkteTyp
	Repo

	maxGenre = Repo
)

const (
	unknown = byte(iota)
	blob    = byte(1 << iota)
	tipe
	tag
	zettel
	config
	inventory_list
	repo
)

func All() (out quiter.Slice[Genre]) {
	out = make([]Genre, 0, maxGenre-1)

	for i := None + 1; i <= maxGenre; i++ {
		g := Genre(i)

		switch g {
		default:
			continue

		case Type, Tag, Zettel, Config, Repo, InventoryList, Blob:
		}

		out = append(out, g)
	}

	return out
}

func Must(g interfaces.GenreGetter) Genre {
	return g.GetGenre().(Genre)
}

func Make(g interfaces.Genre) Genre {
	return Must(g)
}

func MakeOrUnknown(v string) (g Genre) {
	if err := g.Set(v); err != nil {
		g = None
	}

	return g
}

func (genre Genre) GetGenre() interfaces.Genre {
	return genre
}

func (genre Genre) GetGenreBitInt() byte {
	switch genre {
	default:
		panic(fmt.Sprintf("genre does not define bit int: %s", genre))
	case InventoryList:
		return inventory_list
	case Blob:
		return blob
	case Zettel:
		return zettel
	case Tag:
		return tag
	case Repo:
		return repo
	case Type:
		return tipe
	case Config:
		return config
	}
}

func (genre Genre) EqualsAny(b any) bool {
	return values.Equals(genre, b)
}

func (genre Genre) Equals(b Genre) bool {
	return genre == b
}

func (genre Genre) EqualsGenre(b interfaces.GenreGetter) bool {
	return genre.GetGenreString() == b.GetGenre().GetGenreString()
}

func (genre Genre) AssertGenre(b interfaces.GenreGetter) (err error) {
	if genre.GetGenreString() != b.GetGenre().GetGenreString() {
		err = MakeErrUnsupportedGenre(b)
		return err
	}

	return err
}

func (genre Genre) GetGenreString() string {
	return genre.String()
}

func (genre Genre) GetGenreStringVersioned(sv interfaces.StoreVersion) string {
	if store_version.LessOrEqual(sv, store_version.V6) {
		return genre.stringOld()
	} else {
		return genre.String()
	}
}

func (genre Genre) GetGenreStringPlural(sv interfaces.StoreVersion) string {
	if store_version.LessOrEqual(sv, store_version.V6) {
		return genre.getGenreStringPluralOld()
	} else {
		return genre.getGenreStringPluralNew()
	}
}

func (genre Genre) getGenreStringPluralNew() string {
	switch genre {
	case Blob:
		return "blobs"

	case Type:
		return "types"

	case Tag:
		return "tags"

	case Zettel:
		return "zettels"

	case InventoryList:
		return "inventory_lists"

	case Repo:
		return "repos"

	default:
		panic(fmt.Sprintf("undeclared plural for genre: %q", genre))
	}
}

func (genre Genre) getGenreStringPluralOld() string {
	switch genre {
	case Blob:
		return "Akten"

	case Type:
		return "Typen"

	case Tag:
		return "Etiketten"

	case Zettel:
		return "Zettelen"

	case InventoryList:
		return "Bestandsaufnahmen"

	case Repo:
		return "Kisten"

	default:
		return genre.String()
	}
}

func (genre Genre) String() string {
	switch genre {
	case Blob:
		return "Blob"

	case Type:
		return "Type"

	case Tag:
		return "Tag"

	case Zettel:
		return "Zettel"

	case Config:
		return "Config"

	case InventoryList:
		return "InventoryList"

	case Repo:
		return "Repo"

	case None:
		return "none"

	default:
		return fmt.Sprintf("Unknown(%#v)", genre)
	}
}

func (genre Genre) stringOld() string {
	switch genre {
	case Blob:
		return "Akte"

	case Type:
		return "Typ"

	case Tag:
		return "Etikett"

	case Zettel:
		return "Zettel"

	case Config:
		return "Konfig"

	case InventoryList:
		return "Bestandsaufnahme"

	case Repo:
		return "Kasten"

	case None:
		return "none"

	default:
		return fmt.Sprintf("Unknown(%#v)", genre)
	}
}

func hasPrefixOrEquals(v, p string) bool {
	return strings.HasPrefix(v, p) || strings.EqualFold(v, p)
}

func (genre *Genre) Set(v string) (err error) {
	v = strings.TrimSpace(v)

	switch {
	case strings.EqualFold(v, "blob"):
		fallthrough
	case strings.EqualFold(v, "akte"):
		*genre = Blob

	case hasPrefixOrEquals("type", v):
		fallthrough
	case hasPrefixOrEquals("typ", v):
		*genre = Type

	case strings.EqualFold(v, "aktetyp"):
		*genre = Type

	case hasPrefixOrEquals("tag", v):
		fallthrough
	case hasPrefixOrEquals("etikett", v):
		*genre = Tag

	case hasPrefixOrEquals("zettel", v):
		*genre = Zettel

	case strings.EqualFold("config", v):
		fallthrough
	case strings.EqualFold("konfig", v):
		*genre = Config

	case hasPrefixOrEquals("inventorylist", v):
		fallthrough
	case hasPrefixOrEquals("inventory_list", v):
		fallthrough
	case hasPrefixOrEquals("inventory-list", v):
		fallthrough
	case hasPrefixOrEquals("bestandsaufnahme", v):
		*genre = InventoryList

	case hasPrefixOrEquals("repo", v):
		fallthrough
	case hasPrefixOrEquals("kasten", v):
		*genre = Repo

	default:
		err = errors.Wrap(MakeErrUnrecognizedGenre(v))
		return err
	}

	return err
}

func (genre *Genre) Reset() {
	*genre = None
}

func (genre *Genre) ReadFrom(r io.Reader) (n int64, err error) {
	*genre = None

	var b [1]byte

	var n1 int
	n1, err = ohio.ReadAllOrDieTrying(r, b[:])
	n += int64(n1)

	if err != nil {
		return n, err
	}

	*genre = Genre(b[0])

	return n, err
}

func (genre *Genre) WriteTo(w io.Writer) (n int64, err error) {
	b := [1]byte{byte(*genre)}

	var n1 int
	n1, err = ohio.WriteAllOrDieTrying(w, b[:])
	n += int64(n1)

	if err != nil {
		return n, err
	}

	return n, err
}

func (genre Genre) MarshalBinary() (b []byte, err error) {
	b = []byte{genre.Byte()}
	return b, err
}

func (genre *Genre) UnmarshalBinary(b []byte) (err error) {
	if len(b) != 1 {
		err = errors.ErrorWithStackf("expected exactly one byte but got %q", b)
		return err
	}

	*genre = Genre(b[0])

	return err
}

func (genre Genre) Byte() byte {
	return byte(genre)
}

func (genre Genre) ReadByte() (byte, error) {
	return genre.Byte(), nil
}
