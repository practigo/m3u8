package m3u8_test

import (
	"os"
	"testing"

	"github.com/practigo/m3u8"
)

func TestM3U8(t *testing.T) {
	f, err := os.Open("testdata/sample.m3u8")
	if err != nil {
		t.Fatal("open test file", err)
	}

	p, err := m3u8.Decode(f)
	if err != nil {
		t.Fatal("decode m3u8", err)
	}

	first := p.Front()
	if first != nil {
		t.Log(first.Marshal())
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
	})
	p.Remove(second)

	t.Log(p.Marshal())
}

func TestResolveURL(t *testing.T) {
	t.Log(m3u8.ResolveURL("/unix/path/to/index.m3u8", "a.ts"))
	t.Log(m3u8.ResolveURL("http://host/path/to/index.m3u8", "a.ts"))

	// special cases
	t.Log(m3u8.ResolveURL("http://host/path/to/index.m3u8", "/home/a.ts"))
	t.Log(m3u8.ResolveURL("/unix/path/to/index.m3u8", "/home/a.ts"))
	t.Log(m3u8.ResolveURL("http://host/path/to/index.m3u8", "http://host/path/to/a.ts"))
}
