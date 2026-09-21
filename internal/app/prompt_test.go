package app

import (
	"strings"
	"testing"
)

func TestReadUndoKeysIgnoresEditingCharacters(t *testing.T) {
	got, err := readUndoKeys(strings.NewReader("\b\x7fY"))
	if err != nil {
		t.Fatalf("readUndoKeys returned error: %v", err)
	}
	if !got {
		t.Fatal("readUndoKeys returned N after ignored editing characters and Y")
	}
}

func TestReadUndoKeysTreatsEnterAsNo(t *testing.T) {
	got, err := readUndoKeys(strings.NewReader("\n"))
	if err != nil {
		t.Fatalf("readUndoKeys returned error: %v", err)
	}
	if got {
		t.Fatal("readUndoKeys returned Y for Enter")
	}
}

func TestReadUndoResponseUsesLineFallbackForNonTTY(t *testing.T) {
	got, err := readUndoResponse(strings.NewReader("y\n"))
	if err != nil {
		t.Fatalf("readUndoResponse returned error: %v", err)
	}
	if !got {
		t.Fatal("readUndoResponse returned N for piped Y input")
	}
}
