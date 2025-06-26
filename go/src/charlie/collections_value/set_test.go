package collections_value

import (
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
)

func TestSet(t1 *testing.T) {
	t := ui.T{T: t1}

	{
		vals := makeStringValues(
			"1 one",
			"2 two",
			"3 three",
		)

		sut := MakeValueSet[values.String](
			nil,
			vals...,
		)

		assertSet(t, sut, vals)
	}

	{
		vals := makeStringValues(
			"1 one",
			"2 two",
			"3 three",
		)

		sut := MakeMutableValueSet[values.String](
			nil,
			vals...,
		)

		assertSet(t, sut, vals)
	}

	{
		vals := makeStringValues(
			"1 one",
			"2 two",
			"3 three",
		)

		sut := MakeValueSet[values.String](
			nil,
			vals...,
		)

		assertSet(t, sut, vals)
	}

	{
		vals := makeStringValues(
			"1 one",
			"2 two",
			"3 three",
		)

		sut := MakeMutableValueSet[values.String](
			nil,
			vals...,
		)

		assertSet(t, sut, vals)
	}
}
