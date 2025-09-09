package markl_age_id

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/zit/go/zit/src/bravo/bech32"
	age_upstream "filippo.io/age"
)

type (
	X25519Identity  = age_upstream.X25519Identity
	X25519Recipient = age_upstream.X25519Recipient
)

type Id struct {
	// Recipients []Recipient `toml:"recipients,omitempty"`
	Identities []*age.Identity `toml:"identities,omitempty"`
}

var _ interfaces.MarklId = Id{}

func (id Id) String() string {
	if len(id.Identities) == 0 {
		return ""
	} else {
		return fmt.Sprintf("%s", id.Identities)
	}
}

func (id *Id) Set(v string) (err error) {
	var identity age.Identity

	if err = identity.Set(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = id.AddIdentity(identity); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id *Id) AddIdentity(
	identity age.Identity,
) (err error) {
	if identity.IsDisabled() || identity.IsEmpty() {
		return
	}

	// a.Recipients = append(a.Recipients, identity)
	id.Identities = append(id.Identities, &identity)

	return
}

func (id *Id) AddIdentityOrGenerateIfNecessary(
	identity age.Identity,
) (err error) {
	if err = identity.GenerateIfNecessary(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = id.AddIdentity(identity); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func MakeFromIdentity(identity age.Identity) (a *Id, err error) {
	a = &Id{}
	err = a.AddIdentityOrGenerateIfNecessary(identity)
	return
}

func MakeFromIdentityPathOrString(path_or_identity string) (a *Id, err error) {
	var i age.Identity

	if err = i.Set(path_or_identity); err != nil {
		err = errors.Wrap(err)
		return
	}

	return MakeFromIdentity(i)
}

func MakeFromIdentityFile(basePath string) (a *Id, err error) {
	var i age.Identity

	if err = i.SetFromPath(basePath); err != nil {
		err = errors.Wrap(err)
		return
	}

	return MakeFromIdentity(i)
}

func MakeFromIdentityString(contents string) (a *Id, err error) {
	var i age.Identity

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

func (id *Id) GetRecipients() []age_upstream.Recipient {
	r := make([]age_upstream.Recipient, len(id.Identities))

	for i := range r {
		r[i] = id.Identities[i]
	}

	return r
}

func (id Id) StringWithFormat() string {
	return id.String()
}

func (id Id) MarshalBinary() ([]byte, error) {
	return nil, errors.Err405MethodNotAllowed.WithStack()
}

func (id Id) GetBytes() []byte {
	switch len(id.Identities) {
	default:
		panic(
			errors.Errorf(
				"unsupported age state for getting bytes: >1 identity",
			),
		)

	case 0:
		return nil

	case 1:
		_, data, err := bech32.Decode(id.Identities[0].String())
		errors.PanicIfError(err)
		return data
	}
}

func (id Id) GetSize() int {
	errors.PanicIfError(errors.Err405MethodNotAllowed)
	return 0
}

func (id Id) GetMarklType() interfaces.MarklType {
	if id.IsNull() {
		return nil
	} else {
		return tipe{}
	}
}

func (id Id) IsNull() bool {
	return len(id.Identities) == 0
}

func (id Id) GetFormat() string {
	return markl.FormatIdMadderPrivateKeyV1
}
