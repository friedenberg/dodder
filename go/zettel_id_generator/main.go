package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func main() {
	type tokenInfo struct {
		count int
	}

	tokens := make(map[string]tokenInfo)

	br := bufio.NewReader(os.Stdin)
	tr := transform.Chain(
		norm.NFD,
		transform.RemoveFunc(
			func(r rune) bool {
				return unicode.Is(unicode.Mn, r)
			},
		),
		norm.NFC,
	)

	for {
		line, err := br.ReadString('\n')

		if err != io.EOF && err != nil {
			log.Fatalf("%s", err)
		}

		isEOF := err == io.EOF

		if len(line) > 0 {
			line = line[:len(line)-1]
		}

		sc := bufio.NewScanner(strings.NewReader(line))
		sc.Split(bufio.ScanWords)

		theseTokens := make([]string, 0)

		for sc.Scan() {
			t := sc.Text()

			if len(t) <= 3 {
				continue
			}

			if strings.ContainsFunc(t, unicode.IsPunct) {
				continue
			}

			b := make([]byte, len(t))

			_, _, err := tr.Transform(b, []byte(t), true)
			if err != nil {
				log.Fatalf("%s", err)
			}

			t = string(bytes.Trim(b, "\x00"))

			theseTokens = append(theseTokens, t)
		}

		switch len(theseTokens) {
		case 1:
			t := theseTokens[0]
			ti := tokens[t]
			ti.count += 1
			tokens[t] = ti

		case 0:
		default:
			t := theseTokens[len(theseTokens)-1]
			ti := tokens[t]
			ti.count += 1
			tokens[t] = ti
		}

		if isEOF {
			break
		}
	}

	sorted := make([]string, 0, len(tokens))

	for k, v := range tokens {
		if v.count == 1 {
			sorted = append(sorted, k)
			delete(tokens, k)
		}
	}

	sort.Strings(sorted)

	for _, v := range sorted {
		fmt.Println(v)
	}
}
