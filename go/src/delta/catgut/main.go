package catgut

import "code.linenisgreat.com/dodder/go/src/alfa/interfaces"

type (
	StringFormatReader[T any] interface {
		ReadStringFormat(*RingBuffer, T) (int64, error)
	}

	StringFormatWriter[T any] interface {
		interfaces.StringEncoderTo[T]
	}

	StringFormatReadWriter[T any] interface {
		StringFormatReader[T]
		StringFormatWriter[T]
	}
)
