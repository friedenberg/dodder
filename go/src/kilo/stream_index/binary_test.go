package stream_index

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/blob_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func TestBinaryOne(t1 *testing.T) {
	t := ui.T{T: t1}

	b := new(bytes.Buffer)
	coder := binaryEncoder{Sigil: ids.SigilLatest}
	decoder := makeBinary(ids.SigilLatest)
	expected := &sku.Transacted{}
	var expectedN int64
	var err error

	{
		t.AssertNoError(
			expected.ObjectId.SetWithIdLike(ids.MustZettelId("one/uno")),
		)
		expected.SetTai(ids.NowTai())
		t.AssertNoError(expected.Metadata.Blob.Set(
			"ed500e315f33358824203cee073893311e0a80d77989dc55c5d86247d95b2403",
		))
		t.AssertNoError(expected.Metadata.Type.Set("da-typ"))
		t.AssertNoError(expected.Metadata.Description.Set("the bez"))
		t.AssertNoError(expected.AddTagPtr(ids.MustTagPtr("tag")))

		{
			id := "3c5d8b1db2149d279f4d4a6cb9457804aac6944834b62aa283beef99bccd10f0"
			idReader := base64.NewDecoder(
				base64.StdEncoding,
				strings.NewReader(id),
			)

			bufferedReader, repoolBufferedReader := pool.GetBufferedReader(
				idReader,
			)

			defer repoolBufferedReader()

			t.AssertNoError(
				blob_ids.ReadFromInto(
					bufferedReader,
					expected.Metadata.GetMotherDigestMutable(),
				),
			)
		}

		t.AssertNoError(expected.CalculateObjectDigests())

		t.Logf("%s", expected)

		expectedN, err = coder.writeFormat(
			b,
			skuWithSigil{Transacted: expected},
		)
		t.AssertNoError(err)
	}

	actual := skuWithRangeAndSigil{
		skuWithSigil: skuWithSigil{
			Transacted: &sku.Transacted{},
		},
	}

	{
		n, err := decoder.readFormatAndMatchSigil(b, &actual)
		t.AssertNoError(err)
		t.Logf("%s", actual)

		{
			if n != expectedN {
				t.Errorf("expected %d but got %d", expectedN, n)
			}
		}
	}

	if !sku.TransactedEqualer.Equals(expected, actual.Transacted) {
		t.NotEqual(expected, actual)
	}
}
