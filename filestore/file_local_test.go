package filestore

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestFileLocalStore_Add(t *testing.T) {
	dir := t.TempDir()
	fmt.Println(dir)
	store := NewFileLocalStore(dir)

	wantName := "test.json"
	wantContent := `{"segments":["hello","world"]}` + "\n"

	id, err := store.Add(wantName, strings.NewReader(wantContent))
	if err != nil {
		t.Fatalf("AddReader failed: %v", err)
	}
	if id == "" {
		t.Fatalf("AddReader returned empty id")
	}

	gotName, f := store.Get(id)
	if f == nil {
		t.Fatalf("Get returned nil file")
	}
	defer f.Close()

	if gotName != wantName {
		t.Fatalf("Get name mismatch, got=%q want=%q", gotName, wantName)
	}

	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("read stored file: %v", err)
	}

	if !bytes.Equal(b, []byte(wantContent)) {
		t.Fatalf("content mismatch, got=%q want=%q", string(b), wantContent)
	}
}
