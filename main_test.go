package main

import (
	"bytes"
	"testing"

	"golang.org/x/tools/txtar"
)

func Test_htmlFmt(t *testing.T) {
	a, err := txtar.ParseFile("testdata/htmls.txt")
	if err != nil {
		t.Fatal(err)
	}

	n := len(a.Files)
	if n%2 != 0 {
		if err != nil {
			t.Fatalf("got %d test files; want an even number (pairs)", n)
		}
	}

	for i := 0; i < n-1; i += 2 {
		testF, wantF := a.Files[i], a.Files[i+1]
		if wantF.Name != "want" {
			if err != nil {
				t.Fatalf("want-file named %s; want \"want\"", wantF.Name)
			}
		}

		t.Run(testF.Name, func(t *testing.T) {
			data := trim(testF.Data)
			want := trim(wantF.Data)

			got := htmlFmt(bytes.NewReader(data))

			if got = trim(got); !bytes.Equal(got, want) {
				t.Errorf("\n for: %q\n got: %q\nwant: %q", data, got, want)
			}
		})
	}
}

var trim = bytes.TrimSpace
