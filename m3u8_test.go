package m3u8_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/practigo/m3u8"
)

func TestM3U8(t *testing.T) {
	f, err := os.Open("testdata/sample.m3u8")
	if err != nil {
		t.Fatal("open test file", err)
	}
	defer f.Close()

	p, err := m3u8.Decode(f)
	if err != nil {
		t.Fatal("decode m3u8", err)
	}

	// 	firstTags := `#EXT-X-CUSTOM:some-key-values
	// #EXTINF:10.000000,
	// #EXT-X-TAG:VALUE=5954336
	// 00001.ts
	// `
	// first := p.Front()
	// 	if first != nil {
	// 		t.Log(first.Marshal())
	// 		if first.Marshal() != firstTags {
	// 			// t.Log(first.Marshal())
	// 			t.Error("should retains exact tags in order")
	// 		}
	// 	} else {
	// 		t.Fatal("fail to decode")
	// 	}

	second := p.Front().Next()
	if second != nil {
		// t.Log(second.Directives)
	} else {
		t.Fatal("fail to decode next")
	}

	p.InsertAfter(second, m3u8.Entry{
		URI:      "02-b.ts",
		Duration: 6.3,
	})
	p.InsertAfter(second, m3u8.Entry{
		URI:      "02-a.ts",
		Duration: 3.7,
		//
		Discontinuity: true,
	})
	p.Remove(second)

	out := p.Marshal()
	t.Log(out)

	_, err = m3u8.Decode(bytes.NewBuffer([]byte(out)))
	if err != nil {
		t.Error("marshal output should be decode right:", err)
	}
}

func TestMarshal(t *testing.T) {
	f, err := os.Open("testdata/sample.m3u8")
	if err != nil {
		t.Fatal("open test file", err)
	}
	defer f.Close()

	p, err := m3u8.Decode(f)
	if err != nil {
		t.Fatal("decode m3u8", err)
	}
	// error test
	tmpFile := filepath.Join(os.TempDir(), "test.m3u8")
	tf, err := os.OpenFile(tmpFile, os.O_RDONLY|os.O_CREATE|os.O_TRUNC, 0444)
	if err != nil {
		t.Fatal(err)
	}
	err = p.MarshalToWithErr(tf)
	if err == nil {
		t.Error("should have write error")
	} else {
		t.Log(err)
	}
	tf.Close()
	os.Remove(tmpFile)

	tmpFile = filepath.Join(os.TempDir(), "test2.m3u8")
	tf, err = os.Create(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	err = p.MarshalToWithErr(tf)
	if err != nil {
		t.Error("should not have write error")
	}
	tf.Close()
	os.Remove(tmpFile)
}

func TestResolveURL(t *testing.T) {
	t.Log(m3u8.ResolveURL("/unix/path/to/index.m3u8", "a.ts"))
	t.Log(m3u8.ResolveURL("http://host/path/to/index.m3u8", "a.ts"))

	// special cases
	t.Log(m3u8.ResolveURL("http://host/path/to/index.m3u8", "/home/a.ts"))
	t.Log(m3u8.ResolveURL("/unix/path/to/index.m3u8", "/home/a.ts"))
	t.Log(m3u8.ResolveURL("http://host/path/to/index.m3u8", "http://host/path/to/a.ts"))
}

func TestErrors(t *testing.T) {
	errCases := []struct {
		file string
		err  error
	}{
		{
			file: "testdata/no-extinf.m3u8",
			err:  m3u8.ErrMissingInf,
		},
		{
			file: "testdata/illegal.m3u8",
			err:  m3u8.ErrMissingInf,
		},
	}

	for _, c := range errCases {
		f, err := os.Open(c.file)
		if err != nil {
			t.Fatal("open test file", err)
		}
		defer f.Close()

		_, err = m3u8.Decode(f)
		if errors.Is(err, c.err) {
			t.Log("found error:", err)
		} else {
			t.Error("should return error missing extinf")
		}
	}
}

func TestIsMediaPlaylistTag(t *testing.T) {
	line := "#EXT-X-VERSION:3"
	if !m3u8.IsMediaPlaylistTag(line) {
		t.Error(line + " is media playlist tag")
	}
}

func TestMaster(t *testing.T) {
	f, err := os.Open("testdata/alt-videos.m3u8")
	if err != nil {
		t.Fatal("open test file", err)
	}
	defer f.Close()

	m, err := m3u8.DecodeMaster(f)
	if err != nil {
		t.Fatal("should try to decode a master")
	}
	t.Log(m.Marshal())

	f, err = os.Open("testdata/sample.m3u8")
	if err != nil {
		t.Fatal("open test file", err)
	}
	defer f.Close()

	_, err = m3u8.DecodeMaster(f)
	if err != m3u8.ErrMissingStream {
		t.Fatal("should be ErrMissingStream")
	}
}

func TestTryDecodeMaster(t *testing.T) {
	f, err := os.Open("testdata/master.m3u8")
	if err != nil {
		t.Fatal("open test file", err)
	}
	defer f.Close()

	m, err := m3u8.TryDecodeMaster(f)
	if err != nil {
		t.Fatal("should try to decode a master")
	}
	t.Log(m)

	f, err = os.Open("testdata/sample.m3u8")
	if err != nil {
		t.Fatal("open test file 2", err)
	}
	defer f.Close()

	_, err = m3u8.TryDecodeMaster(f)
	if err != m3u8.ErrNotMaster {
		t.Fatal("should report no master")
	}
}

func TestMoreM3U8s(t *testing.T) {
	files := []string{
		"testdata/blank-lines.m3u8",
		"testdata/cmaf-byterange.m3u8",
	}

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			t.Fatal("open test file", file, err)
		}
		defer f.Close()

		p, err := m3u8.Decode(f)
		if err != nil {
			t.Fatal("decode m3u8", file, err)
		}

		out := p.Marshal()
		// t.Log(out)

		_, err = m3u8.Decode(bytes.NewBuffer([]byte(out)))
		if err != nil {
			t.Fatal("marshal output should be decode right:", file, err)
		}
	}
}

func TestSplitAttributeList(t *testing.T) {
	l, err := m3u8.SplitAttributeList(`URI="init.mp4",BYTERANGE="596@0"`)
	if err != nil {
		t.Fatal(err)
	}
	if len(l) != 2 || l["URI"] != "init.mp4" || l["BYTERANGE"] != "596@0" {
		t.Fatal("wrong split")
	}

	l, err = m3u8.SplitAttributeList(`TYPE=AUDIO,URI="audio,with-comma.mp4",BANDWIDTH=1280000`)
	if err != nil {
		t.Fatal(err)
	}
	if len(l) != 3 || l["TYPE"] != "AUDIO" || l["URI"] != "audio,with-comma.mp4" ||
		l["BANDWIDTH"] != "1280000" {
		t.Logf("%#+v", l)
		t.Fatal("wrong split 2")
	}
}
