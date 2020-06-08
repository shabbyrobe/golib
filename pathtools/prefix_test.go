// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This is ripped from dep: https://raw.githubusercontent.com/golang/dep/master/internal/fs/fs.go

package pathtools

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// This function tests HasFilepathPrefix. It should test it on both case
// sensitive and insensitive situations. However, the only reliable way to test
// case-insensitive behaviour is if using case-insensitive filesystem.  This
// cannot be guaranteed in an automated test. Therefore, the behaviour of the
// tests is not to test case sensitivity on *nix and to assume that Windows is
// case-insensitive. Please see link below for some background.
//
// https://superuser.com/questions/266110/how-do-you-make-windows-7-fully-case-sensitive-with-respect-to-the-filesystem
//
// NOTE: NTFS can be made case-sensitive. However many Windows programs,
// including Windows Explorer do not handle gracefully multiple files that
// differ only in capitalization. It is possible that this can cause these tests
// to fail on some setups.
func TestHasFilepathPrefix(t *testing.T) {
	dir, err := ioutil.TempDir("", "dep")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// dir2 is the same as dir but with different capitalization on Windows to
	// test case insensitivity
	var dir2 string
	if runtime.GOOS == "windows" {
		dir = strings.ToLower(dir)
		dir2 = strings.ToUpper(dir)
	} else {
		dir2 = dir
	}

	// For testing trailing and repeated separators
	sep := string(os.PathSeparator)

	cases := []struct {
		path   string
		prefix string
		want   bool
	}{
		{filepath.Join(dir, "a", "b"), filepath.Join(dir2), true},
		{filepath.Join(dir, "a", "b"), dir2 + sep + sep + "a", true},
		{filepath.Join(dir, "a", "b"), filepath.Join(dir2, "a") + sep, true},
		{filepath.Join(dir, "a", "b") + sep, filepath.Join(dir2), true},
		{dir + sep + sep + filepath.Join("a", "b"), filepath.Join(dir2, "a"), true},
		{filepath.Join(dir, "a", "b"), filepath.Join(dir2, "a"), true},
		{filepath.Join(dir, "a", "b"), filepath.Join(dir2, "a", "b"), true},
		{filepath.Join(dir, "a", "b"), filepath.Join(dir2, "c"), false},
		{filepath.Join(dir, "a", "b"), filepath.Join(dir2, "a", "d", "b"), false},
		{filepath.Join(dir, "a", "b"), filepath.Join(dir2, "a", "b2"), false},
		{filepath.Join(dir), filepath.Join(dir2, "a", "b"), false},
		{filepath.Join(dir, "ab"), filepath.Join(dir2, "a", "b"), false},
		{filepath.Join(dir, "ab"), filepath.Join(dir2, "a"), false},
		{filepath.Join(dir, "123"), filepath.Join(dir2, "123"), true},
		{filepath.Join(dir, "123"), filepath.Join(dir2, "1"), false},
		{filepath.Join(dir, "⌘"), filepath.Join(dir2, "⌘"), true},
		{filepath.Join(dir, "a"), filepath.Join(dir2, "⌘"), false},
		{filepath.Join(dir, "⌘"), filepath.Join(dir2, "a"), false},
	}

	for _, c := range cases {
		if err := os.MkdirAll(c.path, 0755); err != nil {
			t.Fatal(err)
		}

		if err = os.MkdirAll(c.prefix, 0755); err != nil {
			t.Fatal(err)
		}

		got, _, _, err := FilepathPrefix(c.path, c.prefix)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if c.want != got {
			t.Fatalf("dir: %q, prefix: %q, expected: %v, got: %v", c.path, c.prefix, c.want, got)
		}
	}
}

