package m3u8

import (
	"bytes"
	"container/list"
	"fmt"
	"io"
	"strings"
)

// Range is a Media Segment from a sub-range of the resource.
type Range struct {
	Len   int
	Start *int
}

func (r *Range) String() string {
	if r.Start != nil {
		return fmt.Sprintf("%d@%d", r.Len, *r.Start)
	}
	return fmt.Sprintf("%d", r.Len)
}

// XMapAttr is the attributes of the Media Segment init section.
type XMapAttr struct {
	URI       string
	ByteRange *Range
}

func (m *XMapAttr) String() string {
	if m.ByteRange != nil {
		return fmt.Sprintf(`URI="%s",BYTERANGE="%s"`, m.URI, m.ByteRange)
	}
	return fmt.Sprintf(`URI="%s"`, m.URI)
}

// Entry is a m3u8 entry.
type Entry struct {
	URI           string
	Duration      float64
	Title         string // optional second parameter for EXTINF tag
	Discontinuity bool

	// optional
	ByteRange *Range
	XMap      *XMapAttr

	Directives string // Directives separated by '\n' to avoid []string or map for easy comparison

	e *list.Element
}

// NewEntry returns a pointer to an empty Entry whose
// Duration is set to -1 since 0 is valid.
func NewEntry() *Entry {
	ret := new(Entry)
	ret.Duration = -1.0
	return ret
}

// MarshalTo writes the string form of the Entry to a Writer.
func (s *Entry) marshalTo(w io.Writer) {
	if s.XMap != nil {
		writeLine(w, fmt.Sprintf("%s:%s", mapTag, s.XMap))
	}

	if s.Discontinuity {
		writeLine(w, disTag)
	}

	if s.Directives != "" {
		writeLine(w, strings.TrimSpace(s.Directives))
	}

	writeLine(w, fmt.Sprintf("#EXTINF:%.6f,%s", s.Duration, s.Title))

	if s.ByteRange != nil {
		writeLine(w, fmt.Sprintf("%s:%s", rangeTag, s.ByteRange))
	}

	writeLine(w, s.URI)
}

// Marshal returns the string form of the Entry
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
