#!/usr/bin/env bash

# Install the shell helpers next to the user's data and load them from the
# interactive configuration file for the active supported shell.
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
source_helpers_dir="$script_dir/shell_helpers"

if [[ ! -d "$source_helpers_dir" ]]; then
    echo "Error: could not find shell_helpers beside this installer." >&2
    exit 1
fi

case "${SHELL##*/}" in
    bash)
        target_profile="$HOME/.bashrc"
        ;;
    zsh)
        target_profile="$HOME/.zshrc"
        ;;
    *)
        echo "Error: shell helpers support Bash and Zsh only (SHELL=${SHELL:-unset})." >&2
        exit 1
        ;;
esac

data_home="${XDG_DATA_HOME:-"$HOME/.local/share"}"
helpers_dir="$data_home/ansipath/shell_helpers"
begin_marker="# >>> ansipath shell helpers >>>"

echo " • Installing helper scripts in $helpers_dir"
mkdir -p "$helpers_dir"
cp "$source_helpers_dir/pathmunge.sh" "$source_helpers_dir/whichfunc.sh" "$helpers_dir/"

touch "$target_profile"

if ! grep -Fqx "$begin_marker" "$target_profile"; then
    echo " • Adding helper load statements to $target_profile"
    cat >> "$target_profile" <<'EOF'

# >>> ansipath shell helpers >>>
ansipath_helpers_dir="${XDG_DATA_HOME:-"$HOME/.local/share"}/ansipath/shell_helpers"
if [ -r "$ansipath_helpers_dir/whichfunc.sh" ]; then
    . "$ansipath_helpers_dir/whichfunc.sh"
fi
if [ -r "$ansipath_helpers_dir/pathmunge.sh" ]; then
    . "$ansipath_helpers_dir/pathmunge.sh"
fi
unset ansipath_helpers_dir
# <<< ansipath shell helpers <<<
EOF
else
    echo " • Helper load statements already exist in $target_profile"
fi

echo "=================================================="
echo " 🎉 Helper installation completed successfully!"
echo " 👉 Restart your terminal or run: source \"$target_profile\""
echo " 👉 Try ansipath, which, and pathmunge"
echo "=================================================="
