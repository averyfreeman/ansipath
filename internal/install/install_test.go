package install_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/averyfreeman/ansipath/internal/install"
)

func TestResolveTargetUsesSudoUserHomeAndSystemPrefix(t *testing.T) {
	home := t.TempDir()
	for _, name := range []string{".bashrc", ".bash_profile"} {
		if err := os.WriteFile(filepath.Join(home, name), []byte(""), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	target, err := install.ResolveTarget(install.ResolveOptions{
		Environment: map[string]string{
			"SUDO_USER":      "avery",
			"SUDO_USER_HOME": home,
			"SHELL":          "/bin/bash",
		},
		EffectiveUID: 0,
		LookupUser: func(string) (install.User, error) {
			t.Fatal("LookupUser should not be called when SUDO_USER_HOME is set")
			return install.User{}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if target.BinaryPath != "/usr/bin/ansipath" {
		t.Fatalf("BinaryPath = %q", target.BinaryPath)
	}
	if target.DataDir != filepath.Join(home, ".local/share/ansipath/shell_helpers") {
		t.Fatalf("DataDir = %q", target.DataDir)
	}
	if len(target.Profiles) != 2 {
		t.Fatalf("Profiles = %v", target.Profiles)
	}
}

func TestInstallWritesExecutableAssetsAndProfile(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.WriteFile(source, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	profile := filepath.Join(root, ".bashrc")
	if err := os.WriteFile(profile, []byte("export PATH=\"/one:$PATH\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	target := install.Target{
		BinaryPath: filepath.Join(root, "bin", "ansipath"),
		DataDir:    filepath.Join(root, "share", "ansipath"),
		Profiles:   []string{profile},
	}
	assets := fstest.MapFS{
		"shell_helpers/pathmunge.sh": {Data: []byte("pathmunge() { :; }\n")},
		"shell_helpers/whichfunc.sh": {Data: []byte("which() { :; }\n")},
	}
	result, err := install.Install(source, target, assets, map[string]string{"PATH": "/usr/bin"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changed) != 4 {
		t.Fatalf("Changed = %v", result.Changed)
	}
	if data, err := os.ReadFile(target.BinaryPath); err != nil || string(data) != "binary" {
		t.Fatalf("installed binary = %q, err = %v", data, err)
	}
	profileData, err := os.ReadFile(profile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(profileData), "export PATH=") || !strings.Contains(string(profileData), "pathmunge prepend") {
		t.Fatalf("profile was not rewritten:\n%s", profileData)
	}
}
