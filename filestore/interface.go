package filestore

import (
	"io"
	"os"
)

type FileStore interface {
	Add(name string, r io.Reader) (string, error)
	Get(string) (string, *os.File)
}
