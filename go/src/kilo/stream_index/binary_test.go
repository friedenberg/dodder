package stream_index

import (
	"bytes"
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func TestBinaryOne(t1 *testing.T) {
	t := ui.T{T: t1}

	buffer := new(bytes.Buffer)
	coder := binaryEncoder{Sigil: ids.SigilLatest}
	decoder := makeBinary(ids.SigilLatest)
	expected := sku.GetTransactedPool().Get()
	var expectedN int64
	var err error

	{
		t.AssertNoError(
			expected.ObjectId.SetWithIdLike(ids.MustZettelId("one/uno")),
		)
		expected.SetTai(ids.NowTai())
		t.AssertNoError(markl.SetHexBytes(
			markl.HashTypeIdSha256,
			expected.Metadata.GetBlobDigestMutable(),
			[]byte(
				"ed500e315f33358824203cee073893311e0a80d77989dc55c5d86247d95b2403",
			),
		))
		t.AssertNoError(expected.Metadata.Type.Set("da-typ"))
		t.AssertNoError(expected.Metadata.Description.Set("the bez"))
		t.AssertNoError(expected.AddTagPtr(ids.MustTagPtr("tag")))

		// TODO add mother digest field and test
		// {
		// 	id :=
		// "3c5d8b1db2149d279f4d4a6cb9457804aac6944834b62aa283beef99bccd10f0"
		// 	idReader := base64.NewDecoder(
		// 		base64.StdEncoding,
		// 		strings.NewReader(id),
		// 	)

		// 	bufferedReader, repoolBufferedReader := pool.GetBufferedReader(
		// 		idReader,
		// 	)

		// 	defer repoolBufferedReader()

		// 	t.AssertNoError(
		// 		merkle_ids.ReadFromInto(
		// 			bufferedReader,
		// 			expected.Metadata.GetMotherDigestMutable(),
		// 		),
		// 	)
		// }

		{
			config := genesis_configs.Default().Blob

			t.AssertNoError(markl.GeneratePrivateKey(
				nil,
				markl.PurposeRepoPrivateKeyV1,
				markl.FormatIdEd25519Sec,
				config.GetPrivateKeyMutable(),
			))
			t.AssertNoError(expected.FinalizeAndSignOverwrite(config))
		}

		t.Logf("%s", expected)

		expectedN, err = coder.writeFormat(
			buffer,
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
		n, err := decoder.readFormatAndMatchSigil(buffer, &actual)
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
