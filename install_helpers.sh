#!/bin/bash

# 1. Figure out which shell we're using
CURRENT_SHELL=$(basename "$SHELL")
echo " • Detected active runtime environment shell: $CURRENT_SHELL"

TARGET_PROFILE=""

case "$CURRENT_SHELL" in
    bash)
        # Choose .bash_profile if it exists, otherwise fall back to .bashrc
        [ -f "$HOME/.bash_profile" ] && TARGET_PROFILE="$HOME"/.bash_profile" || "TARGET_PROFILE="$HOME/.bashrc"
        ;;
    zsh)
        # Choose .zprofile if it exists, otherwise fall back to .zshrc
        [ -f "$HOME/.zprofile" ] && TARGET_PROFILE="$HOME/.zprofile" || TARGET_PROFILE="$HOME/.zshrc"
        ;;
    *)
        # Ensure the fallback profile file physically exists
        [ -f "$HOME/.profile" ] || touch "$HOME/.profile"
        TARGET_PROFILE="$HOME/.profile"
        ;;
esac

export PROFILE_FILE="$TARGET_PROFILE"
echo " • Selected profile initialization target: $TARGET_PROFILE"


# Old logic (nested if statements) 
# use only as fallback if above case statement fails
#TARGET_PROFILE=""
#if [ "$CURRENT_SHELL" = "bash" ]; then
#    if [ -f "$HOME/.bash_profile" ]; then
#        TARGET_PROFILE="$HOME/.bash_profile"
#    else
#        TARGET_PROFILE="$HOME/.bashrc"
#    fi
#elif [ "$CURRENT_SHELL" = "zsh" ]; then
#    if [ -f "$HOME/.zprofile" ]; then
#        TARGET_PROFILE="$HOME/.zprofile"
#    else
#        TARGET_PROFILE="$HOME/.zshrc"
#    fi
#else
#    # null case if shell POSIX or unknown
#    if ! [ -f "$HOME/.profile" ]; then
#        touch "$HOME/.profile";
#    fi
#    TARGET_PROFILE="$HOME/.profile"
#fi
#echo " • Selected profile initialization target: $TARGET_PROFILE"

# 2. Copy pathmunge.sh and whichfunc.sh to $LIBRARY_DIR/shell_helpers 

LIBRARY_DIR="${XDG_DATA_DIR}/ansipath"
echo " 📂 Creating internal tracking directory at $LIBRARY_DIR..."
if ! [ -d "$LIBRARY_DIR" ]; then
    mkdir -pv "$LIBRARY_DIR"
    cp -rv shell_helpers "$LIBRARY_DIR"
fi

echo " 📝 Writing helper load statements to shell profile file..."
cat << 'EOF' >> "$TARGET_PROFILE"
if [[ -f "HOME/.local/share/ansipath/shell_helpers/whichfunc.sh" ]]; then
    source $HOME/.local/share/ansipath/shell_helpers/whichfunc.sh
fi

if [[ -f "HOME/.local/share/ansipath/shell_helpers/pathmunge.sh" ]]; then
    source $HOME/.local/share/ansipath/shell_helpers/pathmunge.sh
fi
EOF

# check for failures, report success if 0 and exit 0
FAILURE="$?"
if [[ -z "$FAILURE" || -e "$?" -eq 0 ]]; then 
    echo "=================================================="
    echo " 🎉 Installation completed successfully!"
    echo " 👉 Restart your terminal or run: `source $TARGET_PROFILE`"
    echo " 👉 Run `ansipath` anywhere for pretty, organized path"
    echo " 👉 Try `which` and `pathmunge`, too"
    echo "=================================================="
else
    echo "exit code $FAILURE"
    exit $FAILURE
fi

exit "$?"
