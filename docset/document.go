package docset

import (
	"bytes"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type Document interface {
	AddIndexes(...IndexEntry) error
}

type document interface {
	Document
	Source() string
	File() string
	Resolve() error
	Data() ([]byte, error)
}

type rawDocument struct {
	docset *Docset
	source string
	file   string
	data   []byte
}

var _ document = &rawDocument{}

func (h *rawDocument) Source() string        { return h.source }
func (h *rawDocument) File() string          { return h.file }
func (h *rawDocument) Resolve() (err error)  { return nil }
func (h *rawDocument) Data() ([]byte, error) { return h.data, nil }

func (h *rawDocument) AddIndexes(index ...IndexEntry) error {
	return h.docset.indexFile(index, h.source)
}

type htmlDocument struct {
	docset *Docset
	source string
	file   string
	node   *html.Node
}

var _ Document = &htmlDocument{}

func (h *htmlDocument) AddIndexes(index ...IndexEntry) error {
	return h.docset.indexFile(index, h.source)
}

func (h *htmlDocument) Source() string { return h.source }
func (h *htmlDocument) File() string   { return h.file }

func (h *htmlDocument) Resolve() (err error) {
	var walk func(n *html.Node) error
	walk = func(n *html.Node) error {
		for e := n.FirstChild; e != nil; e = e.NextSibling {
			if e.DataAtom == atom.A {
				hrefAttr := attr(e, "href")
				if hrefAttr != nil {
					href := hrefAttr.Val
					if h.docset.LinkRewriter != nil {
						href, err = h.docset.LinkRewriter(href)
						if err != nil {
							return err
						}
					}

					if h.docset.docs[href] != nil {
						hrefAttr.Val = h.docset.docs[href].File()
					}
				}
			}

			if err := walk(e); err != nil {
				return err
			}
		}
		return nil
	}

	return walk(h.node)
}

func (h *htmlDocument) Data() ([]byte, error) {
	var buf bytes.Buffer
	if err := html.Render(&buf, h.node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
