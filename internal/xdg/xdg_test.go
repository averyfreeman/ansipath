package xdg_test

import (
	"path/filepath"
	"testing"

	"github.com/averyfreeman/ansipath/internal/xdg"
)

func TestPathsUseXDGOverrides(t *testing.T) {
	t.Setenv("HOME", "/tmp/ansipath-home")
	t.Setenv("XDG_DATA_HOME", "/tmp/ansipath-data")
	t.Setenv("XDG_STATE_HOME", "/tmp/ansipath-state")

	dataDir, err := xdg.DataDir()
	if err != nil {
		t.Fatal(err)
	}
	stateDir, err := xdg.StateDir()
	if err != nil {
		t.Fatal(err)
	}
	if dataDir != filepath.Join("/tmp/ansipath-data", "ansipath") {
		t.Fatalf("DataDir = %q", dataDir)
	}
	if stateDir != filepath.Join("/tmp/ansipath-state", "ansipath") {
		t.Fatalf("StateDir = %q", stateDir)
	}
}

func TestPathsUseLocalDefaults(t *testing.T) {
	t.Setenv("HOME", "/tmp/ansipath-home")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_STATE_HOME", "")

	dataDir, err := xdg.DataDir()
	if err != nil {
		t.Fatal(err)
	}
	stateDir, err := xdg.StateDir()
	if err != nil {
		t.Fatal(err)
	}
	if dataDir != "/tmp/ansipath-home/.local/share/ansipath" {
		t.Fatalf("DataDir = %q", dataDir)
	}
	if stateDir != "/tmp/ansipath-home/.local/state/ansipath" {
		t.Fatalf("StateDir = %q", stateDir)
	}
}
