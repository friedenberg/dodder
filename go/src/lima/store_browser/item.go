package store_browser

import (
	"slices"
	"strings"

	"code.linenisgreat.com/chrest/go/src/charlie/browser_items"
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/descriptions"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type Item struct {
	browser_items.Item
}

func (i *Item) GetExternalObjectId() sku.ExternalObjectId {
	return ids.MakeExternalObjectId(genres.Zettel, i.String())
}

func (i *Item) GetGenre() interfaces.Genre {
	return genres.Zettel
}

func (i *Item) String() string {
	return i.GetKey()
}

func (i *Item) GetKey() string {
	return i.Id.String()
}

// TODO replace with external id
func (i *Item) GetObjectId() *ids.ObjectId {
	var oid ids.ObjectId
	errors.PanicIfError(oid.SetLeft(i.GetKey()))
	// errors.PanicIfError(oid.SetRepoId("browser"))
	return &oid
}

func (i *Item) GetType() (t ids.Type, err error) {
	if err = t.Set("browser-" + i.Id.Type); err != nil {
		err = errors.Wrap(err)
		return t, err
	}

	return t, err
}

// TODO move below to !toml-bookmark type
func (i Item) GetUrlPathTag() (e ids.Tag, err error) {
	ur := i.Url.Url()
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

func (i Item) GetTai() (t ids.Tai, err error) {
	if i.Date == "" {
		return t, err
	}

	if err = t.SetFromRFC3339(i.Date); err != nil {
		err = errors.Wrap(err)
		return t, err
	}

	return t, err
}

var errEmptyUrl = errors.New("empty url")

func (i Item) GetDescription() (b descriptions.Description, err error) {
	if err = b.Set(i.Title); err != nil {
		err = errors.Wrap(err)
		return b, err
	}

	return b, err
}

func (i *Item) WriteToExternal(e *sku.Transacted) (err error) {
	if !i.Id.IsEmpty() {
		if err = e.ExternalObjectId.Set(i.Id.String()); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	e.GetMetadataMutable().GetTypePtr().Set("!toml-bookmark")

	m := &e.Metadata

	if m.Tai, err = i.GetTai(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if e.ExternalType, err = i.GetType(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if e.GetMetadata().GetDescription().IsEmpty() {
		if err = e.GetMetadataMutable().GetDescriptionMutable().Set(i.Title); err != nil {
			err = errors.Wrap(err)
			return err
		}
	} else if i.Title != "" && e.GetMetadata().GetDescription().String() != i.Title {
		e.GetMetadataMutable().GetFieldsMutable().Append(
			string_format_writer.Field{
				Key:       "title",
				Value:     i.Title,
				ColorType: string_format_writer.ColorTypeUserData,
			},
		)
	}

	e.GetMetadataMutable().GetFieldsMutable().Append(
		string_format_writer.Field{
			Key:       "url",
			Value:     i.Url.String(),
			ColorType: string_format_writer.ColorTypeUserData,
		},
	)

	// TODO move to !toml-bookmark type
	var t ids.Tag

	if t, err = i.GetUrlPathTag(); err == nil {
		if err = m.AddTagPtr(&t); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	err = nil

	return err
}

func (i *Item) ReadFromExternal(e *sku.Transacted) (err error) {
	if err = i.Id.Set(
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

			if err = i.Id.Set(e.ExternalObjectId.String()); err != nil {
				err = errors.Wrap(err)
				return err
			}

		case "", "title":
			i.Title = field.Value

		case "url":
			if err = i.Url.Set(field.Value); err != nil {
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
