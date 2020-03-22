// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package bytescan

// Exported for testing only.

// ErrOrEOF is like Err, but returns EOF. Used to test a corner case.
func (s *Scanner) ErrOrEOF() error {
	return s.err
}
