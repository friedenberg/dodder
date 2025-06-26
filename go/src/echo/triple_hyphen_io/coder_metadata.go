package triple_hyphen_io

import (
	"bufio"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/echo/format"
)

type TypedMetadataCoder[O any] struct{}

func (TypedMetadataCoder[O]) DecodeFrom(
	subject *TypedStruct[O],
	reader *bufio.Reader,
) (n int64, err error) {
	bufferedReader := bufio.NewReader(reader)

	// TODO scan for type directly
	if n, err = format.ReadLines(
		bufferedReader,
		ohio.MakeLineReaderRepeat(
			ohio.MakeLineReaderKeyValues(
				map[string]interfaces.FuncSetString{
					"!": subject.Type.Set,
				},
			),
		),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (TypedMetadataCoder[O]) EncodeTo(
	subject *TypedStruct[O],
	writer *bufio.Writer,
) (n int64, err error) {
	var n1 int
	n1, err = fmt.Fprintf(writer, "! %s\n", subject.Type.StringSansOp())
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
