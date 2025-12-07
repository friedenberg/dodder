package genres

import (
	"fmt"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/ohio"
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

var _ interfaces.Genre = None

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

// TODO convert to seq
func All() (out collections_slice.Slice[Genre]) {
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

func Must(genre interfaces.GenreGetter) Genre {
	return genre.GetGenre().(Genre)
}

func Make(genre interfaces.Genre) Genre {
	return Must(genre)
}

func MakeOrUnknown(value string) (genre Genre) {
	if err := genre.Set(value); err != nil {
		genre = None
	}

	return genre
}

func (genre Genre) GetGenre() interfaces.Genre {
	return genre
}

func (genre Genre) IsConfig() bool {
	return genre == Config
}

func (genre Genre) IsZettel() bool {
	return genre == Zettel
}

func (genre Genre) IsTag() bool {
	return genre == Tag
}

func (genre Genre) IsType() bool {
	return genre == Type
}

func (genre Genre) IsNone() bool {
	return genre == None
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

func (genre Genre) Equals(b Genre) bool {
	return genre == b
}

func (genre Genre) AssertGenre(b interfaces.GenreGetter) (err error) {
	if genre.String() != b.GetGenre().String() {
		err = MakeErrUnsupportedGenre(b)
		return err
	}

	return err
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
