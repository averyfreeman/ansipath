#!/bin/sh

if command -v unalias >/dev/null 2>&1; then
    unalias which 2>/dev/null || :
fi

which() {
    if [ "$#" -eq 0 ]; then
        command which
        return "$?"
    fi

    command which "$1" >/dev/null 2>&1
    ansipath_status=$?
    if [ "$ansipath_status" -eq 1 ]; then
        ansipath_real_which=$(command -v which 2>/dev/null || printf '%s' "${ANSIPATH_BIN:-ansipath}")
        printf '%s: no %s in:\\n' "$ansipath_real_which" "$1"
        "${ANSIPATH_BIN:-ansipath}"
    else
        command which "$1"
    fi
    return "$ansipath_status"
}
