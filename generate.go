// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/shurcooL/go-goon"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var outputFile = flag.String("o", "", "write output to `file` (default standard output)")

func run() error {
	// TODO: Use tagged release starting with v4.4.0 when it's out, instead of master.
	resp, err := http.Get("https://raw.githubusercontent.com/primer/octicons/master/build/svg.json")
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 status code: %v", resp.StatusCode)
	}

	var v map[string]string
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return err
	}

	var ss []string
	for k := range v {
		ss = append(ss, k)
	}
	sort.Strings(ss)

	var buf bytes.Buffer
	fmt.Fprint(&buf, `package octicons

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
`)
	for i := range ss {
		processOcticon(&buf, v, ss, i)
	}
	fmt.Fprint(&buf, ")\n")

	b, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	var w io.Writer
	switch *outputFile {
	case "":
		w = os.Stdout
	default:
		f, err := os.Create(*outputFile)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	_, err = w.Write(b)
	return err
}

func processOcticon(w io.Writer, v map[string]string, ss []string, i int) {
	svg := parseOcticon(v[ss[i]])

	parent := svg
	child := svg.FirstChild

	parent.FirstChild = nil
	parent.LastChild = nil
	child.Parent = nil

	fmt.Fprintf(w, "	// %s is an %q Octicon SVG node.\n", dashSepToMixedCaps(ss[i]), ss[i])
	fmt.Fprintf(w, "	%s = func() *html.Node {\n", dashSepToMixedCaps(ss[i]))

	fmt.Fprint(w, "		parent := ")
	goon.Fdump(w, parent)
	fmt.Fprint(w, "		child := ")
	goon.Fdump(w, child)

	fmt.Fprint(w, `		parent.FirstChild = child
		parent.LastChild = child
		child.Parent = parent
		return parent
	}()

`)
}

func main() {
	flag.Parse()

	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}

func parseOcticon(svgXML string) *html.Node {
	e, err := html.ParseFragment(strings.NewReader(svgXML), nil)
	if err != nil {
		panic(fmt.Errorf("internal error: html.Parse failed: %v", err))
	}
	svg := e[0].LastChild.FirstChild // TODO: Is there a better way to just get the <svg>...</svg> element directly, skipping <html><head></head><body><svg>...</svg></body></html>?
	svg.Parent.RemoveChild(svg)
	for i, attr := range svg.Attr {
		if attr.Namespace == "" && attr.Key == "width" {
			svg.Attr[i].Val = "16"
			break
		}
	}
	svg.Attr = append(svg.Attr, html.Attribute{Key: atom.Style.String(), Val: `vertical-align: top;`})
	return svg
}

// dashSepToMixedCaps converts "string-URL-append" to "StringURLAppend" form.
func dashSepToMixedCaps(in string) string {
	var out string
	ss := strings.Split(in, "-")
	for _, s := range ss {
		initialism := strings.ToUpper(s)
		if _, ok := initialisms[initialism]; ok {
			out += initialism
		} else {
			out += strings.Title(s)
		}
	}
	return out
}

// initialisms is the set of initialisms in Go-style Mixed Caps case.
var initialisms = map[string]struct{}{
	"API":   {},
	"ASCII": {},
	"CPU":   {},
	"CSS":   {},
	"DNS":   {},
	"EOF":   {},
	"GUID":  {},
	"HTML":  {},
	"HTTP":  {},
	"HTTPS": {},
	"ID":    {},
	"IP":    {},
	"JSON":  {},
	"LHS":   {},
	"QPS":   {},
	"RAM":   {},
	"RHS":   {},
	"RPC":   {},
	"SLA":   {},
	"SMTP":  {},
	"SQL":   {},
	"SSH":   {},
	"TCP":   {},
	"TLS":   {},
	"TTL":   {},
	"UDP":   {},
	"UI":    {},
	"UID":   {},
	"UUID":  {},
	"URI":   {},
	"URL":   {},
	"UTF8":  {},
	"VM":    {},
	"XML":   {},
	"XSRF":  {},
	"XSS":   {},

	"RSS": {},
}
