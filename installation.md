# Ansipath installation

`ansipath` is distributed as one Go binary. It supports Bash and Zsh shell
initialization files on Unix-like systems.

## Build from source

Install Go, then run:

```sh
go build -trimpath -o ansipath .
./ansipath install
```

The default user installation:

- copies the binary to `$HOME/.local/bin/ansipath`;
- copies shell helpers to `$HOME/.local/share/ansipath/shell_helpers`;
- updates existing `.bashrc`, `.bash_profile`, `.profile`, `.zprofile`, and
  `.zshrc` files selected for the active shell;
- leaves missing shell files untouched.

To install the runtime system-wide while updating the invoking user’s shell
files, run:

```sh
sudo ./ansipath install
```

This uses `SUDO_USER` to find the user’s home directory and shell files, and
installs the binary under `/usr/bin`. A custom prefix is also supported:

```sh
./ansipath install --prefix "$HOME/.local"
```

Source the changed shell file or start a new shell when installation finishes.

## Palettes

The default palette is `catppuccin_mocha`. Palettes are embedded in the binary
and selected at runtime:

```sh
ansipath palette list
ansipath palette dracula
```

The selection is stored in `$XDG_STATE_HOME/ansipath/palette`, or
`$HOME/.local/state/ansipath/palette` when `XDG_STATE_HOME` is unset.

## Cleanup

Remove duplicate or dead literal path entries from existing shell-init files:

```sh
ansipath remove-dupes
```

The command prints an audit trail, asks `Undo? N/y`, and restores the original
files if you answer `y`. For an environment value, use:

```sh
ansipath remove-dupes --env
```

This prints the cleaned export for review; it is output-only because a child
process cannot mutate its parent shell environment.
