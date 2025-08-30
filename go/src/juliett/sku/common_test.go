package sku

import (
	"crypto/sha256"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_ptr"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/echo/descriptions"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

type inlineTypChecker struct {
	answer bool
}

func (t inlineTypChecker) IsInlineTyp(k ids.Type) bool {
	return t.answer
}

func makeTagSet(t *ui.TestContext, vs ...string) (es ids.TagSet) {
	var err error

	if es, err = collections_ptr.MakeValueSetString[ids.Tag](nil, vs...); err != nil {
		t.Fatalf("%s", err)
	}

	return
}

func makeBlobExt(t *ui.TestContext, v string) (es ids.Type) {
	if err := es.Set(v); err != nil {
		t.Fatalf("%s", err)
	}

	return
}

func readFormat(
	t1 *ui.TestContext,
	format object_metadata.TextFormat,
	contents string,
) (metadata *object_metadata.Metadata) {
	var object Transacted

	t := t1

	reader, repool := pool.GetStringReader(contents)
	defer repool()
	n, err := format.ParseMetadata(
		reader,
		&object,
	)
	t.AssertNoError(err)

	if n != int64(len(contents)) {
		t.Fatalf("expected to read %d but only read %d", len(contents), n)
	}

	metadata = object.GetMetadata()

	return
}

func TestMakeTags(t1 *testing.T) {
	t := ui.MakeTestContext(t1)

	vs := []string{
		"tag1",
		"tag2",
		"tag3",
	}

	var sut ids.TagSet
	var err error

	if sut, err = ids.MakeTagSetStrings(vs...); err != nil {
		t.Fatalf("%s", err)
	}

	if sut.Len() != 3 {
		t.Fatalf("expected len 3 but got %d", sut.Len())
	}

	{
		ac := sut.Len()

		if ac != 3 {
			t.Fatalf("expected len 3 but got %d", ac)
		}
	}

	sut2 := sut.CloneSetLike()

	if sut2.Len() != 3 {
		t.Fatalf("expected len 3 but got %d", sut2.Len())
	}

	{
		ac := quiter.SortedStrings[ids.Tag](sut)

		if !reflect.DeepEqual(ac, vs) {
			t.Fatalf("expected %q but got %q", vs, ac)
		}
	}

	{
		ex := "tag1, tag2, tag3"
		ac := quiter.StringCommaSeparated[ids.Tag](sut)

		if ac != ex {
			t.Fatalf("expected %q but got %q", ex, ac)
		}
	}

	{
		ex := "tag1, tag2, tag3"
		ac := quiter.StringCommaSeparated[ids.Tag](
			sut.CloneSetLike(),
		)

		if ac != ex {
			t.Fatalf("expected %q but got %q", ex, ac)
		}
	}
}

func TestEqualitySelf(t1 *testing.T) {
	t := ui.MakeTestContext(t1)

	text := &object_metadata.Metadata{
		Description: descriptions.Make("the title"),
		Type:        makeBlobExt(t, "text"),
	}

	text.SetTags(makeTagSet(t,
		"tag1",
		"tag2",
		"tag3",
	))

	if !text.Equals(text) {
		t.Fatalf("expected %v to equal itself", text)
	}
}

func TestEqualityNotSelf(t1 *testing.T) {
	t := ui.MakeTestContext(t1)

	text := object_metadata.Metadata{
		Description: descriptions.Make("the title"),
		Type:        makeBlobExt(t, "text"),
	}

	text.SetTags(makeTagSet(t,
		"tag1",
		"tag2",
		"tag3",
	))

	text1 := &object_metadata.Metadata{
		Description: descriptions.Make("the title"),
		Type:        makeBlobExt(t, "text"),
	}

	text1.SetTags(makeTagSet(t,
		"tag1",
		"tag2",
		"tag3",
	))

	if !text.Equals(text1) {
		t.Fatalf("expected %v to equal %v", text, text1)
	}
}

func makeTestTextFormat(
	envDir env_dir.Env,
	blobStore interfaces.BlobStore,
) object_metadata.TextFormat {
	return object_metadata.MakeTextFormat(
		object_metadata.Dependencies{
			EnvDir:         envDir,
			BlobStore:      blobStore,
			BlobDigestType: markl.HRPObjectBlobDigestSha256V1,
		},
	)
}

func TestReadWithoutBlob(t1 *testing.T) {
	t := ui.MakeTestContext(t1)
	envRepo := env_repo.MakeTesting(t, nil)

	actual := readFormat(
		t,
		makeTestTextFormat(envRepo, envRepo.GetDefaultBlobStore()),
		`---
# the title
- tag1
- tag2
- tag3
! md
---
`,
	)

	expected := &object_metadata.Metadata{
		Description: descriptions.Make("the title"),
		Type:        makeBlobExt(t, "md"),
	}

	expected.SetTags(makeTagSet(t,
		"tag1",
		"tag2",
		"tag3",
	))

	if !actual.Equals(expected) {
		t.Fatalf(
			"zettel:\nexpected: %s\n  actual: %s",
			StringMetadataSansTaiMerkle2(expected),
			StringMetadataSansTaiMerkle2(actual),
		)
	}

	if !actual.GetBlobDigest().IsNull() {
		t.Fatalf("blob:\nexpected empty but got %q", actual.GetBlobDigest())
	}
}

func TestReadWithoutBlobWithMultilineDescription(t1 *testing.T) {
	t := ui.MakeTestContext(t1)

	envRepo := env_repo.MakeTesting(t, nil)

	actual := readFormat(
		t,
		makeTestTextFormat(envRepo, envRepo.GetDefaultBlobStore()),
		`---
# the title
# continues
- tag1
- tag2
- tag3
! md
---
`,
	)

	expected := &object_metadata.Metadata{
		Description: descriptions.Make("the title continues"),
		Type:        makeBlobExt(t, "md"),
	}

	expected.SetTags(makeTagSet(t,
		"tag1",
		"tag2",
		"tag3",
	))

	if !actual.Equals(expected) {
		t.Fatalf("zettel:\nexpected: %#v\n  actual: %#v", expected, actual)
	}

	if !actual.GetBlobDigest().IsNull() {
		t.Fatalf("blob:\nexpected empty but got %q", actual.GetBlobDigest())
	}
}

func TestReadWithBlob(t1 *testing.T) {
	t := ui.MakeTestContext(t1)

	envRepo := env_repo.MakeTesting(t, nil)

	actual := readFormat(
		t,
		makeTestTextFormat(envRepo, envRepo.GetDefaultBlobStore()),
		`---
# the title
- tag1
- tag2
- tag3
! md
---

the body`,
	)

	expectedSha, _ := markl.HashTypeSha256.GetBlobIdForHexString(
		"fa8242e99f48966ca514092b4233b446851f42b57ad5031bf133e1dd76787f3e",
	)

	expected := &object_metadata.Metadata{
		Description: descriptions.Make("the title"),
		Type:        makeBlobExt(t, "md"),
	}

	expected.GetBlobDigestMutable().ResetWithMerkleId(expectedSha)

	expected.SetTags(makeTagSet(t,
		"tag1",
		"tag2",
		"tag3",
	))

	if !actual.Equals(expected) {
		t.Fatalf("zettel:\nexpected: %#v\n  actual: %#v", expected, actual)
	}
}

type noopCloser struct {
	*strings.Reader
}

func (c noopCloser) Close() error {
	return nil
}

type blobReaderFactory struct {
	t     *ui.TestContext
	blobs map[string]string
}

func (arf blobReaderFactory) BlobReader(
	digest interfaces.BlobId,
) (readCloser interfaces.ReadCloseBlobIdGetter, err error) {
	var v string
	var ok bool

	if v, ok = arf.blobs[digest.String()]; !ok {
		arf.t.Fatalf("request for non-existent blob: %s", digest)
	}

	readCloser = markl.MakeNopReadCloser(
		markl.HashTypeSha256.Get(),
		io.NopCloser(strings.NewReader(v)),
	)

	return
}

func writeFormat(
	t *ui.TestContext,
	m *object_metadata.Metadata,
	f object_metadata.TextFormatter,
	includeBlob bool,
	blobBody string,
	options object_metadata.TextFormatterOptions,
) (out string) {
	hash := sha256.New()
	reader, repool := pool.GetStringReader(blobBody)
	defer repool()
	_, err := io.Copy(hash, reader)
	if err != nil {
		t.Fatalf("%s", err)
	}

	blobDigestRaw := fmt.Sprintf("%x", hash.Sum(nil))
	var blobDigest markl.Id

	if err := blobDigest.SetMaybeSha256(blobDigestRaw); err != nil {
		t.Fatalf("%s", err)
	}

	m.GetBlobDigestMutable().ResetWithMerkleId(&blobDigest)

	sb := &strings.Builder{}

	if _, err := f.FormatMetadata(
		sb,
		object_metadata.TextFormatterContext{
			PersistentFormatterContext: m,
			TextFormatterOptions:       options,
		},
	); err != nil {
		t.Errorf("%s", err)
	}

	out = sb.String()

	return
}

func TestWriteWithoutBlob(t1 *testing.T) {
	t := ui.MakeTestContext(t1)

	z := &object_metadata.Metadata{
		Description: descriptions.Make("the title"),
		Type:        makeBlobExt(t, "md"),
	}

	z.SetTags(makeTagSet(t,
		"tag1",
		"tag2",
		"tag3",
	))

	envRepo := env_repo.MakeTesting(
		t,
		map[string]string{
			"fa8242e99f48966ca514092b4233b446851f42b57ad5031bf133e1dd76787f3e": "the body",
		},
	)

	format := object_metadata.MakeTextFormatterMetadataOnly(
		object_metadata.Dependencies{
			BlobStore:      envRepo.GetDefaultBlobStore(),
			BlobDigestType: markl.HRPObjectBlobDigestSha256V1,
		},
	)

	actual := writeFormat(
		t,
		z,
		format,
		false,
		"the body",
		object_metadata.TextFormatterOptions{},
	)

	expected := `---
# the title
- tag1
- tag2
- tag3
! fa8242e99f48966ca514092b4233b446851f42b57ad5031bf133e1dd76787f3e.md
---
`

	if expected != actual {
		t.Fatalf("zettel:\nexpected: %#v\n  actual: %#v", expected, actual)
	}
}

func TestWriteWithInlineBlob(t1 *testing.T) {
	t := ui.MakeTestContext(t1)

	z := &object_metadata.Metadata{
		Description: descriptions.Make("the title"),
		Type:        makeBlobExt(t, "md"),
	}

	z.SetTags(makeTagSet(t,
		"tag1",
		"tag2",
		"tag3",
	))

	envRepo := env_repo.MakeTesting(
		t,
		map[string]string{
			"fa8242e99f48966ca514092b4233b446851f42b57ad5031bf133e1dd76787f3e": "the body",
		},
	)

	format := object_metadata.MakeTextFormatterMetadataInlineBlob(
		object_metadata.Dependencies{
			BlobStore:      envRepo.GetDefaultBlobStore(),
			BlobDigestType: markl.HRPObjectBlobDigestSha256V1,
		},
	)

	actual := writeFormat(t, z, format, true, "the body",
		object_metadata.TextFormatterOptions{},
	)

	expected := `---
# the title
- tag1
- tag2
- tag3
! md
---

the body`

	t.AssertEqual(expected, actual)
}
