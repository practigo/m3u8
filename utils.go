package m3u8

import (
	"io"
	"net/url"
)

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
