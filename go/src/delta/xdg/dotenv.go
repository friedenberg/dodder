package xdg

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type Dotenv struct {
	*XDG
}

func (d Dotenv) setDefaultOrEnvFromMap(
	defawlt string,
	envKey string,
	out *string,
	env map[string]string,
) (err error) {
	if v, ok := env[envKey]; ok {
		*out = v
	} else {
		*out = os.Expand(defawlt, func(v string) string {
			switch v {
			case "HOME":
				return d.Home

			default:
				return os.Getenv(v)
			}
		})
	}

	return
}

func (d Dotenv) ReadFrom(r io.Reader) (n int64, err error) {
	env := make(map[string]string)

	br := bufio.NewReader(r)
	eof := false

	for !eof {
		var line string
		line, err = br.ReadString('\n')
		n += int64(len(line))

		if err == io.EOF {
			eof = true
			err = nil
		} else if err != nil {
			err = errors.Wrap(err)
			return
		}

		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		left, right, ok := strings.Cut(line, "=")

		if !ok {
			err = errors.ErrorWithStackf("malformed env var entry: %q", line)
			return
		}

		env[left] = right
	}

	toInitialize := d.getInitElements()

	for _, ie := range toInitialize {
		if err = d.setDefaultOrEnvFromMap(
			ie.standard,
			ie.envKey,
			ie.out,
			env,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (d Dotenv) WriteTo(w io.Writer) (n int64, err error) {
	bw := bufio.NewWriter(w)

	toWrite := d.getInitElements()
	var n1 int

	for _, e := range toWrite {
		n1, err = fmt.Fprintf(bw, "%s=%s\n", e.envKey, *e.out)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if err = bw.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
