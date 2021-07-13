package m3u8

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	variantTag = "#EXT-X-STREAM-INF"
	mediaTag   = "#EXT-X-MEDIA"
)

// Attributes is a list of key:value pairs.
type Attributes string

// XStream is a Variant Stream.
type XStream struct {
	URI  string
	Info Attributes // EXT-X-STREAM-INF

	// X-Media are alternative Renditions.
	// They are not necessarily binded to a certain XStream/variant,
	// but put inside a XStream for keeping relative output position
	XMedia []Attributes

	// optional
	Directives string
}

func (s XStream) Marshal() string {
	ret := s.Directives
	for _, m := range s.XMedia {
		ret += fmt.Sprintf("%s:%s\n", mediaTag, m)
	}
	if len(s.XMedia) > 0 {
		ret += "\n" // extra black line
	}
	ret += fmt.Sprintf("%s:%s\n", variantTag, s.Info)
	ret += s.URI + "\n"
	return ret
}

// Master is a Master playlist.
type Master struct {
	// 	A Master Playlist MUST indicate a EXT-X-VERSION of 7 or higher if it
	//    contains:
	//     - "SERVICE" values for the INSTREAM-ID attribute of the EXT-X-MEDIA
	//       tag.
	//    The EXT-X-MEDIA tag and the AUDIO, VIDEO and SUBTITLES attributes of
	//    the EXT-X-STREAM-INF tag are backward compatible to protocol version
	//    1, but playback on older clients may not be desirable.  A server MAY
	//    consider indicating a EXT-X-VERSION of 4 or higher in the Master
	//    Playlist but is not required to do so.
	Version int

	Streams []XStream
}

func (m *Master) Marshal() string {
	ret := starter + "\n"
	if m.Version > 0 {
		ret += fmt.Sprintf("%s:%d\n", verTag, m.Version)
	}
	for _, s := range m.Streams {
		ret += s.Marshal() + "\n" // extra blank line for style
	}
	return ret
}

// IsMaster tells if a playlist is a master playlist.
func IsMaster(r io.Reader) (bool, error) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		l := s.Text()
		if strings.HasPrefix(l, infTag) {
			return false, nil
		}
		if strings.HasPrefix(l, variantTag) {
			return true, nil
		}
	}

	err := s.Err()
	if err == nil {
		err = errors.New("invalid playlist: neither master nor media playlist")
	}

	return false, err
}

// DecodeMaster decodes a master playlist.
func DecodeMaster(r io.Reader) (m *Master, err error) {
	m = &Master{
		Streams: make([]XStream, 0),
	}
	var cur XStream

	s := bufio.NewScanner(r)
	for s.Scan() {
		l := strings.TrimSpace(s.Text())

		switch {
		case l == starter || len(l) < 1:
			// ignore blank lines
		case !strings.HasPrefix(l, "#"):
			if cur.Info == "" {
				return nil, ErrMissingStream
			}
			cur.URI = l
			m.Streams = append(m.Streams, cur)
			cur = XStream{}
		case strings.HasPrefix(l, "#EXT"):
			switch {
			case strings.HasPrefix(l, variantTag):
				cur.Info = Attributes(l[18:]) // 18 = len("#EXT-X-STREAM-INF:")
			case strings.HasPrefix(l, mediaTag):
				cur.XMedia = append(cur.XMedia, Attributes(l[13:])) // 13 = len("#EXT-X-MEDIA:")
			default:
				cur.Directives += l + "\n"
			}
		default:
			// ignore comments
		}

	}
	err = s.Err()

	return
}

// TryDecodeMaster tests if the read content is a master playlist
// and try to decode it; otherwise return ErrNotMaster.
func TryDecodeMaster(r io.Reader) (m *Master, err error) {
	bs, err := io.ReadAll(r)
	if err != nil {
		return
	}

	isMaster, err := IsMaster(bytes.NewBuffer(bs))
	if err != nil {
		return
	}

	if isMaster {
		return DecodeMaster(bytes.NewBuffer(bs))
	}

	err = ErrNotMaster
	return
}
