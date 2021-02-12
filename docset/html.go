package docset

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TOCNode(kind EntryKind, name string) *html.Node {
	anchor := fmt.Sprintf("//apple_ref/%s/%s", kind, url.PathEscape(name))
	tocNode := html.Node{
		DataAtom: atom.A,
		Data:     "a",
		Attr: []html.Attribute{
			{Key: "name", Val: anchor},
			{Key: "class", Val: "dashAnchor"},
		},
	}
	return &tocNode
}

func AddTOCNode(into *html.Node, kind EntryKind, name string) *html.Node {
	into.InsertBefore(TOCNode(kind, name), into.FirstChild)
	return into
}

func HeadingOutlineTOC(into *html.Node) *html.Node {
	var walk func(node *html.Node)
	walk = func(node *html.Node) {
		if node.DataAtom == atom.H1 ||
			node.DataAtom == atom.H2 ||
			node.DataAtom == atom.H3 ||
			node.DataAtom == atom.H4 ||
			node.DataAtom == atom.H5 ||
			node.DataAtom == atom.H6 {

			AddTOCNode(into, Section, allText(node))
		}

		for e := node.FirstChild; e != nil; e = e.NextSibling {
			walk(e)
		}
	}

	return into
}

var multiSpace = regexp.MustCompile(`[ \t\n\r]+`)

func allText(node *html.Node) string {
	var into strings.Builder
	if node == nil {
		return ""
	}

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			into.WriteString(n.Data)
		}
		if n.FirstChild != nil {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}
	f(node)

	return strings.TrimSpace(multiSpace.ReplaceAllString(into.String(), " "))
}
