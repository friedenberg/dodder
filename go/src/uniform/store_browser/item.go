package store_browser

import (
	"slices"
	"strings"

	"code.linenisgreat.com/chrest/go/src/charlie/browser_items"
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/descriptions"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
)

type Item struct {
	browser_items.Item
}

func (item *Item) GetExternalObjectId() sku.ExternalObjectId {
	return ids.MakeExternalObjectId(genres.Zettel, item.String())
}

func (item *Item) GetGenre() interfaces.Genre {
	return genres.Zettel
}

func (item *Item) String() string {
	return item.GetKey()
}

func (item *Item) GetKey() string {
	return item.Id.String()
}

// TODO replace with external id
func (item *Item) GetObjectId() *ids.ObjectId {
	var oid ids.ObjectId
	errors.PanicIfError(oid.SetLeft(item.GetKey()))
	// errors.PanicIfError(oid.SetRepoId("browser"))
	return &oid
}

func (item *Item) GetType() (t ids.Type, err error) {
	if err = t.Set("browser-" + item.Id.Type); err != nil {
		err = errors.Wrap(err)
		return t, err
	}

	return t, err
}

// TODO move below to !toml-bookmark type
func (item Item) GetUrlPathTag() (e ids.Tag, err error) {
	ur := item.Url.Url()
	els := strings.Split(ur.Hostname(), ".")
	slices.Reverse(els)

	if els[0] == "www" {
		els = els[1:]
	}

	host := strings.Join(els, "-")

	if len(host) == 0 {
		err = errors.ErrorWithStackf("empty host: %q", els)
		return e, err
	}

	if err = e.Set("%zz-site-" + host); err != nil {
		err = errors.Wrap(err)
		return e, err
	}

	return e, err
}

func (item Item) GetTai() (t ids.Tai, err error) {
	if item.Date == "" {
		return t, err
	}

	if err = t.SetFromRFC3339(item.Date); err != nil {
		err = errors.Wrap(err)
		return t, err
	}

	return t, err
}

var errEmptyUrl = errors.New("empty url")

func (item Item) GetDescription() (b descriptions.Description, err error) {
	if err = b.Set(item.Title); err != nil {
		err = errors.Wrap(err)
		return b, err
	}

	return b, err
}

func (item *Item) WriteToExternal(object *sku.Transacted) (err error) {
	if !item.Id.IsEmpty() {
		if err = object.ExternalObjectId.Set(item.Id.String()); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	object.GetMetadataMutable().GetTypeMutable().Set("!toml-bookmark")

	metadata := object.GetMetadataMutable()

	var tai ids.Tai

	if tai, err = item.GetTai(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	metadata.GetTaiMutable().ResetWith(tai)

	if object.ExternalType, err = item.GetType(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if object.GetMetadata().GetDescription().IsEmpty() {
		if err = object.GetMetadataMutable().GetDescriptionMutable().Set(item.Title); err != nil {
			err = errors.Wrap(err)
			return err
		}
	} else if item.Title != "" && object.GetMetadata().GetDescription().String() != item.Title {
		object.GetMetadataMutable().GetFieldsMutable().Append(
			string_format_writer.Field{
				Key:       "title",
				Value:     item.Title,
				ColorType: string_format_writer.ColorTypeUserData,
			},
		)
	}

	object.GetMetadataMutable().GetFieldsMutable().Append(
		string_format_writer.Field{
			Key:       "url",
			Value:     item.Url.String(),
			ColorType: string_format_writer.ColorTypeUserData,
		},
	)

	// TODO move to !toml-bookmark type
	var t ids.Tag

	if t, err = item.GetUrlPathTag(); err == nil {
		if err = metadata.AddTagPtr(&t); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	err = nil

	return err
}

func (item *Item) ReadFromExternal(e *sku.Transacted) (err error) {
	if err = item.Id.Set(
		strings.TrimSuffix(
			e.ExternalObjectId.String(),
			"/",
		),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	for field := range e.GetMetadata().GetFields() {
		switch field.Key {
		case "id":
			if field.Value == "" {
				continue
			}

			if err = item.Id.Set(e.ExternalObjectId.String()); err != nil {
				err = errors.Wrap(err)
				return err
			}

		case "", "title":
			item.Title = field.Value

		case "url":
			if err = item.Url.Set(field.Value); err != nil {
				err = errors.Wrap(err)
				return err
			}

		default:
			err = errors.ErrorWithStackf(
				"unsupported field type: %q=%q. Fields: %#v",
				field.Key,
				field.Value,
				e.GetMetadata().GetFields(),
			)

			return err
		}
	}

	// err = todo.Implement()
	return err
}
