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

	firstTags := `#EXT-X-CUSTOM:some-key-values
#EXTINF:10.000000,
#EXT-X-TAG:VALUE=5954336
00001.ts
`
	first := p.Front()
	if first != nil {
		if first.Marshal() != firstTags {
			t.Log(first.Marshal())
			t.Error("should retains exact tags in order")
		}
	} else {
		t.Fatal("fail to decode")
	}

	second := first.Next()
	if second != nil {
		t.Log(second.Directives)
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
	f, err := os.Open("testdata/no-extinf.m3u8")
	if err != nil {
		t.Fatal("open test file", err)
	}
	defer f.Close()

	_, err = m3u8.Decode(f)
	if errors.Is(err, m3u8.ErrMissingInf) {
		t.Log("found error:", err)
	} else {
		t.Error("should return error missing extinf")
	}
}

func TestBlankLines(t *testing.T) {
	f, err := os.Open("testdata/blank-lines.m3u8")
	if err != nil {
		t.Fatal("open test file", err)
	}
	defer f.Close()

	p, err := m3u8.Decode(f)

	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(p.Marshal())
	}
}

func TestDecodeStrict(t *testing.T) {
	f, err := os.Open("testdata/illegal.m3u8")
	if err != nil {
		t.Fatal("open test file", err)
	}
	defer f.Close()

	_, err = m3u8.Decode(f)
	if errors.Is(err, m3u8.ErrUnknownLine) {
		t.Log("found error if strict:", err)
	} else {
		t.Error("should return error unknown-line")
	}

	f2, err := os.Open("testdata/illegal.m3u8")
	if err != nil {
		t.Fatal("open test file", err)
	}
	defer f2.Close()
	pl, err := m3u8.DecodeStrict(f2, false)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(pl.Marshal())
	}
}
