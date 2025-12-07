package doddish

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
)

func ScanExactlyOneSeqWithDotAllowedInIdenfierFromString(value string) (seq Seq, err error) {
	reader, repool := pool.GetStringReader(value)
	defer repool()

	var boxScanner Scanner
	boxScanner.Reset(reader)

	if seq, err = boxScanner.ScanDotAllowedInIdentifiersOrError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
