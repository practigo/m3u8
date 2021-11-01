package m3u8

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// by rfc8216
// https://datatracker.ietf.org/doc/html/draft-pantos-http-live-streaming
const (
	// basic tags
	starter = "#EXTM3U"
	verTag  = "#EXT-X-VERSION"

	// media playlist tags
	ender  = "#EXT-X-ENDLIST"
	durTag = "#EXT-X-TARGETDURATION" // required
	seqTag = "#EXT-X-MEDIA-SEQUENCE" // default 0

	// segment tags
	infTag   = "#EXTINF"
	disTag   = "#EXT-X-DISCONTINUITY"
	rangeTag = "#EXT-X-BYTERANGE"
	mapTag   = "#EXT-X-MAP"
)

var playlistTags = []string{
	// media only
	durTag,                          // required
	seqTag,                          // default 0
	"#EXT-X-DISCONTINUITY-SEQUENCE", // default 0
	"#EXT-X-PLAYLIST-TYPE",          // EVENT or VOD
	"#EXT-X-I-FRAMES-ONLY",
	ender,
	// basic
	starter,
	verTag,
	// both media & master
	"#EXT-X-INDEPENDENT-SEGMENTS",
	"#EXT-X-START", // contains a attribute list
	// master only
}

func IsMediaPlaylistTag(line string) bool {
	for _, tag := range playlistTags[0:10] {
		if strings.HasPrefix(line, tag) {
			return true
		}
	}
	return false
}

// decoding states
type decodecState struct {
	cur *Entry
}

func parseInt(line, tag string) (int, error) {
	return strconv.Atoi(line[len(tag)+1:]) // plus one for ":"
}

func parseRangeTag(l string) (r *Range, err error) {
	// #EXT-X-BYTERANGE:<n>[@<o>]
	r = new(Range)
	if sepIndex := strings.Index(l, "@"); sepIndex > -1 {
		r.Len, err = strconv.Atoi(l[:sepIndex])
		if err != nil {
			return
		}
		var st int
		st, err = strconv.Atoi(l[sepIndex+1:])
		r.Start = &st
	} else {
		r.Len, err = strconv.Atoi(l)
	}
	return
}

func parseMapTag(l string) (m *XMapAttr, err error) {
	m = new(XMapAttr)
	attr, err := SplitAttributeList(l)
	if err != nil {
		return
	}
	if uri, ok := attr["URI"]; ok {
		m.URI = uri
	} else {
		return m, ErrMissingAttr
	}
	if br, ok := attr["BYTERANGE"]; ok {
		m.ByteRange, err = parseRangeTag(strings.Trim(br, "\""))
	}
	return
}

func decodeLine(l string, p *Playlist, state *decodecState) (err error) {
	// Each line is a URI, is blank, or starts with the
	// character '#'.  Blank lines are ignored.
	// Whitespace MUST NOT be present
	l = strings.TrimSpace(l) // here leading or trailing spaces are removed
	if len(l) < 1 {
		return
	}

	if !strings.HasPrefix(l, "#") {
		// URI
		if state.cur.Duration == 0.0 {
			return ErrMissingInf
		}
		state.cur.URI = l
		p.Append(*state.cur)
		state.cur = new(Entry)
		return nil
	}

	if !strings.HasPrefix(l, "#EXT") {
		// ignore comments
		return nil
	}

	// handle tags
	if strings.HasPrefix(l, infTag) {
		if sepIndex := strings.Index(l, ","); sepIndex > -1 {
			state.cur.Duration, err = strconv.ParseFloat(l[8:sepIndex], 64)
			if len(l) > sepIndex {
				state.cur.Title = l[sepIndex+1:]
			}
		} else {
			err = errors.New("no \",\" in info tag")
		}
		return err
	}

	if strings.HasPrefix(l, rangeTag) {
		state.cur.ByteRange, err = parseRangeTag(l[len(rangeTag)+1:]) // + ":"
		return err
	}

	if strings.HasPrefix(l, mapTag) {
		state.cur.XMap, err = parseMapTag(l[len(mapTag)+1:]) // + ":"
		return err
	}

	if l == disTag {
		state.cur.Discontinuity = true
		return nil
	}

	if IsMediaPlaylistTag(l) {
		switch {
		case l == starter:
			// pass
		case l == ender:
			p.Closed = true
		case strings.HasPrefix(l, verTag):
			p.Version, err = parseInt(l, verTag)
		case strings.HasPrefix(l, durTag):
			p.TargetDuration, err = parseInt(l, durTag)
		case strings.HasPrefix(l, seqTag):
			p.SeqNo, err = parseInt(l, seqTag)
		default:
			p.Directives += l + "\n"
		}
		if err != nil {
			return fmt.Errorf("parse line %s: %w", l, err)
		}
		return nil
	}

	// all other tags including private ones
	state.cur.Directives += l + "\n"
	return nil
}

// Decode reads from r and tries to decode it as a
// m3u8 index file.
func Decode(r io.Reader) (p *Playlist, err error) {
	p = New()

	state := &decodecState{
		cur: new(Entry),
	}

	s := bufio.NewScanner(r)
	for s.Scan() {
		if err = decodeLine(s.Text(), p, state); err != nil {
			return
		}
	}
	err = s.Err()

	return
}

// DecodeStrict is DEPRECATED. Use Decodee instead.
func DecodeStrict(r io.Reader, strict bool) (p *Playlist, err error) {
	return Decode(r)
}
