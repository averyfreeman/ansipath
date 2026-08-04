unalias which 2>/dev/null
which () {
    # 1. Handle empty arguments
    if [[ -z "$1" ]]; then
        command which
        return "$?"
    fi

    # 2. Silently check if the command exists
    command which "$1" >/dev/null 2>&1
    local exit_code="$?"

    # 3. Handle Failure (Binary not found)
    if [[ "$exit_code" -eq 1 ]]; then
        # Dynamically locate the system's actual 'which' binary
        local real_which
        real_which=$(command -v which 2>/dev/null || echo "/usr/bin/which")

        # Print the canonical system prefix error message
        printf "%s: no %s in:\n" "$real_which" "$1"

        # Simply execute your custom tool without arguments to display the colorized $PATH
        ansipath
    else
        # 4. Handle Success (Print the path to the found binary)
        command which "$1"
    fi

    return "$exit_code"
}
