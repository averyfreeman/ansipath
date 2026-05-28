#!/bin/sh
# NOTE: We use #!/bin/sh (POSIX standard) so this installer can execute 
# safely on any host shell, even before we verify if Bash is available.

set -e 
printf "
#########################################################
#                                                       #
#      NOTE! YOU SHOULD EXAMINE THIS SCRIPT BEFORE      #
#    RUNNING IT! Everyone's environment is different    #
#    and you never know when something might take a     #
#    shit on you without looking through it first!      #
#                                                       #
#            ---- YOU HAVE BEEN WARNED! ----            #
#                                                       #
#########################################################

**** ATTENTION! SCRIPT MUST BE RUN FROM ANSIPATH ROOT ****
      copy it to ./ansipath/. if you haven't already
"
read -p "Press enter to continue installation process, CTRL-C to abort..."

echo "=================================================="
echo "      🚀 Installing ansipath CLI Utility        "
echo "=================================================="

# 1. Environment and Shell Detection
# Check if 'uv' is installed
if ! command -v uv >/dev/null 2>&1; then
    echo "❌ Error: 'uv' package manager was not found." >&2
    echo "Please install uv first: https://astral.sh" >&2
    exit 1
fi

# Detect what shell the current user session is targetting
# We check $SHELL but fall back to checking the parent process if running nested
CURRENT_SHELL=$(basename "$SHELL")
echo "  • Detected active runtime environment shell: $CURRENT_SHELL"

TARGET_PROFILE=""

case "$CURRENT_SHELL" in
    bash)
        # Choose .bash_profile if it exists, otherwise fall back to .bashrc
        [ -f "$HOME/.bash_profile" ] && TARGET_PROFILE="$HOME/.bash_profile" || TARGET_PROFILE="$HOME/.bashrc"
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

echo "  • Selected profile initialization target: $TARGET_PROFILE"

### Old logic: nested if statements - kept in case
### switch statement inoperable
# Select the appropriate profile configuration target based on the shell type
# TARGET_PROFILE=""
# if [ "$CURRENT_SHELL" = "bash" ]; then
#     if [ -f "$HOME/.bash_profile" ]; then
#         TARGET_PROFILE="$HOME/.bash_profile"
#     else
#         TARGET_PROFILE="$HOME/.bashrc"
#     fi
# elif [ "$CURRENT_SHELL" = "zsh" ]; then
#     if [ -f "$HOME/.zprofile" ]; then
#         TARGET_PROFILE="$HOME/.zprofile"
#     else
#         TARGET_PROFILE="$HOME/.zshrc"
#     fi
# else
#     # Fallback default target for standard POSIX / unknown environments
#     if ! [ -f "$HOME/.profile" ]; then
#         touch $HOME/.profile;
#     fi
#     TARGET_PROFILE="$HOME/.profile"
# fi

echo "  • Selected profile initialization target: $TARGET_PROFILE"


# 2. Build and Install the Python ansipath Tool Globally
echo "📦 Building Python application binary via uv..."
uv tool install . --force


# 3. Create and Write the Portable pathmunge Library
LIBRARY_DIR="$HOME/.local/share/ansipath"
LIBRARY_FILE="$LIBRARY_DIR/pathmunge.sh"

echo "📂 Creating internal tracking directory at $LIBRARY_DIR..."
mkdir -pv "$LIBRARY_DIR"

echo "📝 Writing optimized pathmunge logic to library file..."
cat << 'EOF' > "$LIBRARY_FILE"
# ---------------------------------------------------------------------
# Portable pathmunge engine generated automatically by ansipath installer
# ---------------------------------------------------------------------
pathmunge () {
    if [ -z "$1" ]; then
        echo "pathmunge error: Missing directory argument." >&2
        return 1
    fi

    local target_path="$1"
    if [ -h "$target_path" ]; then
        target_path=$(readlink -f "$target_path")
    fi

    # Cross-shell safe string extraction for absolute paths
    local abs_path
    abs_path=$(cd "$(dirname "$target_path")" 2>/dev/null && pwd)/$(basename "$target_path") 2>/dev/null

    if [ ! -d "$abs_path" ]; then
        echo "pathmunge warning: Directory '$abs_path' does not exist. Skipping." >&2
        return 2
    fi

    # Strict POSIX compatibility logic for matching internal strings
    case ":${PATH}:" in
        *:"$abs_path":*)
            return 3
            ;;
        *)
            if [ "$2" = "after" ] ; then
                PATH="$PATH:$abs_path"
            else
                PATH="$abs_path:$PATH"
            fi
            return 0
            ;;
    esac
}
EOF


# 4. Safely Link the Library to the Shell Profile Configuration
echo "🔗 Registering pathmunge mapping to your profile..."

# Use strict POSIX grep to ensure we do not append multiple sourcing definitions
if grep -q "$LIBRARY_FILE" "$TARGET_PROFILE" 2>/dev/null; then
    echo "  • Library sourcing block already present in $TARGET_PROFILE. Skipping append."
else
    echo "  • Injecting source configuration hooks..."
    cat << EOF >> "$TARGET_PROFILE"

# --- ansipath Shell Helper Configuration Hooks ---
if [ -f "$LIBRARY_FILE" ]; then
    . "$LIBRARY_FILE"
fi
EOF
fi

echo "=================================================="
echo "🎉 Installation completed successfully!"
echo "👉 Restart your terminal session or run: source $TARGET_PROFILE"
echo "👉 Run 'ansipath' anywhere to visualize your clean workspace!"
echo "=================================================="
