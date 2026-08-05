# Ansipath manual installation

The helper installer supports Bash and Zsh on Unix-like systems. Run these
commands from a shell, not Windows CMD or PowerShell.

1. Confirm that the active shell is supported.

```sh
case "${SHELL##*/}" in
    bash|zsh) ;;
    *)
        echo "Error: install_helpers.sh supports Bash and Zsh only." >&2
        exit 1
        ;;
esac
```

2. Check that `uv` is installed.

```sh
if ! command -v uv >/dev/null 2>&1; then
    echo "Error: 'uv' package manager was not found." >&2
    echo "Install uv first: https://docs.astral.sh/uv/getting-started/installation/" >&2
    exit 1
fi
```

3. Select the palette and install ansipath as a global `uv` tool from the
repository root. The build script copies the selected palette into the package
configuration before invoking `uv tool install`.

```sh
cp -v palettes/catppuccin_mocha.json palettes/palette.json
./scripts/build.sh
```

Restart the terminal, or source the profile file reported by the installer.
