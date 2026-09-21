package install

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/averyfreeman/ansipath/internal/shellinit"
)

type User struct {
	Home  string
	Shell string
	UID   int
	GID   int
}

type ResolveOptions struct {
	Environment  map[string]string
	EffectiveUID int
	LookupUser   func(string) (User, error)
	Prefix       string
}

type Owner struct {
	UID int
	GID int
}

type Target struct {
	Home       string
	Shell      string
	Prefix     string
	BinaryPath string
	DataDir    string
	Profiles   []string
	Owner      *Owner
}

type Result struct {
	Changed []string
}

func ResolveTarget(options ResolveOptions) (Target, error) {
	environment := options.Environment
	if environment == nil {
		environment = environmentMap(os.Environ())
	}
	if options.LookupUser == nil {
		options.LookupUser = func(name string) (User, error) {
			return User{}, fmt.Errorf("lookup for %q is unavailable", name)
		}
	}

	target := Target{}
	if options.EffectiveUID == 0 && environment["SUDO_USER"] != "" {
		target.Owner = nil
		if home := environment["SUDO_USER_HOME"]; home != "" {
			target.Home = filepath.Clean(home)
		} else {
			user, err := options.LookupUser(environment["SUDO_USER"])
			if err != nil {
				return Target{}, fmt.Errorf("resolve SUDO_USER home: %w", err)
			}
			target.Home = filepath.Clean(user.Home)
			target.Owner = &Owner{UID: user.UID, GID: user.GID}
			if environment["SUDO_USER_SHELL"] == "" {
				environment["SUDO_USER_SHELL"] = user.Shell
			}
		}
		target.Shell = firstNonEmpty(environment["SUDO_USER_SHELL"], environment["SHELL"])
		if target.Owner == nil {
			if uid, uidErr := strconv.Atoi(environment["SUDO_UID"]); uidErr == nil {
				if gid, gidErr := strconv.Atoi(environment["SUDO_GID"]); gidErr == nil {
					target.Owner = &Owner{UID: uid, GID: gid}
				}
			}
		}
	} else {
		target.Home = filepath.Clean(firstNonEmpty(environment["HOME"], os.Getenv("HOME")))
		target.Shell = environment["SHELL"]
	}
	if target.Home == "." || target.Home == "" {
		return Target{}, fmt.Errorf("HOME is not set")
	}
	if target.Shell == "" {
		return Target{}, fmt.Errorf("SHELL is not set")
	}

	target.Prefix = options.Prefix
	if target.Prefix == "" {
		if options.EffectiveUID == 0 {
			target.Prefix = "/usr"
		} else {
			target.Prefix = filepath.Join(target.Home, ".local")
		}
	}
	target.Prefix = filepath.Clean(target.Prefix)
	target.BinaryPath = filepath.Join(target.Prefix, "bin", "ansipath")
	dataHome := ""
	if options.EffectiveUID == 0 && environment["SUDO_USER"] != "" {
		dataHome = environment["SUDO_USER_XDG_DATA_HOME"]
	} else {
		dataHome = environment["XDG_DATA_HOME"]
	}
	if dataHome == "" {
		dataHome = filepath.Join(target.Home, ".local", "share")
	}
	target.DataDir = filepath.Join(dataHome, "ansipath", "shell_helpers")
	target.Profiles = existingProfiles(target.Home, target.Shell)
	return target, nil
}

func Install(source string, target Target, assets fs.FS, environment map[string]string) (Result, error) {
	if _, err := os.Stat(source); err != nil {
		return Result{}, fmt.Errorf("read executable: %w", err)
	}
	result := Result{Changed: make([]string, 0)}
	if !samePath(source, target.BinaryPath) {
		data, err := os.ReadFile(source)
		if err != nil {
			return Result{}, fmt.Errorf("read executable: %w", err)
		}
		changed, err := writeAtomic(target.BinaryPath, data, 0o755, nil)
		if err != nil {
			return Result{}, fmt.Errorf("install executable: %w", err)
		}
		if changed {
			result.Changed = append(result.Changed, target.BinaryPath)
		}
	}

	for _, name := range []string{"pathmunge.sh", "whichfunc.sh"} {
		data, err := fs.ReadFile(assets, "shell_helpers/"+name)
		if err != nil {
			return Result{}, fmt.Errorf("read helper %q: %w", name, err)
		}
		filename := filepath.Join(target.DataDir, name)
		changed, err := writeAtomic(filename, data, 0o644, target.Owner)
		if err != nil {
			return Result{}, fmt.Errorf("install helper %q: %w", name, err)
		}
		if changed {
			result.Changed = append(result.Changed, filename)
		}
	}

	for _, profile := range target.Profiles {
		original, err := os.ReadFile(profile)
		if err != nil {
			return Result{}, fmt.Errorf("read profile %q: %w", profile, err)
		}
		info, err := os.Stat(profile)
		if err != nil {
			return Result{}, fmt.Errorf("stat profile %q: %w", profile, err)
		}
		rewritten, _, err := shellinit.Rewrite(string(original), shellinit.Options{
			DataDir:     target.DataDir,
			BinaryPath:  target.BinaryPath,
			Environment: environment,
		})
		if err != nil {
			return Result{}, fmt.Errorf("rewrite profile %q: %w", profile, err)
		}
		if rewritten == string(original) {
			continue
		}
		changed, err := writeAtomic(profile, []byte(rewritten), info.Mode().Perm(), target.Owner)
		if err != nil {
			return Result{}, fmt.Errorf("write profile %q: %w", profile, err)
		}
		if changed {
			result.Changed = append(result.Changed, profile)
		}
	}
	return result, nil
}

// WriteAtomic writes a file without exposing a partially written version.
func WriteAtomic(filename string, data []byte, mode os.FileMode, owner *Owner) error {
	if _, err := writeAtomic(filename, data, mode, owner); err != nil {
		return err
	}
	return nil
}

func existingProfiles(home, shell string) []string {
	base := filepath.Base(shell)
	paths := make([]string, 0, 2)
	interactive := ""
	login := []string{}
	switch base {
	case "bash":
		interactive = ".bashrc"
		login = []string{".bash_profile", ".profile"}
	case "zsh":
		interactive = ".zshrc"
		login = []string{".zprofile", ".profile"}
	default:
		return nil
	}
	if exists(filepath.Join(home, interactive)) {
		paths = append(paths, filepath.Join(home, interactive))
	}
	for _, name := range login {
		filename := filepath.Join(home, name)
		if exists(filename) {
			paths = append(paths, filename)
			break
		}
	}
	return paths
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func samePath(left, right string) bool {
	leftAbs, leftErr := filepath.Abs(left)
	rightAbs, rightErr := filepath.Abs(right)
	return leftErr == nil && rightErr == nil && filepath.Clean(leftAbs) == filepath.Clean(rightAbs)
}

func writeAtomic(filename string, data []byte, mode os.FileMode, owner *Owner) (bool, error) {
	if existing, err := os.ReadFile(filename); err == nil && string(existing) == string(data) {
		return false, nil
	} else if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return false, err
	}
	temporary, err := os.CreateTemp(filepath.Dir(filename), ".ansipath-*")
	if err != nil {
		return false, err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(mode); err != nil {
		_ = temporary.Close()
		return false, err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return false, err
	}
	if err := temporary.Close(); err != nil {
		return false, err
	}
	if err := os.Rename(temporaryName, filename); err != nil {
		return false, err
	}
	if owner != nil && os.Geteuid() == 0 {
		if err := os.Chown(filename, owner.UID, owner.GID); err != nil {
			return false, err
		}
	}
	return true, nil
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
