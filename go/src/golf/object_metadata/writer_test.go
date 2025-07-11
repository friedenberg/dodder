package object_metadata

import (
	"strings"
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

func TestWriter1(t1 *testing.T) {
	t := ui.T{
		T: t1,
	}

	expectedOut := `---
metadatei
---

blob
`

	out := &strings.Builder{}

	sut := triple_hyphen_io.Writer{
		Metadata: strings.NewReader("metadatei\n"),
		Blob:     strings.NewReader("blob\n"),
	}

	sut.WriteTo(out)

	if out.String() != expectedOut {
		t.Errorf("expected %q but got %q", expectedOut, out.String())
	}
}

func TestWriter2(t1 *testing.T) {
	t := ui.T{
		T: t1,
	}

	expectedOut := `---
metadatei
---
`

	out := &strings.Builder{}

	sut := triple_hyphen_io.Writer{
		Metadata: strings.NewReader("metadatei\n"),
	}

	sut.WriteTo(out)

	if out.String() != expectedOut {
		t.Errorf("expected %q but got %q", expectedOut, out.String())
	}
}

func TestWriter3(t1 *testing.T) {
	t := ui.T{
		T: t1,
	}

	expectedOut := `blob
`

	out := &strings.Builder{}

	sut := triple_hyphen_io.Writer{
		Blob: strings.NewReader("blob\n"),
	}

	sut.WriteTo(out)

	if out.String() != expectedOut {
		t.Errorf("expected %q but got %q", expectedOut, out.String())
	}
}
