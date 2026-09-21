# ansipath 🌈

`ansipath` makes `$PATH` legible: each entry is printed on its own line with
colors based on directory depth, duplicate entries are highlighted, and dead
directories are called out.

## Features

- A portable Go binary with no runtime dependencies.
- Colorized `$PATH` output and duplicate/dead-path diagnostics.
- Palette switching stored under `$XDG_STATE_HOME/ansipath` (or
  `$HOME/.local/state/ansipath`).
- Bash and Zsh `pathmunge` and `which` helpers.
- Shell-init rewriting for `.bashrc`, `.bash_profile`, `.profile`, `.zprofile`,
  and `.zshrc`.

## Installation

Build and install the binary from the repository root:

```sh
go build -trimpath -o ansipath .
./ansipath install
```

`ansipath install` installs the executable at `$HOME/.local/bin/ansipath`,
copies helper scripts to `$HOME/.local/share/ansipath/shell_helpers`, and
updates existing shell init files. It does not create missing init files.

For a system installation, use:

```sh
sudo ./ansipath install
```

With `sudo`, the executable is installed under `/usr/bin` and the shell files
belonging to `SUDO_USER` are updated. Source the changed profile or open a new
shell after installation.

## Usage

```sh
ansipath                 # render the current PATH
ansipath --help          # show all commands
ansipath palette list    # list palettes
ansipath palette dracula # select a palette for future runs
ansipath remove-dupes    # clean existing shell-init PATH assignments
```

`remove-dupes` prints every changed file, line, and path. It then asks
`Undo? N/y`; entering `y` restores the original files. To inspect an
environment value and apply cleanup to the current shell, use:

```sh
eval "$(ansipath remove-dupes --env)"
```

The environment mode emits shell-safe audit comments, one `pathmunge remove`
call for each entry being removed, and a final cleaned `export PATH=...` line.
Duplicate entries are removed `N-1` times so one copy remains. The final export
is copied to the clipboard automatically when a supported clipboard utility is
available; otherwise, copy that last line manually. It does not prompt because
a child process cannot directly change its parent shell.

The shell rewriter recognizes `export PATH="..."`, `PATH="..."`, and bare
`PATH` anchors. It loads `pathmunge` before generated calls, preserves the
execution context of functions and conditionals, and keeps discovered path
ordering deterministic.

## Development

```sh
go test ./...
go vet ./...
go build -trimpath -o ansipath .
```

Palettes are JSON files in `palettes/` and are embedded into the binary at
build time.
