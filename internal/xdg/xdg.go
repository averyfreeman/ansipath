package xdg

import (
	"fmt"
	"os"
	"path/filepath"
)

func HomeDir() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Clean(home), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Clean(home), nil
}

func DataDir() (string, error) {
	home, err := HomeDir()
	if err != nil {
		return "", err
	}
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, "ansipath"), nil
}

func StateDir() (string, error) {
	home, err := HomeDir()
	if err != nil {
		return "", err
	}
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		base = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(base, "ansipath"), nil
}
