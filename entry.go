package m3u8

import (
	"bytes"
	"container/list"
	"fmt"
	"io"
	"strings"
)

// Entry is a m3u8 entry.
type Entry struct {
	URI           string
	Duration      float64
	Title         string // optional second parameter for EXTINF tag
	Discontinuity bool

	Directives string // Directives separated by '\n' to avoid []string or map for easy comparison

	e *list.Element
}

func (s *Entry) marshalTo(w io.Writer) {
	if s.Discontinuity {
		writeLine(w, disTag)
	}

	if s.Directives != "" {
		writeLine(w, strings.TrimSpace(s.Directives))
	}

	writeLine(w, fmt.Sprintf("#EXTINF:%.6f,%s", s.Duration, s.Title))

	writeLine(w, s.URI)
}

// Marshal returns the string form of a entry
// with all its directives.
func (s *Entry) Marshal() string {
	var b bytes.Buffer
	s.marshalTo(&b)
	return b.String()
}

// Next returns the entry next to itself if
// it's in a m3u8 playlist, or nil otherwise.
func (s *Entry) Next() *Entry {
	if s.e != nil && s.e.Next() != nil {
		return s.e.Next().Value.(*Entry)
	}
	return nil
}
