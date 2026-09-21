package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/averyfreeman/ansipath/internal/install"
	"github.com/averyfreeman/ansipath/internal/palette"
	"github.com/averyfreeman/ansipath/internal/pathview"
	"github.com/averyfreeman/ansipath/internal/shellinit"
	"github.com/averyfreeman/ansipath/internal/xdg"
)

const Version = "0.2.0"

type originalFile struct {
	name string
	data []byte
	mode os.FileMode
}

// Run dispatches one ansipath command and returns its process exit code.
func Run(args []string, in io.Reader, out, errOut io.Writer, assets fs.FS) int {
	if len(args) == 0 {
		return render(args, out, errOut, assets)
	}

	switch args[0] {
	case "--help", "-h":
		_, _ = io.WriteString(out, helpText)
		return 0
	case "--version":
		_, _ = fmt.Fprintln(out, Version)
		return 0
	case "install":
		return runInstall(args[1:], out, errOut, assets)
	case "palette":
		return runPalette(args[1:], out, errOut, assets)
	case "remove-dupes":
		return runRemoveDupes(args[1:], in, out, errOut)
	default:
		if strings.HasPrefix(args[0], "-") {
			return fail(errOut, "unknown option %q", args[0])
		}
		return render(args, out, errOut, assets)
	}
}

func runRemoveDupes(args []string, in io.Reader, out, errOut io.Writer) int {
	environmentMode := false
	for _, arg := range args {
		switch arg {
		case "--env":
			environmentMode = true
		case "--help", "-h":
			_, _ = io.WriteString(out, "Usage: ansipath remove-dupes [--env]\n")
			return 0
		default:
			return fail(errOut, "unknown remove-dupes option %q", arg)
		}
	}

	if environmentMode {
		cleaned, changes := shellinit.CleanPathValue(os.Getenv("PATH"), nil)
		for _, change := range changes {
			_, _ = fmt.Fprintf(out, "changed: env PATH removed %s %s\n", change.Reason, change.Path)
		}
		if len(changes) == 0 {
			_, _ = io.WriteString(out, "No duplicate or dead PATH entries found.\n")
			return 0
		}
		_, _ = fmt.Fprintf(out, "export PATH=%s\n", shellQuote(cleaned))
		return 0
	}

	environment := environmentMap(os.Environ())
	target, err := install.ResolveTarget(install.ResolveOptions{
		Environment:  environment,
		EffectiveUID: os.Geteuid(),
		LookupUser:   lookupUser,
	})
	if err != nil {
		return fail(errOut, "%v", err)
	}
	if len(target.Profiles) == 0 {
		return fail(errOut, "no existing supported shell init file found")
	}

	originals := make([]originalFile, 0)
	for _, profile := range target.Profiles {
		original, err := os.ReadFile(profile)
		if err != nil {
			_ = restoreOriginals(originals, target.Owner)
			return fail(errOut, "read profile %q: %v", profile, err)
		}
		info, err := os.Stat(profile)
		if err != nil {
			_ = restoreOriginals(originals, target.Owner)
			return fail(errOut, "stat profile %q: %v", profile, err)
		}
		cleaned, changes, err := shellinit.Clean(string(original), shellinit.CleanOptions{
			Options: shellinit.Options{
				DataDir:     target.DataDir,
				BinaryPath:  target.BinaryPath,
				Environment: environment,
			},
		})
		if err != nil {
			_ = restoreOriginals(originals, target.Owner)
			return fail(errOut, "clean profile %q: %v", profile, err)
		}
		if len(changes) == 0 || cleaned == string(original) {
			continue
		}
		if err := install.WriteAtomic(profile, []byte(cleaned), info.Mode().Perm(), target.Owner); err != nil {
			_ = restoreOriginals(originals, target.Owner)
			return fail(errOut, "write profile %q: %v", profile, err)
		}
		originals = append(originals, originalFile{name: profile, data: original, mode: info.Mode().Perm()})
		for _, change := range changes {
			_, _ = fmt.Fprintf(out, "changed: %s:%d: removed %s %s\n", profile, change.Line, change.Reason, change.Path)
		}
	}

	if len(originals) == 0 {
		_, _ = io.WriteString(out, "No duplicate or dead PATH entries found.\n")
		return 0
	}
	_, _ = io.WriteString(out, "Undo? N/y\n")
	answer, readErr := bufio.NewReader(in).ReadString('\n')
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return fail(errOut, "read undo response: %v", readErr)
	}
	if strings.EqualFold(strings.TrimSpace(answer), "y") {
		if err := restoreOriginals(originals, target.Owner); err != nil {
			return fail(errOut, "undo changes: %v", err)
		}
		for _, original := range originals {
			_, _ = fmt.Fprintf(out, "undo: restored %s\n", original.name)
		}
	}
	return 0
}

func restoreOriginals(files []originalFile, owner *install.Owner) error {
	for _, file := range files {
		if err := install.WriteAtomic(file.name, file.data, file.mode, owner); err != nil {
			return err
		}
	}
	return nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func runInstall(args []string, out, errOut io.Writer, assets fs.FS) int {
	prefix := ""
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--help", "-h":
			_, _ = io.WriteString(out, "Usage: ansipath install [--prefix DIR]\n")
			return 0
		case "--prefix":
			if index+1 >= len(args) {
				return fail(errOut, "--prefix requires a directory")
			}
			prefix = args[index+1]
			index++
		default:
			return fail(errOut, "unknown install option %q", args[index])
		}
	}

	environment := environmentMap(os.Environ())
	target, err := install.ResolveTarget(install.ResolveOptions{
		Environment:  environment,
		EffectiveUID: os.Geteuid(),
		Prefix:       prefix,
		LookupUser:   lookupUser,
	})
	if err != nil {
		return fail(errOut, "%v", err)
	}
	source, err := os.Executable()
	if err != nil {
		return fail(errOut, "resolve executable: %v", err)
	}
	result, err := install.Install(source, target, assets, environment)
	if err != nil {
		return fail(errOut, "%v", err)
	}
	_, _ = fmt.Fprintf(out, "Installed ansipath at %s\n", target.BinaryPath)
	for _, filename := range result.Changed {
		_, _ = fmt.Fprintf(out, "changed: %s\n", filename)
	}
	if len(target.Profiles) > 0 {
		_, _ = fmt.Fprintf(out, "Reload your shell or source %s\n", target.Profiles[0])
	} else {
		_, _ = io.WriteString(out, "No existing supported shell init file was modified.\n")
	}
	return 0
}

