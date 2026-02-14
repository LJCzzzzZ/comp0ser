package filestore

import (
	"fmt"
	"os"
	"testing"
)

func TestFileLocalStore_New(t *testing.T) {
	fs := NewFileLocalStore("store")
	filename, err := fs.New("test")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(filename)
}

func TestFileLocalStore_Add(t *testing.T) {
	fs := NewFileLocalStore("/mnt/media/data")
	if err := fs.Add("jupiter", "0000", map[string]any{"audio_id": "ttt"}, nil); err != nil {
		t.Fatal(err)
	}
}

func TestFileLocalStore_Append(t *testing.T) {
	fs := NewFileLocalStore("store")
	id, err := fs.Append("test", "", "hello wolrd", map[string]any{
		"field_1": 10,
		"field_2": "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(id)
}

func TestFileLocalStore_Save(t *testing.T) {
	fs := NewFileLocalStore("store")
	f, err := os.Open("/mnt/media/data/0001.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	a, b, err := fs.Save("test", "test1", ".wav", f)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s %s\n", a, b)
}

func TestFileLocalStore_List(t *testing.T) {
	fs := NewFileLocalStore("/mnt/media/data")
	nars, err := fs.List("neutron_star")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(nars)
}
