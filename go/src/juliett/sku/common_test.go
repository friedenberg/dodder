package sku

import (
	"crypto/sha256"
	"io"
	"reflect"
	"strings"
	"testing"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_ptr"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/markl_io"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
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

	return es
}

func makeBlobExt(t *ui.TestContext, v string) (es ids.Type) {
	if err := es.Set(v); err != nil {
		t.Fatalf("%s", err)
	}

	return es
}

func readFormat(
	t1 *ui.TestContext,
	format object_metadata.TextFormat,
	contents string,
) (metadata object_metadata.IMetadataMutable) {
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

	metadata = object.GetMetadataMutable()

	return metadata
}

func TestMakeTags(t1 *testing.T) {
	ui.RunTestContext(t1, testMakeTags)
}

func testMakeTags(t *ui.TestContext) {
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
	ui.RunTestContext(t1, testEqualitySelf)
}

func testEqualitySelf(t *ui.TestContext) {
	text := &object_metadata.Metadata{
		Description: descriptions.Make("the title"),
		Type:        makeBlobExt(t, "text"),
	}

	text.SetTags(makeTagSet(t,
		"tag1",
		"tag2",
		"tag3",
	))

	if !object_metadata.Equaler.Equals(text, text) {
		t.Fatalf("expected %v to equal itself", text)
	}
}

func TestEqualityNotSelf(t1 *testing.T) {
	ui.RunTestContext(t1, testEqualityNotSelf)
}

func testEqualityNotSelf(t *ui.TestContext) {
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

	if !object_metadata.Equaler.Equals(&text, text1) {
		t.Fatalf("expected %v to equal %v", text, text1)
	}
}

func makeTestTextFormat(
	envDir env_dir.Env,
	blobStore interfaces.BlobStore,
) object_metadata.TextFormat {
	return object_metadata.MakeTextFormat(
		object_metadata.Dependencies{
			EnvDir:    envDir,
			BlobStore: blobStore,
		},
	)
}

func TestReadWithoutBlob(t1 *testing.T) {
	ui.RunTestContext(t1, testReadWithoutBlob)
}

