==================================================================
         🚀 Ansipath Manual Installation Instructions        
==================================================================

1. Check if running inside Windows CMD or PowerShell (unsupported)

```sh
if [ -n "$COMSPEC" ] || [ -n "$PSModulePath" ]; then
    echo "Error: This script cannot be run from Windows CMD or PowerShell." >&2
    echo "Maybe you meant to install inside WSL?"
    exit 1
fi
```

2. Check if `uv` is installed

```sh
if ! command -v uv >/dev/null 2>&1; then
    echo " ❌ Error: 'uv' package manager was not found." >&2
    echo "Please install uv first: " >&2
    echo "curl -LsSf https://astral.sh/uv/install.sh | sh" >&2
    exit 1
fi
```

3. Build and Install ansipath to run Globally

```sh
echo " 📦 Building Python application binary via uv..."
uv tool install . --force --no-managed-python
```

4. Install the helper scripts:

```sh
echo " 📦 Installing and configuring helper scripts..."
chmod +x && ./install_helpers.sh

```

🎉 And you're done!
