// package gio, implements io helper functions
// which are not available in stdlib
package gio

import (
	"errors"
	"io"
)

// // CopyBufferAt , copies `src` to `dst` at `off` using `buf`
// //
// // copied from io.copyBuffer
// func CopyBufferAt(dst io.WriterAt, src io.Reader, off int64, buf []byte) (written int64, err error) {
// 	for {
// 		nr, er := src.Read(buf)
// 		if nr > 0 {
// 			nw, ew := dst.WriteAt(buf[0:nr], off+written)
// 			if nw < 0 || nr < nw {
// 				nw = 0
// 				if ew == nil {
// 					ew = errors.New("invalid write result")
// 				}
// 			}
// 			written += int64(nw)
// 			if ew != nil {
// 				err = ew
// 				break
// 			}
// 			if nr != nw {
// 				err = io.ErrShortWrite
// 				break
// 			}
// 		}
// 		if er != nil {
// 			if er != io.EOF {
// 				err = er
// 			}
// 			break
// 		}
// 	}

// 	return written, err
// }

// SectionWriter implements Read, Seek, and ReadAt on a section
// of an underlying WriterAt.
type SectionWriter struct {
	w      io.WriterAt
	base   int64
	off    int64
	limit  int64
	Done   func(from, to int64)
	Failed func(from, to int64)
}

// NewSectionWriter returns a SectionWriter that writes to r
// starting at offset off and stops with EOF after n bytes.
func NewSectionWriter(w io.WriterAt, off int64, limit int64, done, failed func(from, to int64)) *SectionWriter {
	return &SectionWriter{w, off, off, limit, done, failed}
}

func (s *SectionWriter) Write(p []byte) (n int, err error) {
	if s.off >= s.limit {
		return 0, io.EOF
	}
	if max := s.limit - s.off; int64(len(p)) > max {
		p = p[0:max]
	}
	n, err = s.w.WriteAt(p, s.off)
	if err != nil {
		s.Failed(s.off+int64(n), s.off+int64(len(p)))
	}
	s.Done(s.off, s.off+int64(n))

	s.off += int64(n)
	return
}

var errWhence = errors.New("Seek: invalid whence")
var errOffset = errors.New("Seek: invalid offset")

func (s *SectionWriter) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	default:
		return 0, errWhence
	case io.SeekStart:
		offset += s.base
	case io.SeekCurrent:
		offset += s.off
	case io.SeekEnd:
		offset += s.limit
	}
	if offset < s.base {
		return 0, errOffset
	}
	s.off = offset
	return offset - s.base, nil
}

func (s *SectionWriter) WriteAt(p []byte, off int64) (n int, err error) {
	if off < 0 || off >= s.limit-s.base {
		return 0, io.EOF
	}
	off += s.base
	if max := s.limit - off; int64(len(p)) > max {
		p = p[0:max]
		n, err = s.w.WriteAt(p, off)
		if err != nil {
			s.Failed(off+int64(n), off+int64(len(p)))
		}
		s.Done(off, off+int64(n))

		return n, err
	}

	n, err = s.w.WriteAt(p, off)
	if err != nil {
		s.Failed(off+int64(n), off+int64(len(p)))
	}
	s.Done(off, off+int64(n))

	return
}

// Size returns the size of the section in bytes.
func (s *SectionWriter) Size() int64 { return s.limit - s.base }
