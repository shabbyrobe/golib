// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is taken from dep, with some minor modifications:
// https://raw.githubusercontent.com/golang/dep/master/internal/fs/fs.go
//
// In order to satisfy a more generalised use-case, we don't ensure the
// existence of prefix or path and leave that up to the caller.
//
// We also return the matched prefix portion regardless of whether it is a full
// match or not.

package pathtools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// HasFilepathPrefix will determine if "path" starts with "prefix" from
// the point of view of a filesystem.
//
// Unlike filepath.HasPrefix, this function is path-aware, meaning that
// it knows that two directories /foo and /foobar are not the same
// thing, and therefore HasFilepathPrefix("/foobar", "/foo") will return
// false.
//
// This function also handles the case where the involved filesystems
// are case-insensitive, meaning /foo/bar and /Foo/Bar correspond to the
// same file. In that situation HasFilepathPrefix("/Foo/Bar", "/foo")
// will return true. The implementation is *not* OS-specific, so a FAT32
// filesystem mounted on Linux will be handled correctly.
//
// Unlike the dep code this was cribbed from, path is not required to exist,
// but prefix still must exist. A future enhancement should require only
// one component of prefix with a letter in it to exist as this is adequate to
// determine case sensitivity.
//
func HasFilepathPrefix(path, prefix string) (ok bool, err error) {
	ok, _, _, err = FilepathPrefix(path, prefix)
	return ok, err
}

func FilepathPrefix(path, prefix string) (ok bool, matched string, left string, err error) {
	// this function is more convoluted then ideal due to need for special
	// handling of volume name/drive letter on Windows. vnPath and vnPrefix
	// are first compared, and then used to initialize initial values of p and
	// d which will be appended to for incremental checks using
	// IsCaseSensitiveFilesystem and then equality.

	// no need to check IsCaseSensitiveFilesystem because VolumeName return
	// empty string on all non-Windows machines
	vnPath := strings.ToLower(filepath.VolumeName(path))
	vnPrefix := strings.ToLower(filepath.VolumeName(prefix))
	if vnPath != vnPrefix {
		return false, "", "", nil
	}

	// Because filepath.Join("c:","dir") returns "c:dir", we have to manually
	// add path separator to drive letters. Also, we need to set the path root
	// on *nix systems, since filepath.Join("", "dir") returns a relative path.
	vnPath += string(os.PathSeparator)
	vnPrefix += string(os.PathSeparator)

	var dn, fn string

	isDir, err := isDir(path)
	if err != nil {
		isDir = path[len(path)-1] == os.PathSeparator
	}
	if isDir {
		dn = path
	} else {
		dn, fn = filepath.Split(path)
	}

	dn = filepath.Clean(dn)
	prefix = filepath.Clean(prefix)

	// [1:] in the lines below eliminates empty string on *nix and volume name on Windows
	dirs := strings.Split(dn, string(os.PathSeparator))[1:]
	prefixes := strings.Split(prefix, string(os.PathSeparator))[1:]

	if len(prefixes) > len(dirs) {
		return false, "", "", nil
	}

	// d,p are initialized with "/" on *nix and volume name on Windows
	d := vnPath
	p := vnPrefix
	matched = vnPrefix

	rem := func(in []string) string {
		if !isDir {
			in = append(in, fn)
		}
		out := strings.Join(in, string(os.PathSeparator))
		if isDir && out != "" {
			out += string(os.PathSeparator)
		}
		return out
	}

	for i := range prefixes {
		// need to test each component of the path for
		// case-sensitiveness because on Unix we could have
		// something like ext4 filesystem mounted on FAT
		// mountpoint, mounted on ext4 filesystem, i.e. the
		// problematic filesystem is not the last one.
		caseSensitive, err := isCaseSensitiveFilesystem(filepath.Join(d, dirs[i]))
		if err != nil {
			return false, matched, strings.Join(dirs[i:], string(os.PathSeparator)),
				fmt.Errorf("pathtools: failed to check filepath prefix. error: %v", err)
		}

		if caseSensitive {
			d = filepath.Join(d, dirs[i])
			p = filepath.Join(p, prefixes[i])
		} else {
			d = filepath.Join(d, strings.ToLower(dirs[i]))
			p = filepath.Join(p, strings.ToLower(prefixes[i]))
		}

		if p != d {
			return false, matched, rem(dirs[i:]), nil
		}
		matched = filepath.Join(matched, prefixes[i])
	}

	return true, matched, rem(dirs[len(prefixes):]), nil
}

