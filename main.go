package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

var (
	wFlag = flag.Bool("w", false, "write formatted HTML back to input file(s)")
	exit  = 0
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: htmlfmt [[-w] file...]\nMultiple files assumes -w.\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	log.SetPrefix("htmlfmt: ")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	switch flag.NArg() {
	case 0:
		if *wFlag {
			log.Fatal("-w flag cannot be used without input file(s)")
		}
		convert(os.Stdin, "")
	case 1:
		path := flag.Arg(0)
		f, err := os.Open(path)
		if err != nil {
			log.Print(err)
			exit = 1
			break
		}
		convert(f, path)
	default:
		*wFlag = true
		for _, path := range flag.Args() {
			f, err := os.Open(path)
			if err != nil {
				log.Print(err)
				exit = 1
			}
			convert(f, path)
		}
	}

	os.Exit(exit)
}

func convert(r io.Reader, path string) {
	out := htmlFmt(r)
	switch *wFlag {
	case true:
		if err := os.WriteFile(path, out, 0666); err != nil {
			log.Print(err)
			exit = 1
			return
		}
	default:
		os.Stdout.Write(out)
	}
}

func htmlFmt(r io.Reader) []byte {
	indent := []byte(strings.Repeat("\t", 100))
	buf := &bytes.Buffer{}

	z := html.NewTokenizer(r)
	depth := 0
	for tt := z.Next(); tt != html.ErrorToken; tt = z.Next() {
		switch tt {
		case html.StartTagToken:
			buf.Write(indent[:depth])
			depth++

			name, hasAttr := z.TagName()
			buf.WriteByte('<')
			buf.Write(name)
			if hasAttr {
				buf.WriteString(sprintAttrs(z))
			}
			buf.WriteByte('>')

		case html.SelfClosingTagToken:
			buf.Write(indent[:depth])

			name, hasAttr := z.TagName()
			buf.WriteByte('<')
			buf.Write(name)
			if hasAttr {
				buf.WriteString(sprintAttrs(z))
			}
			buf.Write([]byte{'/', '>'})

		case html.TextToken:
			s := bytes.TrimSpace(z.Text())
			if len(s) == 0 {
				continue
			}

			buf.Write(indent[:depth])

			s = bytes.Join(bytes.Fields(s), []byte{' '})
			buf.Write(s)

		case html.EndTagToken:
			depth--
			buf.Write(indent[:depth])

			name, _ := z.TagName()
			buf.Write([]byte{'<', '/'})
			buf.Write(name)
			buf.WriteByte('>')
		}
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

func sprintAttrs(z *html.Tokenizer) string {
	attrs := []string{}
	var (
		k, v      []byte
		moreAttr  = true
		id, class string
	)
	for moreAttr {
		k, v, moreAttr = z.TagAttr()
		s := fmt.Sprintf("%s=\"%s\"", k, html.EscapeString(string(v)))
		switch {
		case strings.HasPrefix(s, "id"):
			id = s
		case strings.HasPrefix(s, "class"):
			class = s
		default:
			attrs = append(attrs, s)
		}
	}
	slices.Sort(attrs)
	// insert in reverse order
	if class != "" {
		attrs = slices.Insert(attrs, 0, class)
	}
	if id != "" {
		attrs = slices.Insert(attrs, 0, id)
	}
	return " " + strings.Join(attrs, " ")
}