func lookupUser(name string) (install.User, error) {
	account, err := user.Lookup(name)
	if err != nil {
		return install.User{}, err
	}
	uid, err := strconv.Atoi(account.Uid)
	if err != nil {
		return install.User{}, fmt.Errorf("parse UID for %q: %w", name, err)
	}
	gid, err := strconv.Atoi(account.Gid)
	if err != nil {
		return install.User{}, fmt.Errorf("parse GID for %q: %w", name, err)
	}
	return install.User{Home: account.HomeDir, UID: uid, GID: gid}, nil
}

func render(args []string, out, errOut io.Writer, assets fs.FS) int {
	if len(args) > 1 {
		return fail(errOut, "expected at most one PATH argument")
	}
	catalog, err := palette.Load(assets)
	if err != nil {
		return fail(errOut, "%v", err)
	}
	name, definition, err := selectedPalette(catalog, errOut)
	if err != nil {
		return fail(errOut, "%v", err)
	}
	_ = name

	pathSource := os.Getenv("PATH")
	if len(args) == 1 {
		pathSource = args[0]
	}
	output, err := pathview.Render(pathSource, pathview.Style{
		Palette:    definition.Colors,
		Reset:      definition.Reset,
		YellowBold: definition.YellowBold,
		DimRed:     definition.DimRed,
	}, nil)
	if err != nil {
		return fail(errOut, "Error: %v", err)
	}
	_, _ = io.WriteString(out, output)
	return 0
}

func runPalette(args []string, out, errOut io.Writer, assets fs.FS) int {
	catalog, err := palette.Load(assets)
	if err != nil {
		return fail(errOut, "%v", err)
	}
	if len(args) == 1 && args[0] == "list" {
		active, _, err := selectedPalette(catalog, errOut)
		if err != nil {
			return fail(errOut, "%v", err)
		}
		for _, name := range catalog.Names() {
			prefix := "  "
			if name == active {
				prefix = "* "
			}
			_, _ = fmt.Fprintf(out, "%s%s\n", prefix, name)
		}
		return 0
	}
	if len(args) != 1 {
		return fail(errOut, "usage: ansipath palette list | ansipath palette NAME")
	}
	if _, err := catalog.Get(args[0]); err != nil {
		return fail(errOut, "%v", err)
	}
	stateDir, err := xdg.StateDir()
	if err != nil {
		return fail(errOut, "%v", err)
	}
	if err := saveSelection(filepath.Join(stateDir, "palette"), args[0]); err != nil {
		return fail(errOut, "%v", err)
	}
	_, _ = fmt.Fprintf(out, "palette: %s\n", args[0])
	return 0
}

func selectedPalette(catalog palette.Catalog, errOut io.Writer) (string, palette.Definition, error) {
	name := palette.DefaultName
	stateDir, err := xdg.StateDir()
	if err != nil {
		return "", palette.Definition{}, err
	}
	data, err := os.ReadFile(filepath.Join(stateDir, "palette"))
	switch {
	case err == nil:
		candidate := strings.TrimSpace(string(data))
		if candidate != "" {
			if _, getErr := catalog.Get(candidate); getErr == nil {
				name = candidate
			} else {
				_, _ = fmt.Fprintf(errOut, "Warning: ignoring invalid palette state %q\n", candidate)
			}
		}
	case errors.Is(err, os.ErrNotExist):
		// The default is used until the user selects a palette.
	default:
		return "", palette.Definition{}, fmt.Errorf("read palette state: %w", err)
	}
	definition, err := catalog.Get(name)
	if err != nil {
		return "", palette.Definition{}, fmt.Errorf("resolve palette: %w", err)
	}
	return name, definition, nil
}

func saveSelection(filename, name string) error {
	directory := filepath.Dir(filename)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".palette-*")
	if err != nil {
		return fmt.Errorf("create palette state: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set palette state permissions: %w", err)
	}
	if _, err := temporary.WriteString(name + "\n"); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write palette state: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close palette state: %w", err)
	}
	if err := os.Rename(temporaryName, filename); err != nil {
		return fmt.Errorf("commit palette state: %w", err)
	}
	return nil
}

func fail(errOut io.Writer, format string, args ...any) int {
	_, _ = fmt.Fprintf(errOut, "Error: "+format+"\n", args...)
	return 1
}

func environmentMap(values []string) map[string]string {
	result := make(map[string]string, len(values))
	for _, value := range values {
		key, raw, ok := strings.Cut(value, "=")
		if ok {
			result[key] = raw
		}
	}
	return result
}

const helpText = `ansipath - colorized PATH diagnostics

Usage:
  ansipath [PATH]
  ansipath install [--prefix DIR]
  ansipath palette list
  ansipath palette NAME
  ansipath remove-dupes [--env]

With no command, ansipath renders the current PATH. A positional PATH
argument renders that value instead.

remove-dupes edits existing shell-init files, prints an audit trail, and asks
Undo? N/y. With --env it only prints a cleaned export because child processes
cannot change their parent shell environment.
`