func isDir(name string) (bool, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return false, err
	}
	if !fi.IsDir() {
		return false, fmt.Errorf("%q is not a directory", name)
	}
	return true, nil
}

// EquivalentPaths compares the paths passed to check if they are equivalent.
// It respects the case-sensitivity of the underlying filesysyems.
func equivalentPaths(p1, p2 string) (bool, error) {
	p1 = filepath.Clean(p1)
	p2 = filepath.Clean(p2)

	fi1, err := os.Stat(p1)
	if err != nil {
		return false, fmt.Errorf("pathtools: could not check for path equivalence. error: %v", err)
	}
	fi2, err := os.Stat(p2)
	if err != nil {
		return false, fmt.Errorf("pathtools: could not check for path equivalence. error: %v", err)
	}

	p1Filename, p2Filename := "", ""

	if !fi1.IsDir() {
		p1, p1Filename = filepath.Split(p1)
	}
	if !fi2.IsDir() {
		p2, p2Filename = filepath.Split(p2)
	}

	if isPrefix1, _, _, err := FilepathPrefix(p1, p2); err != nil {
		return false, fmt.Errorf("pathtools: could not check for path equivalence. error: %v", err)
	} else if isPrefix2, _, _, err := FilepathPrefix(p2, p1); err != nil {
		return false, fmt.Errorf("pathtools: could not check for path equivalence. error: %v", err)
	} else if !isPrefix1 || !isPrefix2 {
		return false, nil
	}

	if p1Filename != "" || p2Filename != "" {
		caseSensitive, err := isCaseSensitiveFilesystem(filepath.Join(p1, p1Filename))
		if err != nil {
			return false, fmt.Errorf("pathtools: could not check for filesystem case-sensitivity. error: %v", err)
		}
		if caseSensitive {
			if p1Filename != p2Filename {
				return false, nil
			}
		} else {
			if strings.ToLower(p1Filename) != strings.ToLower(p2Filename) {
				return false, nil
			}
		}
	}

	return true, nil
}

// isCaseSensitiveFilesystem determines if the filesystem where dir
// exists is case sensitive or not.
//
// CAVEAT: this function works by taking the last component of the given
// path and flipping the case of the first letter for which case
// flipping is a reversible operation (/foo/Bar → /foo/bar), then
// testing for the existence of the new filename. There are two
// possibilities:
//
// 1. The alternate filename does not exist. We can conclude that the
// filesystem is case sensitive.
//
// 2. The filename happens to exist. We have to test if the two files
// are the same file (case insensitive file system) or different ones
// (case sensitive filesystem).
//
// If the input directory is such that the last component is composed
// exclusively of case-less codepoints (e.g.  numbers), this function will
// return false.
func isCaseSensitiveFilesystem(dir string) (bool, error) {
	alt := filepath.Join(filepath.Dir(dir), genTestFilename(filepath.Base(dir)))

	dInfo, err := os.Stat(dir)
	if err != nil {
		return false, fmt.Errorf("pathtools: could not determine the case-sensitivity of the filesystem. error: %v", err)
	}

	aInfo, err := os.Stat(alt)
	if err != nil {
		// If the file doesn't exists, assume we are on a case-sensitive filesystem.
		if os.IsNotExist(err) {
			return true, nil
		}

		return false, fmt.Errorf("pathtools: could not determine the case-sensitivity of the filesystem. error: %v", err)
	}

	return !os.SameFile(dInfo, aInfo), nil
}

// genTestFilename returns a string with at most one rune case-flipped.
//
// The transformation is applied only to the first rune that can be
// reversibly case-flipped, meaning:
//
// * A lowercase rune for which it's true that lower(upper(r)) == r
// * An uppercase rune for which it's true that upper(lower(r)) == r
//
// All the other runes are left intact.
func genTestFilename(str string) string {
	flip := true
	return strings.Map(func(r rune) rune {
		if flip {
			if unicode.IsLower(r) {
				u := unicode.ToUpper(r)
				if unicode.ToLower(u) == r {
					r = u
					flip = false
				}
			} else if unicode.IsUpper(r) {
				l := unicode.ToLower(r)
				if unicode.ToUpper(l) == r {
					r = l
					flip = false
				}
			}
		}
		return r
	}, str)
}
