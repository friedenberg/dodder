package age

import (
	"bufio"
	"io"
	"os"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"filippo.io/age"
)

// necessary because the age.Identity interface does not include Stringer, but
// all of the actual identities do implement Stringer
type identity interface {
	age.Identity
	interfaces.Stringer
}

type Identity struct {
	identity identity
	age.Recipient

	path     string
	disabled bool
}

func (i Identity) Unwrap(stanzas []*age.Stanza) (fileKey []byte, err error) {
	if fileKey, err = i.identity.Unwrap(stanzas); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (i *Identity) IsDisabled() bool {
	return i.disabled
}

func (i *Identity) IsEmpty() bool {
	return i.identity == nil
}

func (i *Identity) String() string {
	if i.identity == nil {
		return ""
	} else {
		return i.identity.String()
	}
}

func (i *Identity) MarshalText() (b []byte, err error) {
	b = []byte(i.String())
	return
}

func (i *Identity) UnmarshalText(b []byte) (err error) {
	if err = i.SetFromX25519Identity(string(b)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (i *Identity) SetFromX25519Identity(identity string) (err error) {
	var x *X25519Identity

	if x, err = age.ParseX25519Identity(identity); err != nil {
		err = errors.Wrapf(err, "Identity: %s", identity)
		return
	}

	i.SetX25519Identity(x)

	return
}

func (i *Identity) SetX25519Identity(x *age.X25519Identity) {
	i.disabled = false
	i.identity = x
	i.Recipient = x.Recipient()
}

func (i *Identity) SetFromPath(path string) (err error) {
	var f *os.File

	if f, err = files.Open(path); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, f)

	br := bufio.NewReader(f)
	isEOF := false
	var key string

	for !isEOF {
		var line string
		line, err = br.ReadString('\n')

		if err == io.EOF {
			isEOF = true
			err = nil
		} else if err != nil {
			err = errors.Wrap(err)
			return
		}

		if len(line) > 0 {
			key = strings.TrimSpace(line)
		}
	}

	if err = i.SetFromX25519Identity(key); err != nil {
		err = errors.Wrapf(err, "Key: %q", key)
		return
	}

	return
}

func (i *Identity) Set(path_or_identity string) (err error) {
	switch {
	case path_or_identity == "":

	case path_or_identity == "disabled" || path_or_identity == "none":
		i.disabled = true
		// no-op

	case files.Exists(path_or_identity):
		if err = i.SetFromPath(path_or_identity); err != nil {
			err = errors.Wrap(err)
			return
		}

	case path_or_identity == "generate":
		if err = i.GenerateIfNecessary(); err != nil {
			err = errors.Wrap(err)
			return
		}

	default:
		if err = i.SetFromX25519Identity(path_or_identity); err != nil {
			err = errors.Wrapf(err, "Identity: %q", path_or_identity)
			return
		}
	}

	return
}

func (i *Identity) GenerateIfNecessary() (err error) {
	if i.IsDisabled() || !i.IsEmpty() {
		return
	}

	var x *X25519Identity

	if x, err = age.GenerateX25519Identity(); err != nil {
		err = errors.Wrap(err)
		return
	}

	i.SetX25519Identity(x)

	return
}
