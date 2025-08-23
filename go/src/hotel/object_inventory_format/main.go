package object_inventory_format

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
)

type (
	FormatterContext interface {
		object_metadata.PersistentFormatterContext
		GetObjectId() *ids.ObjectId
	}

	ParserContext interface {
		object_metadata.PersistentParserContext
		SetObjectIdLike(interfaces.ObjectId) error
	}

	Options struct {
		Tai           bool
		ExcludeMutter bool
		Verzeichnisse bool
		PrintFinalSha bool
	}

	nopFormatterContext struct {
		object_metadata.PersistentFormatterContext
	}
)

func (nopFormatterContext) GetObjectId() *ids.ObjectId {
	return nil
}

func (o Options) SansVerzeichnisse() Options {
	o.Verzeichnisse = false
	return o
}
