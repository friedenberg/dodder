package fd

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

func Ext(p string) string {
	return strings.ToLower(path.Ext(p))
}

func ExtSansDot(p string) string {
	return strings.ToLower(strings.TrimPrefix(path.Ext(p), "."))
}

func FileNameSansExt(p string) string {
	base := filepath.Base(p)
	ext := Ext(p)
	return base[:len(base)-len(ext)]
}

func FileNameSansExtRelTo(p, d string) (string, error) {
	rel, err := filepath.Rel(d, p)
	if err != nil {
		return "", err
	}

	base := filepath.Base(rel)
	ext := Ext(rel)

	return base[:len(base)-len(ext)], nil
}

func DirBaseOnly(p string) string {
	return filepath.Base(filepath.Dir(p))
}

func ZettelId(p string) string {
	return fmt.Sprintf("%s/%s", DirBaseOnly(p), FileNameSansExt(p))
}