// This function tests HadFilepathPrefix. It should test it on both case
// sensitive and insensitive situations. However, the only reliable way to test
// case-insensitive behaviour is if using case-insensitive filesystem.  This
// cannot be guaranteed in an automated test. Therefore, the behaviour of the
// tests is not to test case sensitivity on *nix and to assume that Windows is
// case-insensitive. Please see link below for some background.
//
// https://superuser.com/questions/266110/how-do-you-make-windows-7-fully-case-sensitive-with-respect-to-the-filesystem
//
// NOTE: NTFS can be made case-sensitive. However many Windows programs,
// including Windows Explorer do not handle gracefully multiple files that
// differ only in capitalization. It is possible that this can cause these tests
// to fail on some setups.
func TestHasFilepathPrefix_Files(t *testing.T) {
	dir, err := ioutil.TempDir("", "dep")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// dir2 is the same as dir but with different capitalization on Windows to
	// test case insensitivity
	var dir2 string
	if runtime.GOOS == "windows" {
		dir = strings.ToLower(dir)
		dir2 = strings.ToUpper(dir)
	} else {
		dir2 = dir
	}

	existingFile := filepath.Join(dir, "exists")
	if err = os.MkdirAll(existingFile, 0755); err != nil {
		t.Fatal(err)
	}

	nonExistingFile := filepath.Join(dir, "does_not_exists")

	cases := []struct {
		path   string
		prefix string
		want   bool
		err    bool
	}{
		{existingFile, filepath.Join(dir2), true, false},

		// BW: This differs from the dep behaviour. We actually want to check
		// prefixes for things that don't exist. Existence should be the job
		// of the caller.
		{nonExistingFile, filepath.Join(dir2), true, false},
	}

	for _, c := range cases {
		got, _, _, err := FilepathPrefix(c.path, c.prefix)
		if err != nil && !c.err {
			t.Fatalf("unexpected error: %s", err)
		}
		if c.want != got {
			t.Fatalf("dir: %q, prefix: %q, expected: %v, got: %v", c.path, c.prefix, c.want, got)
		}
	}
}

// This function tests FilepathPrefix. See TestHasFilePrefix for caveats.
func TestFilepathPrefix(t *testing.T) {
	dir, err := ioutil.TempDir("", "dep")
	if err != nil {
		t.Fatal(err)
	}

	const (
		createNone = 0
		createFile = 1
		createDir  = 2
	)

	cases := []struct {
		want   bool
		path   string
		prefix string
		suffix string
		create int
	}{
		{true, filepath.Join(dir), filepath.Join(dir), "", createDir},
		{true, filepath.Join(dir, "a"), filepath.Join(dir), "a/", createDir},
		{true, filepath.Join(dir, "a"), filepath.Join(dir), "a", createNone},
		{true, filepath.Join(dir, "a", "b"), filepath.Join(dir), "a/b/", createDir},
		{true, filepath.Join(dir, "a", "b"), filepath.Join(dir), "a/b", createNone},
		{true, filepath.Join(dir, "a", "b"), filepath.Join(dir), "a/b", createFile},
		{true, filepath.Join(dir, "a", "b"), filepath.Join(dir, "a"), "b/", createDir},
		{true, filepath.Join(dir, "a", "b") + string(os.PathSeparator), filepath.Join(dir), "a/b/", createDir},
		{true, filepath.Join(dir, "a", "b") + string(os.PathSeparator), filepath.Join(dir), "a/b/", createNone},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)

			if c.create == createDir {
				if err := os.MkdirAll(c.path, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(c.prefix, 0755); err != nil {
					t.Fatal(err)
				}
			}
			if c.create == createFile {
				if err := os.MkdirAll(filepath.Dir(c.path), 0755); err != nil {
					t.Fatal(err)
				}
				f, err := os.Create(c.path)
				if err != nil {
					t.Fatal(err)
				}
				if err := f.Close(); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(c.prefix, 0755); err != nil {
					t.Fatal(err)
				}
			}

			got, p, s, err := FilepathPrefix(c.path, c.prefix)
			if err != nil {
				t.Fatal(err)
			}
			if c.want != got {
				t.Fatal()
			}
			if c.prefix != p {
				t.Fatal()
			}
			if c.suffix != s {
				t.Fatal()
			}
		})
	}
}
