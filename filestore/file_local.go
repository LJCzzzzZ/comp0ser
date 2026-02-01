package filestore

import (
	"bufio"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type fileLocalStore struct {
	dir string // directory to store file
	mu  sync.RWMutex
}

func NewFileLocalStore(dir string) FileStore {
	return &fileLocalStore{
		dir: filepath.Clean(dir),
	}
}

func (s *fileLocalStore) New(name string) (string, error) {
	tar := filepath.Join(s.dir, name)
	if err := os.MkdirAll(filepath.Join(tar, "audio"), 0o755); err != nil {
		return "", err
	}

	return tar, nil
}

func (s *fileLocalStore) Add(name, id string, set map[string]any, unset []string) error {
	tar := filepath.Join(s.dir, name)
	file := filepath.Join(tar, "narration.txt")
	fmt.Println(file)

	in, err := os.OpenFile(file, os.O_CREATE|os.O_RDONLY, 0o644)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp, err := os.CreateTemp(tar, ".narration.tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()

	sc := bufio.NewScanner(in)
	found := false

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		var o map[string]any
		if err := json.Unmarshal([]byte(line), &o); err != nil {
			return fmt.Errorf("bad jsonl line: %w", err)
		}

		if o["id"] == id {
			maps.Copy(o, set)
			for _, k := range unset {
				delete(o, k)
			}
			found = true
		}

		b, _ := json.Marshal(o)
		if _, err := tmp.Write(append(b, '\n')); err != nil {
			return err
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("not found")
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpName, file)
}

func (s *fileLocalStore) Append(name, id, text string, extra map[string]any) (string, error) {
	tar := filepath.Join(s.dir, name)
	o := map[string]any{
		"id":       id,
		"text":     text,
		"audio_id": "",
	}

	maps.Copy(o, extra)
	nar := filepath.Join(tar, "narration.txt")
	file, err := os.OpenFile(nar, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return "", err
	}

	b, _ := json.Marshal(o)
	_, err = file.Write(append(b, '\n'))
	return id, err
}

func (s *fileLocalStore) Save(name, filename, ext string, r io.Reader) (string, string, error) {
	tar := filepath.Join(s.dir, name)
	if ext == "" || !strings.HasPrefix(ext, ".") {
		return "", "", fmt.Errorf("bad ext: %q", ext)
	}
	fmt.Println(filename)

	dst := filepath.Join(tar, "audio", filename+ext)

	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return "", "", err
	}
	ok := false
	defer func() {
		_ = f.Close()
		if !ok {
			_ = os.Remove(dst)
		}
	}()

	if _, err := io.Copy(f, r); err != nil {
		return "", "", err
	}
	ok = true
	return filename, dst, nil
}

func (s *fileLocalStore) List(name string) ([]map[string]any, error) {
	nar := filepath.Join(s.dir, name, "narration.txt")
	fmt.Println(nar)
	f, err := os.OpenFile(nar, os.O_CREATE|os.O_RDONLY, 0o644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)

	var nars []map[string]any
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var o map[string]any
		if err := json.Unmarshal([]byte(line), &o); err != nil {
			return nil, fmt.Errorf("bad jsonl line: %w", err)
		}
		nars = append(nars, o)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return nars, nil
}

func (s *fileLocalStore) Dir() string {
	return s.dir
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
