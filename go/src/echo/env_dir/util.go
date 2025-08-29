package env_dir

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/unicorn"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
)

func MakeHashBucketPathFromMerkleId(
	merkleId interfaces.BlobId,
	buckets []int,
	pathComponents ...string,
) string {
	return MakeHashBucketPath(
		// TODO must be a way to make this more performant instead of a double
		// copy
		[]byte(merkle.Format(merkleId)),
		buckets,
		pathComponents...,
	)
}

func MakeHashBucketPath(
	hashBytes []byte,
	buckets []int,
	pathComponents ...string,
) string {
	var buffer bytes.Buffer

	for _, pathComponent := range pathComponents {
		// pathComponent = strings.TrimPrefix(
		// 	pathComponent,
		// 	string(filepath.Separator),
		// )

		pathComponent = strings.TrimRight(
			pathComponent,
			string(filepath.Separator),
		)

		buffer.WriteString(pathComponent)
		buffer.WriteByte(filepath.Separator)
	}

	remaining := hashBytes

	for _, bucket := range buckets {
		if len(remaining) < bucket {
			panic(
				fmt.Sprintf(
					"buckets too large for string. buckets: %v, string: %q, remaining: %q",
					buckets,
					hashBytes,
					remaining,
				),
			)
		}

		var added []byte

		// TODO check that added and remaining to not contain filepath.Separator
		added, remaining = unicorn.CutNCharacters(remaining, bucket)

		buffer.Write(added)
		buffer.WriteByte(filepath.Separator)
	}

	if len(remaining) > 0 {
		buffer.Write(remaining)
	}

	return buffer.String()
}

func PathFromHeadAndTail(
	stringer interfaces.StringerWithHeadAndTail,
	pathComponents ...string,
) string {
	pathComponents = append(
		pathComponents,
		stringer.GetHead(),
		stringer.GetTail(),
	)

	return filepath.Join(pathComponents...)
}

func MakeHashBucketPathJoinFunc(
	buckets []int,
) func(string, ...string) string {
	return func(initial string, pathComponents ...string) string {
		return MakeHashBucketPath(
			[]byte(initial),
			buckets,
			pathComponents...)
	}
}

// TODO migrate to using Env and accepting path generation function
func MakeDirIfNecessary(
	base string,
	joinFunc func(string, ...string) string,
	pathComponents ...string,
) (path string, err error) {
	path = joinFunc(base, pathComponents...)
	dir := filepath.Dir(path)

	if err = os.MkdirAll(dir, os.ModeDir|0o755); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO migrate to using Env and accepting path generation function
func MakeDirIfNecessaryForStringerWithHeadAndTail(
	stringer interfaces.StringerWithHeadAndTail,
	pathComponents ...string,
) (path string, err error) {
	path = PathFromHeadAndTail(stringer, pathComponents...)
	dir := filepath.Dir(path)

	if err = os.MkdirAll(dir, os.ModeDir|0o755); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
