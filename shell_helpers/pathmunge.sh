# ### USAGE EXAMPLES ###
#
# # How to load `pathmunge` command:
# #     Add these 3 lines your $HOME/.zprofile or $HOME/.bash_profile:
#
# if [[ -f "HOME/.local/share/ansipath/shell_helpers/pathmunge.sh" ]]; then
#     source $HOME/.local/share/ansipath/shell_helpers/pathmunge.sh
# fi
#
# # Explanation: 
# #     `pathmunge` is a function stored inside this file that will
# #     only load on login if called from `.{z,bash_}profile`. 
# #     The `if` condition confirms the file exists in that location
# #     before attempting to `source`.
#  
# #     To load the function without logging out and back in, you can
# #     re-run the inside of the `if` block in your terminal:
#
# source $HOME/.local/share/ansipath/shell_helpers/pathmunge.sh 
#
# # Important: do NOT put `$PATH` manipulators in `.zshrc` or `.bashrc`, as 
# # these will re-run every time you open a new terminal tab or window --
#
# # Uncertainty about which command is running can have devastating consequences!
# # Tip: Invoke the `which [ command ]` helper to be certain.

# # Important behaviors to be aware of: 
# #     `$PATH` is read sequentally from left to right.  If a command
# #     is located in a directory on the left, command will run and 
# #     interpreter stop searching. If two commands of same name exist, 
# #     right-directory command only run via explicit location. 
#
# # Practical example:
# #     Pretend a binary named `buzz` exists in the following two directories:
# #         `/usr/bin` 
# #         `/opt/buzz_suite/bin`
#
# #     The simplest $PATH you could have would be:
# #         `PATH=/opt/buzz_new/bin:/usr/bin`
#
# #     In this example, running `buzz` without the location will find 
# #     `/opt/buzz_suite/bin/buzz`, and it will always take precedence over
# #     `/usr/bin/buzz`. 
#
# #     The only way to run the `buzz` command located at `/usr/bin/buzz` in
# #     our example would be to use the full path and type `/usr/bin/buzz`.
#
# # Tip: Use this knowledge for configuring PATH to avoid unexpected behavior
#
# # Interactive usage:
# # Prepend -- affix directory to the left:
# pathmunge prepend $HOME/.local/bin
# report_pathmunge_status $? $HOME/.local/bin
#
# # Append -- affix directory to the right:
# pathmunge append /opt/custom/bin
# report_pathmunge_status $? /opt/custom/bin

# # Remove -- stop running commands in dir:
# pathmunge remove /usr/local/games
# report_pathmunge_status $? /usr/local/games


pathmunge () {
    # --- GUARD CLAUSES (Input Validation) ---
    
    # Guard 1: Ensure both arguments are provided
    if [ -z "$1" ] || [ -z "$2" ]; then
        echo "Usage: pathmunge <prepend|append|remove> <directory_path>" >&2
        return 1
    fi

    # Guard 2: Validate that the keyword is correct
    case "$1" in
        prepend|append|remove) ;; # Valid keyword, continue
        *)
            echo "pathmunge error: Invalid keyword '$1'." >&2
            echo "Must be one of: prepend, append, remove" >&2
            return 1
            ;;
    esac

    local action="$1"
    local target_path="$2"

    # --- PORTABLE PATH NORMALIZATION (Bash & Zsh / Linux & macOS) ---
    local abs_path
    if [ -d "$target_path" ]; then
        # If the directory exists, resolve its absolute physical path safely
        abs_path=$(cd "$target_path" && pwd -P)
    else
        # If it doesn't exist, we can't 'cd' into it.
        # If adding, this will fail Guard 3. If removing, we clean the raw string.
        abs_path="$target_path"
    fi

    # Strip trailing slash for consistency (unless it's the root directory "/")
    if [ "$abs_path" != "/" ]; then
        abs_path="${abs_path%/}"
    fi

    # --- REMOVE LOGIC ---
    if [ "$action" = "remove" ]; then
        case ":${PATH}:" in
            *:"$abs_path":*)
                # Match found: safely strip it out using shell parameter expansion
                PATH=":${PATH}:"
                PATH="${PATH//:$abs_path:/:}"
                PATH="${PATH#:}" # Strip leading colon
                PATH="${PATH%:}" # Strip trailing colon
                export PATH
                return 4
                ;;
            *)
                # Path wasn't in $PATH anyway
                return 5
                ;;
        esac
    fi

    # --- ADD LOGIC (prepend / append) ---
    # Guard 3: Ensure the target folder physically exists on the disk (Only for adding)
    if [ ! -d "$abs_path" ]; then
        echo "pathmunge warning: Directory '$abs_path' does not exist. Skipping." >&2
        return 2
    fi

    case ":${PATH}:" in
        *:"$abs_path":*)
            # Match found! Do nothing and exit cleanly
            return 3
            ;;
        *)
            # No match found: safely append or prepend the new block
            if [ "$action" = "append" ] ; then
                PATH="$PATH:$abs_path"
            else
                PATH="$abs_path:$PATH"
            fi
            export PATH
            return 0
            ;;
    esac
}

report_pathmunge_status() {
    local status_code="$1"
    local path_name="$2"

    case "$status_code" in
        0)
            echo "Success: Added '$path_name' to \$PATH."
            ;;
        2)
            echo "Error 2: Directory '$path_name' does not exist." >&2
            ;;
        3)
            echo "Notice: '$path_name' is already present in \$PATH."
            ;;
        4)
            echo "Success: Removed '$path_name' from \$PATH."
            ;;
        5)
            echo "Notice: '$path_name' was not found in \$PATH."
            ;;
        1|*)
            echo "Error 1: pathmunge failed for unknown reasons on '$path_name'." >&2
            ;;
    esac
}
