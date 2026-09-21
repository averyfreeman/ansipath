# Ansipath installation

`ansipath` is a single Go binary for Bash and Zsh on Unix-like systems.

## Build and install

```sh
go build -trimpath -o ansipath .
./ansipath install
```

The default installation writes the binary to `$HOME/.local/bin/ansipath`,
helpers to `$HOME/.local/share/ansipath/shell_helpers`, and updates only shell
init files that already exist. It does not create missing profiles.

For a system prefix while targeting the invoking user’s shell files:

```sh
sudo ./ansipath install
```

When run through `sudo`, `SUDO_USER` determines the home directory and shell;
the executable is installed under `/usr/bin` by default. Use
`--prefix "$HOME/.local"` to select another prefix.

## Palettes

```sh
ansipath palette list
ansipath palette dracula
```

Palette selection is retained in `$XDG_STATE_HOME/ansipath/palette`, falling
back to `$HOME/.local/state/ansipath/palette`.

## Cleanup

```sh
ansipath remove-dupes
```

This removes duplicate or dead literal path entries from existing shell-init
assignments, prints an audit trail, and prompts `Undo? N/y`. Answer `y` to
restore the original files. `ansipath remove-dupes --env` instead prints a
cleaned environment export because child processes cannot mutate their parent.
