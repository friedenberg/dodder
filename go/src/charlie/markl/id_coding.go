package markl

import (
	"slices"
)

type IdCodingDataOnly Id

func (id IdCodingDataOnly) MarshalBinary() (bytes []byte, err error) {
	bytes = make([]byte, len(id.data))
	copy(bytes, id.data)
	return
}

func (id *IdCodingDataOnly) UnmarshalBinary(
	bites []byte,
) (err error) {
	id.data = id.data[:0]
	id.data = slices.Grow(id.data, len(bites))
	copy(id.data, bites)
	return
}
