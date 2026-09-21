package app_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/averyfreeman/ansipath/internal/app"
)

const testPalette = `{
  "palette": ["R", "G"],
  "reset": "RESET",
  "yellow_bold": "YELLOW",
  "dim_red": "DIMRED"
}`

func testAssets() fstest.MapFS {
	return fstest.MapFS{
		"palettes/catppuccin_mocha.json": {Data: []byte(testPalette)},
		"palettes/dracula.json":          {Data: []byte(testPalette)},
		"shell_helpers/pathmunge.sh":     {Data: []byte("pathmunge() { :; }\n")},
		"shell_helpers/whichfunc.sh":     {Data: []byte("which() { :; }\n")},
	}
}

func TestRunInstallCopiesBinaryAssetsAndUpdatesExistingProfile(t *testing.T) {
	home := t.TempDir()
	prefix := filepath.Join(t.TempDir(), "prefix")
	profile := filepath.Join(home, ".bashrc")
	if err := os.WriteFile(profile, []byte("export PATH=\"/one:$PATH\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("SHELL", "/bin/bash")
	t.Setenv("PATH", "/usr/bin")
	var stdout, stderr bytes.Buffer

	code := app.Run([]string{"install", "--prefix", prefix}, strings.NewReader(""), &stdout, &stderr, testAssets())
	if code != 0 {
		t.Fatalf("Run exit code = %d, stderr = %q", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(prefix, "bin", "ansipath")); err != nil {
		t.Fatalf("installed binary missing: %v", err)
	}
	profileData, err := os.ReadFile(profile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(profileData), "ansipath shell helpers") {
		t.Fatalf("profile was not updated:\n%s", profileData)
	}
}

func TestRunRemoveDupesAppliesAndKeepsChangesOnN(t *testing.T) {
	home := t.TempDir()
	profile := filepath.Join(home, ".bashrc")
	live := filepath.Join(home, "live")
	if err := os.Mkdir(live, 0o755); err != nil {
		t.Fatal(err)
	}
	original := "export PATH=\"" + live + ":" + live + ":" + filepath.Join(home, "missing") + "\"\n"
	if err := os.WriteFile(profile, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("SHELL", "/bin/bash")
	t.Setenv("PATH", live)
	var stdout, stderr bytes.Buffer

	code := app.Run([]string{"remove-dupes"}, strings.NewReader("N\n"), &stdout, &stderr, testAssets())
	if code != 0 {
		t.Fatalf("Run exit code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "removed duplicate") || !strings.Contains(stdout.String(), "removed dead") || !strings.Contains(stdout.String(), "Undo? N/y") {
		t.Fatalf("missing audit or prompt: %q", stdout.String())
	}
	updated, err := os.ReadFile(profile)
	if err != nil {
		t.Fatal(err)
	}
	if string(updated) == original {
		t.Fatal("profile was not changed")
	}
}

func TestRunRemoveDupesRestoresChangesOnY(t *testing.T) {
	home := t.TempDir()
	profile := filepath.Join(home, ".zshrc")
	live := filepath.Join(home, "live")
	if err := os.Mkdir(live, 0o755); err != nil {
		t.Fatal(err)
	}
	original := "PATH=\"" + live + ":" + live + "\"\n"
	if err := os.WriteFile(profile, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("PATH", live)
	var stdout, stderr bytes.Buffer

	code := app.Run([]string{"remove-dupes"}, strings.NewReader("y\n"), &stdout, &stderr, testAssets())
	if code != 0 {
		t.Fatalf("Run exit code = %d, stderr = %q", code, stderr.String())
	}
	restored, err := os.ReadFile(profile)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != original {
		t.Fatalf("profile was not restored: %q", restored)
	}
}

func TestRunRemoveDupesEnvEmitsSourceableCleanupAndCopiesExport(t *testing.T) {
	home := t.TempDir()
	live := filepath.Join(home, "live")
	clipboardDir := filepath.Join(home, "clipboard")
	capture := filepath.Join(home, "captured")
	dead := filepath.Join(home, "missing")
	for _, directory := range []string{live, clipboardDir} {
		if err := os.Mkdir(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	pbcopy := filepath.Join(clipboardDir, "pbcopy")
	if err := os.WriteFile(pbcopy, []byte("#!/bin/sh\n/bin/cat > \"$ANSIPATH_CLIPBOARD_CAPTURE\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", strings.Join([]string{live, live, dead, dead, clipboardDir}, ":"))
	t.Setenv("ANSIPATH_CLIPBOARD_CAPTURE", capture)
	var stdout, stderr bytes.Buffer

	code := app.Run([]string{"remove-dupes", "--env"}, strings.NewReader(""), &stdout, &stderr, testAssets())
	if code != 0 {
		t.Fatalf("Run exit code = %d, stderr = %q", code, stderr.String())
	}
	output := stdout.String()
	if strings.Count(output, "pathmunge remove '"+live+"'") != 1 {
		t.Fatalf("duplicate removal count is wrong: %q", output)
	}
	if strings.Count(output, "pathmunge remove '"+dead+"'") != 2 {
		t.Fatalf("dead removal count is wrong: %q", output)
	}
	if !strings.Contains(output, "# changed: env PATH removed duplicate") ||
		!strings.Contains(output, "# changed: env PATH removed dead") {
		t.Fatalf("sourceable audit comments missing: %q", output)
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	wantExport := "export PATH='" + live + ":" + clipboardDir + "'"
	if lines[len(lines)-1] != wantExport {
		t.Fatalf("final export = %q, want %q", lines[len(lines)-1], wantExport)
	}
	copied, err := os.ReadFile(capture)
	if err != nil {
		t.Fatalf("read clipboard capture: %v", err)
	}
	if string(copied) != wantExport+"\n" {
		t.Fatalf("clipboard content = %q, want %q", copied, wantExport+"\n")
	}
}

func TestRunRendersPathFromEnvironment(t *testing.T) {
	t.Setenv("PATH", "/usr/bin:/bin")
	t.Setenv("HOME", t.TempDir())
	var stdout, stderr bytes.Buffer

	code := app.Run(nil, strings.NewReader(""), &stdout, &stderr, testAssets())
	if code != 0 {
		t.Fatalf("Run exit code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "usr") || !strings.Contains(stdout.String(), "bin") {
		t.Fatalf("unexpected output: %q", stdout.String())
	}
}

func TestRunHelpDocumentsCommands(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := app.Run([]string{"--help"}, strings.NewReader(""), &stdout, &stderr, testAssets())
	if code != 0 {
		t.Fatalf("Run exit code = %d, stderr = %q", code, stderr.String())
	}
	for _, want := range []string{"ansipath install", "ansipath palette list", "ansipath remove-dupes"} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("help missing %q: %s", want, stdout.String())
		}
	}
}

func TestPaletteSelectionPersists(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", "/usr/bin")
	var stdout, stderr bytes.Buffer
	if code := app.Run([]string{"palette", "dracula"}, strings.NewReader(""), &stdout, &stderr, testAssets()); code != 0 {
		t.Fatalf("palette set exit code = %d, stderr = %q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := app.Run([]string{"palette", "list"}, strings.NewReader(""), &stdout, &stderr, testAssets()); code != 0 {
		t.Fatalf("palette list exit code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "* dracula") {
		t.Fatalf("active palette missing: %q", stdout.String())
	}
}
