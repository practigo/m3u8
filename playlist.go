package m3u8

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"io"
	"strings"
)

// exported errors
var (
	// ErrUnknownLine = errors.New("unknown line") // DEPRECATED
	ErrMissingInf      = errors.New("missing #EXTINF tag")
	ErrInvalidSeg      = errors.New("invalid Entry")
	ErrMissingStream   = errors.New("missing #EXT-X-STREAM-INF tag")
	ErrInvalidPlaylist = errors.New("invalid playlist: neither master nor media playlist")
	ErrNotMaster       = errors.New("not master playlist")
	ErrMissingAttr     = errors.New("missing required attribute")
)

// Playlist is a m3u8 playlist.
type Playlist struct {
	TargetDuration int    // EXT-X-TARGETDURATION
	SeqNo          int    // EXT-X-MEDIA-SEQUENCE
	Version        int    // EXT-X-VERSION
	Closed         bool   // EXT-X-ENDLIST
	Directives     string // all the un-parsed
	// IOnly          bool
	// Indi           bool
	// Type           string
	// StartAttr      string
	// DisSeq         int

	// internal list
	l *list.List
}

// New returns an empty m3u8 playlist.
func New() *Playlist {
	p := new(Playlist)
	p.l = list.New()
	p.SeqNo = -1
	p.Version = -1
	return p
}

// Append adds an entry to the end of the playlist.
func (p *Playlist) Append(s Entry) {
	e := p.l.PushBack(&s)
	s.e = e
}

// Front returns the first entry of the playlist or
// nil if it's empty.
func (p *Playlist) Front() *Entry {
	e := p.l.Front()
	if e != nil {
		return e.Value.(*Entry)
	}
	return nil
}

// Remove ...
func (p *Playlist) Remove(s *Entry) error {
	if s.e != nil {
		p.l.Remove(s.e)
		return nil
	}
	return ErrInvalidSeg
}

// InsertAfter ...
func (p *Playlist) InsertAfter(ref *Entry, target Entry) error {
	if ref.e != nil {
		e := p.l.InsertAfter(&target, ref.e)
		target.e = e
		return nil
	}
	return ErrInvalidSeg
}

// MarshalTo ...
func (p *Playlist) MarshalTo(w io.Writer) {
	writeLine(w, starter)
	if p.Version > 0 {
		writeLine(w, fmt.Sprintf(verTag+":%d", p.Version))
	}
	if p.TargetDuration > 0 {
		writeLine(w, fmt.Sprintf(durTag+":%d", int(p.TargetDuration)))
	}
	if p.SeqNo >= 0 {
		writeLine(w, fmt.Sprintf(seqTag+":%d", p.SeqNo))
	}
	if p.Directives != "" {
		writeLine(w, strings.TrimSpace(p.Directives))
	}

	for s := p.Front(); s != nil; s = s.Next() {
		s.marshalTo(w)
	}

	if p.Closed {
		writeLine(w, ender)
	}
}

// MarshalToWithErr ...
func (p *Playlist) MarshalToWithErr(iw io.Writer) error {
	w := &writer{iw: iw}
	p.MarshalTo(w)

	return w.err
}

// Marshal ...
func (p *Playlist) Marshal() string {
	var b bytes.Buffer
	// b.Write never returns error
	p.MarshalTo(&b)
	return b.String()
}
