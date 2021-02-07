package m3u8

import (
	"io"
	"net/url"
)

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
