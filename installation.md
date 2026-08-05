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

3. Install ansipath as a global `uv` tool from the repository root.

```sh
uv tool install . --force
```

4. Install the helper scripts.

```sh
chmod +x ./install_helpers.sh
./install_helpers.sh
```

Restart the terminal, or source the profile file reported by the installer.
