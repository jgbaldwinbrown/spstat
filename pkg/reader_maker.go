package spstat

import (
	"bytes"
	"os"
	"strings"
	"io"
	"github.com/jgbaldwinbrown/csvh"
)

type ReadCloserMaker interface {
	NewReadCloser() (io.ReadCloser, error)
}

type Path string

func (p Path) NewReadCloser() (io.ReadCloser, error) {
	return os.Open(string(p))
}

type MaybeGzPath string

func (p MaybeGzPath) NewReadCloser() (io.ReadCloser, error) {
	return csvh.OpenMaybeGz(string(p))
}

type ReadCloser struct {
	io.Reader
}

func (r ReadCloser) Close() error {
	return nil
}

type String string

func (s String) NewReadCloser() (io.ReadCloser, error) {
	return ReadCloser{strings.NewReader(string(s))}, nil
}

type Bytes []byte

func (b Bytes) NewReadCloser() (io.ReadCloser, error) {
	return ReadCloser{bytes.NewReader(([]byte)(b))}, nil
}
