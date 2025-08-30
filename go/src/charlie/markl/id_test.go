package markl

import (
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func TestIdNullAndEqual(t1 *testing.T) {
	t := ui.MakeTestContext(t1)

	hashType := HashTypeSha256

	{
		t.AssertError(MakeErrIsNull(hashType.null, ""))
		t.AssertNoError(MakeErrIsNotNull(hashType.null))
		t.AssertNoError(
			MakeErrLength(hashType.GetSize(), len(hashType.null.GetBytes())),
		)
	}

	var idZero Id
	hash := hashType.Get()

	{
		idNull, _ := hash.GetBlobId()

		t.AssertNoError(MakeErrIsNotNull(idZero))
		t.AssertNoError(MakeErrIsNotNull(idNull))
		t.AssertError(MakeErrIsNull(idZero, ""))
		t.AssertError(MakeErrIsNull(idNull, ""))
		t.AssertNoError(MakeErrNotEqual(idZero, idNull))
		t.AssertNoError(MakeErrNotEqual(idNull, idZero))
	}

	{
		idNull, _ := hashType.GetBlobIdForString("")

		t.AssertNoError(MakeErrIsNotNull(idZero))
		t.AssertNoError(MakeErrIsNotNull(idNull))
		t.AssertError(MakeErrIsNull(idZero, ""))
		t.AssertError(MakeErrIsNull(idNull, ""))
		t.AssertNoError(MakeErrNotEqual(idZero, idNull))
		t.AssertNoError(MakeErrNotEqual(idNull, idZero))
	}

	{
		idNull, _ := hashType.GetBlobIdForHexString(
			"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		)

		t.AssertNoError(MakeErrIsNotNull(idZero))
		t.AssertNoError(MakeErrIsNotNull(idNull))
		t.AssertError(MakeErrIsNull(idZero, ""))
		t.AssertError(MakeErrIsNull(idNull, ""))
		t.AssertNoError(MakeErrNotEqual(idZero, idNull))
		t.AssertNoError(MakeErrNotEqual(idNull, idZero))
	}

	{
		idNonZero, _ := hashType.GetBlobIdForString("nonZero")
		t.AssertNoError(MakeErrIsNull(idNonZero, ""))
		t.AssertError(MakeErrIsNotNull(idNonZero))
		t.AssertError(MakeErrNotEqual(idNonZero, idZero))
		t.AssertError(MakeErrNotEqual(idZero, idNonZero))
	}
}
