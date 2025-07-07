package files

import (
	"io/fs"
	"iter"
	"os"
	"path"
	"path/filepath"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
)

// TODO migrate all dir walking to this package

type WalkDirEntry struct {
	Path    string
	RelPath string
	os.DirEntry
}

type WalkDirEntryIgnoreFunc func(WalkDirEntry) bool

func WalkDirIgnoreFuncHidden(dirEntry WalkDirEntry) bool {
	if strings.HasPrefix(dirEntry.RelPath, ".") {
		return true
	}

	return false
}

func WalkDir(
	base string,
) iter.Seq2[WalkDirEntry, error] {
	return func(yield func(WalkDirEntry, error) bool) {
		if err := filepath.WalkDir(
			base,
			func(path string, dirEntry os.DirEntry, in error) (out error) {
				if in != nil {
					out = in
					return
				}

				entry := WalkDirEntry{
					Path:     path,
					DirEntry: dirEntry,
				}

				if entry.RelPath, out = filepath.Rel(base, path); out != nil {
					out = errors.Wrap(out)
					return
				}

				if entry.RelPath == "." {
					return
				}

				if !yield(entry, nil) {
					out = fs.SkipAll
					return
				}

				return
			},
		); err != nil {
			yield(WalkDirEntry{}, errors.Wrap(err))
			return
		}
	}
}

func DirNames2(p string) iter.Seq2[os.DirEntry, error] {
	return func(yield func(os.DirEntry, error) bool) {
		var names []os.DirEntry

		{
			var err error

			if names, err = ReadDir(p); err != nil {
				yield(nil, errors.Wrap(err))
				return
			}
		}

		for _, dirEntry := range names {
			if !yield(dirEntry, nil) {
				return
			}
		}
	}
}

func DirNames(dirPath string) (slice quiter.Slice[string], err error) {
	var names []os.DirEntry

	if names, err = ReadDir(dirPath); err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, dirEntry := range names {
		slice.Append(path.Join(dirPath, dirEntry.Name()))
	}

	return
}

func DirNameWriterIgnoringHidden(
	seq iter.Seq2[string, error],
) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for path, err := range seq {
			if err != nil {
				yield(path, err)
				return
			}

			b := filepath.Base(path)

			if strings.HasPrefix(b, ".") {
				return
			}

			if !yield(path, err) {
				return
			}
		}
	}
}

func DirNamesLevel2(
	dirPath string,
) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		var topLevel quiter.Slice[string]

		{
			var err error

			if topLevel, err = DirNames(dirPath); err != nil {
				yield("", err)
				return
			}
		}

		for topLevelDir := range topLevel.All() {
			var secondLevel quiter.Slice[string]

			{
				var err error

				if secondLevel, err = DirNames(topLevelDir); err != nil {
					yield("", err)
					return
				}
			}

			for secondLevelDir := range secondLevel.All() {
				if !yield(secondLevelDir, nil) {
					return
				}
			}
		}
	}
}
