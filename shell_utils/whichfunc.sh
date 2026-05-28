# Overlay 'which' with a smart, colorized ansipath diagnostic fallback
# save file as ~/.local/lib/whichfunc.sh (create folder if necessary)
# 
# ~/.{bash,zsh}rc files re-execute every time a shell is opened, so import
# this in ~/.bash_profile, ~/.profile, or ~/.zprofile (etc.) using snippet:
#
# --- start whichfunc.sh import snippet ---- 
# if ! [ -z "$XDG_USER_LIB" ] && [ "$XDG_USER_LIB" -ne "$HOME/.local/lib" ]; then
#     export XDG_USER_LIB="$HOME/.local/lib:$XDG_USER_LIB";
# else
#     export XDG_USER_LIB="$HOME/.local/lib;
# fi
# if [[ -f "/usr/bin/which" && -f "$XDG_USER_LIB/whichfunc.sh" ]]; then
#     source ${XDG_USER_LIB}/whichfunc.sh
# fi 
# --- finish whichfunc.sh import snippet ---- 

which() {
    # If no arguments are passed, fall back to the standard system 'which'
    if [ -z "$1" ]; then
        command which
        return $?
    fi
    # Execute the real 'which' command, capturing both stdout and stderr
    local raw_output
    raw_output=$(command which "$1" 2>&1)
    local exit_code=$?

    # Pass exit code 1 output to ansipath for formatting
    if [[ "$exit_code" -eq 1 ]]; then
        ansipath "$raw_output"; else
        printf "$raw_output";
    fi

    # Preserve the original exit code of the 'which' command
    return $exit_code
}
