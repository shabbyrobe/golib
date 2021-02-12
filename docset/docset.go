package docset

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"golang.org/x/net/html"
)

type IndexEntry struct {
	Kind EntryKind
	Name string
}

type Docset struct {
	Ident        string
	Name         string
	Path         string
	DocPath      string
	DbFile       string
	ContentsPath string
	IndexFile    string
	PlistFile    string
	AllowJs      bool
	AllowOnline  bool

	LinkRewriter func(link string) (string, error)

	// The keyword Dash uses for Google/Stack Overflow searches for the docset.
	WebSearchKeyword string

	// The default user-set keyword for docsets is the DocSetPlatformFamily. Defaults to
	// ident.
	PlatformFamily string

	Db   *sqlite.Conn
	stmt *sqlite.Stmt

	docs map[string]document
}

var identPattern = regexp.MustCompile("^[A-Za-z0-9]+$")

func sanitiseName(name string) string {
	return regexp.MustCompile("[^A-Za-z0-9]").ReplaceAllString(name, "_")
}

func New(base, ident, name string) (*Docset, error) {
	if !identPattern.MatchString(ident) {
		return nil, fmt.Errorf("docset: invalid ident: %s", ident)
	}

	path := filepath.Join(base, sanitiseName(name)+".docset")

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil, fmt.Errorf("docset: path exists")
	}

	ds := &Docset{
		Path:           path,
		ContentsPath:   filepath.Join(path, "Contents"),
		DbFile:         filepath.Join(path, "Contents", "Resources", "docSet.dsidx"),
		DocPath:        filepath.Join(path, "Contents", "Resources", "Documents"),
		PlistFile:      filepath.Join(path, "Contents", "Info.plist"),
		IndexFile:      "index.html",
		PlatformFamily: ident,

		docs: map[string]document{},
	}

	if err := os.MkdirAll(ds.ContentsPath, 0700); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(ds.DocPath, 0700); err != nil {
		return nil, err
	}

	conn, err := sqlite.OpenConn(ds.DbFile, sqlite.SQLITE_OPEN_CREATE|sqlite.SQLITE_OPEN_READWRITE)
	if err != nil {
		return nil, err
	}
	ds.Db = conn

	if err := sqlitex.ExecScript(conn, createSql); err != nil {
		ds.Close()
		return nil, err
	}

	ds.stmt, _, err = ds.Db.PrepareTransient(insertSql)
	if err != nil {
		ds.Close()
		return nil, err
	}

	return ds, nil
}

func (d *Docset) indexFile(indexes []IndexEntry, file string) (rerr error) {
	defer sqlitex.Save(d.Db)(&rerr)
	for _, entry := range indexes {
		if err := d.stmt.Reset(); err != nil {
			return err
		}
		d.stmt.SetText("name", entry.Name)
		d.stmt.SetText("type", string(entry.Kind))
		d.stmt.SetText("path", file)
		if _, err := d.stmt.Step(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Docset) AddHTML(source string, node *html.Node, indexes ...IndexEntry) (Document, error) {
	if _, ok := d.docs[source]; ok {
		return nil, fmt.Errorf("docset: duplicate document added")
	}
	file := ensureExt(fsSafeName(source), ".html")
	doc := &htmlDocument{
		docset: d,
		source: source,
		file:   file,
		node:   node,
	}
	d.docs[source] = doc
	if err := d.indexFile(indexes, file); err != nil {
		return nil, err
	}
	return doc, nil
}

func (d *Docset) AddRaw(source string, data []byte, indexes ...IndexEntry) (Document, error) {
	if _, ok := d.docs[source]; ok {
		return nil, fmt.Errorf("docset: duplicate document added")
	}
	file := ensureExt(fsSafeName(source), ".html")
	doc := &rawDocument{
		docset: d,
		source: source,
		file:   file,
		data:   data,
	}
	d.docs[source] = doc
	if err := d.indexFile(indexes, file); err != nil {
		return nil, err
	}
	return doc, nil
}

func (d *Docset) flush() (rerr error) {
	for _, doc := range d.docs {
		if err := doc.Resolve(); err != nil {
			return fmt.Errorf("could not resolve %s: %w", doc.Source(), err)
		}
	}

	for _, doc := range d.docs {
		bts, err := doc.Data()
		if err != nil {
			return fmt.Errorf("could not render %s: %w", doc.Source(), err)
		}

		outPath := filepath.Join(d.DocPath, doc.File())
		if err := ioutil.WriteFile(outPath, bts, 0600); err != nil {
			return err
		}
	}

	var buf bytes.Buffer
	if err := renderPlist(&buf, d); err != nil {
		return err
	}
	if err := ioutil.WriteFile(d.PlistFile, buf.Bytes(), 0600); err != nil {
		return err
	}

	return nil
}

func (d *Docset) DeferClose() func(rerr *error) {
	return func(rerr *error) {
		if err := d.close(*rerr); err != nil && *rerr == nil {
			*rerr = err
		}
	}
}

func (d *Docset) Close() (rerr error) {
	return d.close(nil)
}

func (d *Docset) close(err error) (rerr error) {
	if err == nil {
		rerr = d.flush()
	}
	if err := d.stmt.Finalize(); err != nil && rerr == nil {
		rerr = err
	}
	if err := d.Db.Close(); err != nil && rerr == nil {
		rerr = err
	}
	return rerr
}

func attr(node *html.Node, name string) (v *html.Attribute) {
	if node == nil {
		return nil
	}
	for idx, attr := range node.Attr {
		if attr.Key == name {
			return &node.Attr[idx]
		}
	}
	return nil
}

var unsafeCharReplace = regexp.MustCompile("[^A-Za-z0-9_-]")

func fsSafeName(in string) string {
	hash := md5.Sum([]byte(in))

	fname := in
	fname = strings.Replace(fname, ".", "_", -1)
	fname = strings.Replace(fname, "/", "_", -1)
	fname = unsafeCharReplace.ReplaceAllString(fname, "")

	return fname + "_" + hex.EncodeToString(hash[:])
}

func ensureExt(name string, ext string) string {
	if filepath.Ext(name) != ext {
		return name + ext
	}
	return name
}
