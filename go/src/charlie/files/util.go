package files

import (
	"fmt"
	"path/filepath"
	"strings"
)

func PathElements(p string) []string {
	extRaw := filepath.Ext(p)
	ext := strings.TrimPrefix(strings.ToLower(extRaw), ".")

	var name string
	p, name = filepath.Split(strings.TrimSuffix(p, extRaw))
	out := []string{ext, name}

	for p != "" {
		f := filepath.Base(p)

		if f != "" {
			out = append(out, f)
		}

		p = strings.TrimSuffix(p, fmt.Sprintf("%s%c", f, filepath.Separator))

		if p == string(filepath.Separator) {
			out = append(out, p)
			break
		}
	}

	return out
}

func DirectoriesRelativeTo(p string) []string {
	extRaw := filepath.Ext(p)
	p, _ = filepath.Split(strings.TrimSuffix(p, extRaw))
	out := []string{}

	for p != "" {
		f := filepath.Base(p)

		if f != "" {
			out = append(out, f)
		}

		p = strings.TrimSuffix(p, fmt.Sprintf("%s%c", f, filepath.Separator))

		if p == string(filepath.Separator) {
			out = append(out, p)
			break
		}
	}

	return out
}