func testReadWithoutBlob(t *ui.TestContext) {
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

	if !object_metadata.Equaler.Equals(actual, expected) {
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
	ui.RunTestContext(t1, testReadWithoutBlobWithMultilineDescription)
}

func testReadWithoutBlobWithMultilineDescription(t *ui.TestContext) {
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

	if !object_metadata.Equaler.Equals(actual, expected) {
		t.Fatalf("zettel:\nexpected: %#v\n  actual: %#v", expected, actual)
	}

	if !actual.GetBlobDigest().IsNull() {
		t.Fatalf("blob:\nexpected empty but got %q", actual.GetBlobDigest())
	}
}

func TestReadWithBlob(t1 *testing.T) {
	ui.RunTestContext(t1, testReadWithBlob)
}

func testReadWithBlob(t *ui.TestContext) {
	envRepo := env_repo.MakeTesting(
		t,
		nil,
	)

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

	var expectedBlobDigest markl.Id
	t.AssertNoError(expectedBlobDigest.Set(
		"blake2b256-9j5cj9mjnk43k9rq4k2h3lezpl2sn3ura7cf8pa58cgfujw6nwgst7gtwz",
	))

	expected := &object_metadata.Metadata{
		Description: descriptions.Make("the title"),
		Type:        makeBlobExt(t, "md"),
	}

	expected.GetBlobDigestMutable().ResetWithMarklId(expectedBlobDigest)

	expected.SetTags(makeTagSet(t,
		"tag1",
		"tag2",
		"tag3",
	))

	if !object_metadata.Equaler.Equals(actual, expected) {
		t.Fatalf("zettel:\nexpected: %#v\n  actual: %#v", expected, actual)
	}
}

type blobReaderFactory struct {
	t     *ui.TestContext
	blobs map[string]string
}

func (blobStore blobReaderFactory) BlobReader(
	digest interfaces.MarklId,
) (readCloser interfaces.BlobReader, err error) {
	var value string
	var ok bool

	if value, ok = blobStore.blobs[digest.String()]; !ok {
		blobStore.t.Fatalf("request for non-existent blob: %s", digest)
	}

	hashType, err := markl.GetFormatHashOrError(
		digest.GetMarklFormat().GetMarklFormatId(),
	)
	blobStore.t.AssertNoError(err)

	readCloser = markl_io.MakeNopReadCloser(
		hashType.Get(),
		ohio.NopCloser(strings.NewReader(value)),
	)

	return readCloser, err
}

func writeFormat(
	t *ui.TestContext,
	metadata object_metadata.IMetadataMutable,
	formatter object_metadata.TextFormatter,
	includeBlob bool,
	blobBody string,
	options object_metadata.TextFormatterOptions,
	hashType interfaces.FormatHash,
) (out string) {
	hash := sha256.New()
	reader, repool := pool.GetStringReader(blobBody)
	defer repool()
	_, err := io.Copy(hash, reader)
	if err != nil {
		t.Fatalf("%s", err)
	}

	blobDigest, _ := hashType.GetMarklIdForString(blobBody)

	metadata.GetBlobDigestMutable().ResetWithMarklId(blobDigest)

	stringBuilder := &strings.Builder{}

	if _, err := formatter.FormatMetadata(
		stringBuilder,
		object_metadata.TextFormatterContext{
			PersistentFormatterContext: metadata,
			TextFormatterOptions:       options,
		},
	); err != nil {
		t.Errorf("%s", err)
	}

	out = stringBuilder.String()

	return out
}

func TestWriteWithoutBlob(t1 *testing.T) {
	ui.RunTestContext(t1, testWriteWithoutBlob)
}

func testWriteWithoutBlob(t *ui.TestContext) {
	object := &object_metadata.Metadata{
		Description: descriptions.Make("the title"),
		Type:        makeBlobExt(t, "md"),
	}

	object.SetTags(makeTagSet(t,
		"tag1",
		"tag2",
		"tag3",
	))

	envRepo := env_repo.MakeTesting(
		t,
		map[string]string{
			"blake2b256-9j5cj9mjnk43k9rq4k2h3lezpl2sn3ura7cf8pa58cgfujw6nwgst7gtwz": "the body",
		},
	)

	format := object_metadata.MakeTextFormatterMetadataOnly(
		object_metadata.Dependencies{
			BlobStore: envRepo.GetDefaultBlobStore(),
		},
	)

	actual := writeFormat(
		t,
		object,
		format,
		false,
		"the body",
		object_metadata.TextFormatterOptions{},
		envRepo.GetDefaultBlobStore().GetDefaultHashType(),
	)

	expected := `---
# the title
- tag1
- tag2
- tag3
! blake2b256-9j5cj9mjnk43k9rq4k2h3lezpl2sn3ura7cf8pa58cgfujw6nwgst7gtwz.md
---
`

	if expected != actual {
		t.Fatalf("zettel:\nexpected: %#v\n  actual: %#v", expected, actual)
	}
}

func TestWriteWithInlineBlob(t1 *testing.T) {
	ui.RunTestContext(t1, testWriteWithInlineBlob)
}

func testWriteWithInlineBlob(t *ui.TestContext) {
	object := &object_metadata.Metadata{
		Description: descriptions.Make("the title"),
		Type:        makeBlobExt(t, "md"),
	}

	object.SetTags(makeTagSet(t,
		"tag1",
		"tag2",
		"tag3",
	))

	envRepo := env_repo.MakeTesting(
		t,
		map[string]string{
			"blake2b256-9j5cj9mjnk43k9rq4k2h3lezpl2sn3ura7cf8pa58cgfujw6nwgst7gtwz": "the body",
		},
	)

	format := object_metadata.MakeTextFormatterMetadataInlineBlob(
		object_metadata.Dependencies{
			BlobStore: envRepo.GetDefaultBlobStore(),
		},
	)

	actual := writeFormat(
		t,
		object,
		format,
		true,
		"the body",
		object_metadata.TextFormatterOptions{},
		envRepo.GetDefaultBlobStore().GetDefaultHashType(),
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
