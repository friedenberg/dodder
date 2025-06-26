package age

import (
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"filippo.io/age"
)

type (
	X25519Identity  = age.X25519Identity
	X25519Recipient = age.X25519Recipient
)

type Age struct {
	// Recipients []Recipient `toml:"recipients,omitempty"`
	Identities []*Identity `toml:"identities,omitempty"`
}

func (a *Age) GetBlobEncryption() interfaces.BlobEncryption {
	return a
}

func (a *Age) String() string {
	return fmt.Sprintf("%s", a.Identities)
}

// TODO add support for recipients in addition to identities
func (a *Age) Set(v string) (err error) {
	var identity Identity

	if err = identity.Set(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = a.AddIdentity(identity); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (a *Age) AddIdentity(
	identity Identity,
) (err error) {
	if identity.IsDisabled() || identity.IsEmpty() {
		return
	}

	// a.Recipients = append(a.Recipients, identity)
	a.Identities = append(a.Identities, &identity)

	return
}

func (a *Age) AddIdentityOrGenerateIfNecessary(
	identity Identity,
) (err error) {
	if err = identity.GenerateIfNecessary(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = a.AddIdentity(identity); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func MakeFromIdentity(identity Identity) (a *Age, err error) {
	a = &Age{}
	err = a.AddIdentityOrGenerateIfNecessary(identity)
	return
}

func MakeFromIdentityPathOrString(path_or_identity string) (a *Age, err error) {
	var i Identity

	if err = i.Set(path_or_identity); err != nil {
		err = errors.Wrap(err)
		return
	}

	return MakeFromIdentity(i)
}

func MakeFromIdentityFile(basePath string) (a *Age, err error) {
	var i Identity

	if err = i.SetFromPath(basePath); err != nil {
		err = errors.Wrap(err)
		return
	}

	return MakeFromIdentity(i)
}

func MakeFromIdentityString(contents string) (a *Age, err error) {
	var i Identity

	if err = i.SetFromX25519Identity(contents); err != nil {
		err = errors.Wrap(err)
		return
	}

	return MakeFromIdentity(i)
}

// func (a *Age) AddBech32PivYubikeyEC256(bech string) (err error) {
// 	var r *age.PivYubikeyEC256Recipient

// 	if r, err = age.ParseBech32PivYubikeyEC256Recipient(bech); err != nil {
// 		err = errors.Wrap(err)
// 		return
// 	}

// 	var i *age.PivYubikeyEC256Identity

// 	if i, err = age.ReadPivYubikeyEC256Identity(r); err != nil {
// 		err = errors.Wrap(err)
// 		return
// 	}

// 	a.recipients = append(a.recipients, r)
// 	a.identities = append(a.identities, i)

// 	return
// }

func (a *Age) GetRecipients() []age.Recipient {
	r := make([]age.Recipient, len(a.Identities))

	for i := range r {
		r[i] = a.Identities[i]
	}

	return r
}

func (a *Age) WrapReader(src io.Reader) (out io.ReadCloser, err error) {
	is := make([]age.Identity, len(a.Identities))

	for i := range is {
		is[i] = a.Identities[i]
	}

	switch len(is) {
	case 0:
		// no-op

	default:
		if src, err = age.Decrypt(src, is...); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	out = io.NopCloser(src)

	return
}

func (a *Age) WrapWriter(dst io.Writer) (out io.WriteCloser, err error) {
	r := a.GetRecipients()

	switch len(r) {
	case 0:
		out = writeCloser{dst}
		return

	default:
		if out, err = age.Encrypt(dst, r...); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
