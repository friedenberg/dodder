package markl

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func (id Id) Verify(
	mes, sig interfaces.MarklId,
) (err error) {
	var formatPub FormatPub

	{
		var ok bool

		if formatPub, ok = id.format.(FormatPub); !ok {
			err = errors.Errorf(
				"id format does not support pub operation: %T",
				id.format,
			)
			return
		}
	}

	if formatPub.Verify == nil {
		err = errors.Errorf(
			"format does not support verification: %q",
			formatPub.Id,
		)
		return
	}

	defer errors.DeferredRecover(&err)

	if err = formatPub.Verify(id, mes, sig); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
