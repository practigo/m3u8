package m3u8

import (
	"errors"
	"io"
	"net/url"
	"regexp"
	"strings"
)

// writer writes to the underlying io.Writer, record the
// first error encounter and stop writing from there.
type writer struct {
	err error
	iw  io.Writer
}

func (w *writer) Write(p []byte) (int, error) {
	if w.err == nil {
		n, err := w.iw.Write(p)
		w.err = err
		return n, err
	}
	return 0, nil
}

func writeLine(w io.Writer, l string) (int, error) {
	return w.Write([]byte(l + "\n"))
}

// ResolveURL resolves an entry of a m3u8 index to an absolute path.
func ResolveURL(index, entry string) (resolved string, err error) {
	i, err := url.Parse(index)
	if err != nil {
		return
	}
	ref, err := url.Parse(entry)
	if err != nil {
		return
	}
	resolved = i.ResolveReference(ref).String()
	return
}

var regSplitAttr = regexp.MustCompile(`[^,]+="([^"]*)"|[^,]+=[^,]+`)

// SplitAttributeList split the attribute list where the separator ","
// might be surrounded by "". The return value of the attribute map is
// un-quoted.
func SplitAttributeList(l string) (attr map[string]string, err error) {
	attr = make(map[string]string)
	matches := regSplitAttr.FindAllString(l, -1)
	for _, m := range matches {
		ps := strings.Split(m, "=")
		if len(ps) < 2 {
			err = errors.New("invalid attribute list: " + m)
			return
		}
		attr[ps[0]] = strings.Trim(ps[1], "\"")
	}
	return
}
