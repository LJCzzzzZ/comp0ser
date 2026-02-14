package filestore

import "io"

type FileStore interface {
	New(name string) (string, error)
	Add(name, id string, set map[string]any, unset []string) error
	Append(name, id, text string, extra map[string]any) (string, error)
	Save(name, filename, ext string, r io.Reader) (string, string, error)
	List(name string) ([]map[string]any, error)

	Dir() string
}
