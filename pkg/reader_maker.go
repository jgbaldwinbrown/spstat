package spstat

import (
	"bytes"
	"os"
	"strings"
	"io"
	"github.com/jgbaldwinbrown/csvh"
)

// A reusable object that opens an io.ReadCloser; widely used here to open
// tab-separated tables for reading.
type ReadCloserMaker interface {
	NewReadCloser() (io.ReadCloser, error)
}

// A path to open uncompressed.
type Path string

func (p Path) NewReadCloser() (io.ReadCloser, error) {
	return os.Open(string(p))
}

// A path to open uncompressed, or compressed if suffix is '.gz'
type MaybeGzPath string

func (p MaybeGzPath) NewReadCloser() (io.ReadCloser, error) {
	return csvh.OpenMaybeGz(string(p))
}

// Wrapper that adds a no-op Close() method to an io.Reader
type ReadCloser struct {
	io.Reader
}

func (r ReadCloser) Close() error {
	return nil
}

// A string with a NewReadCloser() method
type String string

func (s String) NewReadCloser() (io.ReadCloser, error) {
	return ReadCloser{strings.NewReader(string(s))}, nil
}

// A byte slice with a NewReadCloser() method
type Bytes []byte

func (b Bytes) NewReadCloser() (io.ReadCloser, error) {
	return ReadCloser{bytes.NewReader(([]byte)(b))}, nil
}
