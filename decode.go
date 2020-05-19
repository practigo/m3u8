package m3u8

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	starter = "#EXTM3U"
	ender   = "#EXT-X-ENDLIST"

	infTag = "#EXTINF"
	disTag = "#EXT-X-DISCONTINUITY"

	// infTagFormat = "#EXTINF:%.6f,%s"
	verTagFormat = "#EXT-X-VERSION:%d"
	durTagFormat = "#EXT-X-TARGETDURATION:%d"
	seqTagFormat = "#EXT-X-MEDIA-SEQUENCE:%d"
)

var formatPrefix = map[string]string{
	// infTagFormat: infTagFormat[:8],
	verTagFormat: verTagFormat[:15],
	durTagFormat: durTagFormat[:22],
	seqTagFormat: seqTagFormat[:22],
}

// #EXTM3U
// #EXT-X-VERSION:3
// #EXT-X-TARGETDURATION:10
// #EXT-X-MEDIA-SEQUENCE:0
// #EXT-X-DISCONTINUITY
// #EXTINF:10.000000,
// #EXT-X-MISC:SOME-DATA
// a.ts
// #EXT-X-ENDLIST
func decodeLine(l string, p *Playlist) (err error) {
	l = strings.TrimSpace(l)

	if !strings.HasPrefix(l, "#") {
		p.cur.URI = l
		p.Append(*p.cur)
		p.cur = new(Entry)
		return nil
	}

	parseLine := func(line, format string, a ...interface{}) bool {
		if strings.HasPrefix(line, formatPrefix[format]) {
			_, err = fmt.Sscanf(line, format, a...) // err as closure
			return true
		}
		return false
	}

	switch {
	case l == starter:
		// normal
	case l == ender:
		p.Closed = true
	case l == disTag:
		p.cur.Discontinuity = true
	case parseLine(l, verTagFormat, &p.Version) ||
		parseLine(l, seqTagFormat, &p.SeqNo) ||
		parseLine(l, durTagFormat, &p.TargetDuration):
	case strings.HasPrefix(l, infTag):
		p.entriesStarted = true
		if sepIndex := strings.Index(l, ","); sepIndex > -1 {
			p.cur.Duration, err = strconv.ParseFloat(l[8:sepIndex], 64)
			if len(l) > sepIndex {
				p.cur.Title = l[sepIndex+1:]
			}
		} else {
			err = errors.New("no \",\" in info tag")
		}
	case strings.HasPrefix(l, "#EXT-X"):
		if p.entriesStarted {
			p.cur.Directives += l + "\n"
		} else {
			p.Directives += l + "\n"
		}
	default:
		// if strict should be error
		err = ErrUnknownLine
	}

	if err != nil {
		err = fmt.Errorf("parse line %s: %s", l, err.Error())
	}
	return
}

// Decode reads from r and tries to decode it as a
// m3u8 index file.
func Decode(r io.Reader) (p *Playlist, err error) {
	p = New()
	p.cur = new(Entry)

	s := bufio.NewScanner(r)
	for s.Scan() {
		if err = decodeLine(s.Text(), p); err != nil {
			return
		}
	}
	err = s.Err()

	return
}
