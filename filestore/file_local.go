package filestore

import (
	"encoding/base32"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sync"
)

type fileLocalStore struct {
	dir  string            // directory to store file
	name map[string]string // id -> original filename
	mu   sync.RWMutex
}

func NewFileLocalStore(dir string) FileStore {
	return &fileLocalStore{
		dir:  filepath.Clean(dir),
		name: make(map[string]string),
	}
}

func (s *fileLocalStore) Add(name string, r io.Reader) (string, error) {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir store dir %q: %w", s.dir, err)
	}

	ext := filepath.Ext(name)

	id, err := generateID()
	if err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}

	dst := filepath.Join(s.dir, id+ext)
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return "", fmt.Errorf("open dst file %q: %w", dst, err)
	}

	if f == nil {
		return "", fmt.Errorf("cannot allocate unique id after retries")
	}

	// 如果写入失败，删除半截文件
	ok := false
	defer func() {
		_ = f.Close()
		if !ok {
			_ = os.Remove(dst)
		}
	}()

	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("write dst file: %w", err)
	}

	ok = true

	s.mu.Lock()
	s.name[id] = name
	s.mu.Unlock()

	return id, nil
}

func (s *fileLocalStore) Get(id string) (string, *os.File) {
	s.mu.RLock()
	name := s.name[id]
	s.mu.RUnlock()

	// 1) try by ext from name
	if name != "" {
		ext := filepath.Ext(name)
		p := filepath.Join(s.dir, id+ext)
		if f, err := os.Open(p); err == nil {
			return name, f
		}
	}

	// 2) fallback: try without ext
	p0 := filepath.Join(s.dir, id)
	if f, err := os.Open(p0); err == nil {
		if name == "" {
			name = filepath.Base(p0)
		}
		return name, f
	}

	// 3) fallback: glob search (useful after restart when map is empty)
	matches, _ := filepath.Glob(filepath.Join(s.dir, id+".*"))
	if len(matches) > 0 {
		if f, err := os.Open(matches[0]); err == nil {
			if name == "" {
				name = filepath.Base(matches[0])
			}
			return name, f
		}
	}

	// not found
	return name, nil
}

func generateID() (string, error) {
	const randIDLength = 5
	b := make([]byte, randIDLength)
	r := rand.Int64N(1 << 40)
	b[0] = byte(r)
	b[1] = byte(r >> 8)
	b[2] = byte(r >> 16)
	b[3] = byte(r >> 24)
	b[4] = byte(r >> 32)
	return base32.StdEncoding.EncodeToString(b), nil
}
