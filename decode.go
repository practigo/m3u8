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
const (
	// basic tags
	starter = "#EXTM3U"
	verTag  = "#EXT-X-VERSION"

	// media playlist tags
	ender     = "#EXT-X-ENDLIST"
	durTag    = "#EXT-X-TARGETDURATION"         // required
	seqTag    = "#EXT-X-MEDIA-SEQUENCE"         // default 0
	disSeqTag = "#EXT-X-DISCONTINUITY-SEQUENCE" // default 0
	typeTag   = "#EXT-X-PLAYLIST-TYPE"          // EVENT or VOD
	iOnlyTag  = "#EXT-X-I-FRAMES-ONLY"
	// both master and playlist tags
	indiTag  = "#EXT-X-INDEPENDENT-SEGMENTS"
	startTag = "#EXT-X-START" // contains a attribute list

	// segment tags
	infTag = "#EXTINF"
	disTag = "#EXT-X-DISCONTINUITY"
)

var tagFormats = map[string]string{
	verTag: verTag + ":%d",
	durTag: durTag + ":%d",
	seqTag: seqTag + ":%d",
	// disSeqTag: disSeqTag + ":%d",
	// typeTag:   typeTag + ":%s",
	// startTag:  startTag + ":%s",
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
	// Each line is a URI, is blank, or starts with the
	// character '#'.  Blank lines are ignored.
	// Whitespace MUST NOT be present
	l = strings.TrimSpace(l) // here leading or trailing spaces are removed
	if len(l) < 1 {
		return
	}

	parseLine := func(line, tag string, a ...interface{}) bool {
		if strings.HasPrefix(line, tag) {
			_, err = fmt.Sscanf(line, tagFormats[tag], a...) // err as closure
			return true
		}
		return false
	}

	if !p.entriesStarted {
		// parse playlist tags
		switch {
		case l == starter:
			// normal
		case strings.HasPrefix(l, typeTag) ||
			strings.HasPrefix(l, indiTag) ||
			strings.HasPrefix(l, iOnlyTag) ||
			strings.HasPrefix(l, disSeqTag) ||
			strings.HasPrefix(l, startTag):
			p.Directives += l + "\n"
		// case l == indiTag:
		// 	p.Indi = true
		// case l == iOnlyTag:
		// 	p.IOnly = true
		case parseLine(l, verTag, &p.Version) ||
			parseLine(l, seqTag, &p.SeqNo) ||
			// parseLine(l, startTag, &p.StartAttr) ||
			// parseLine(l, typeTag, &p.Type) ||
			// parseLine(l, disSeqTag, &p.DisSeq) ||
			parseLine(l, durTag, &p.TargetDuration):
			// pass
		case strings.HasPrefix(l, "#EXT"):
			// other tags than previous
			p.entriesStarted = true
		case strings.HasPrefix(l, "#"):
			// comments ingored
		default:
			err = ErrUnknownLine
		}
	}

	if p.entriesStarted {
		switch {
		case l == ender:
			p.Closed = true
		case !strings.HasPrefix(l, "#"):
			if p.cur.Duration == 0.0 {
				return ErrMissingInf
			}
			p.cur.URI = l
			p.Append(*p.cur)
			p.cur = new(Entry)
		case strings.HasPrefix(l, "#EXT"):
			if l == disTag {
				p.cur.Discontinuity = true
			}
			if strings.HasPrefix(l, infTag) {
				if sepIndex := strings.Index(l, ","); sepIndex > -1 {
					p.cur.Duration, err = strconv.ParseFloat(l[8:sepIndex], 64)
					if len(l) > sepIndex {
						p.cur.Title = l[sepIndex+1:]
					}
				} else {
					err = errors.New("no \",\" in info tag")
				}
			}
			// all add to directives
			p.cur.Directives += l + "\n"
		default:
			// same as strings.HasPrefix(l, "#"):
			// comments ingored
		}
	}

	if err != nil {
		err = fmt.Errorf("parse line %s: %w", l, err)
	}
	return
}

// Decode reads from r and tries to decode it as a
// m3u8 index file.
func Decode(r io.Reader) (p *Playlist, err error) {
	return DecodeStrict(r, true)
}

// DecodeStrict reads from r and tries to decode it as a
// m3u8 index file. If strict, directive line not start
// #EXT will be consider error and returns an ErrUnknownLine.
func DecodeStrict(r io.Reader, strict bool) (p *Playlist, err error) {
	p = New()
	p.cur = new(Entry)

	s := bufio.NewScanner(r)
	for s.Scan() {
		if err = decodeLine(s.Text(), p); err != nil {
			if errors.Is(err, ErrUnknownLine) && !strict {
				continue
			}
			return
		}
	}
	err = s.Err()

	return
}
