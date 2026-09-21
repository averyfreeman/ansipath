#!/bin/sh

# pathmunge changes PATH in the current shell.
# It remains a sourced adapter because a child process cannot modify its
# parent shell's environment.
pathmunge() {
    if [ "$#" -ne 2 ]; then
        echo "Usage: pathmunge <prepend|append|remove> <directory_path>" >&2
        return 1
    fi

    ansipath_action=$1
    ansipath_target=$2
    case "$ansipath_action" in
        prepend|append|remove) ;;
        *)
            echo "pathmunge error: Invalid keyword '$ansipath_action'." >&2
            echo "Must be one of: prepend, append, remove" >&2
            return 1
            ;;
    esac

    if [ -d "$ansipath_target" ]; then
        ansipath_path=$(cd -- "$ansipath_target" 2>/dev/null && pwd -P) || ansipath_path=$ansipath_target
    else
        ansipath_path=$ansipath_target
    fi
    case "$ansipath_path" in
        /) ;;
        */) ansipath_path=${ansipath_path%/} ;;
    esac

    if [ "$ansipath_action" = remove ]; then
        ansipath_found=0
        ansipath_new_path=
        ansipath_old_ifs=$IFS
        IFS=:
        for ansipath_entry in ${PATH-}; do
            if [ "$ansipath_entry" = "$ansipath_path" ]; then
                ansipath_found=1
                continue
            fi
            if [ -n "$ansipath_new_path" ]; then
                ansipath_new_path=$ansipath_new_path:$ansipath_entry
            else
                ansipath_new_path=$ansipath_entry
            fi
        done
        IFS=$ansipath_old_ifs
        if [ "$ansipath_found" -eq 0 ]; then
            return 5
        fi
        PATH=$ansipath_new_path
        export PATH
        return 4
    fi

    if [ ! -d "$ansipath_path" ]; then
        echo "pathmunge warning: Directory '$ansipath_path' does not exist. Skipping." >&2
        return 2
    fi
    case ":${PATH-}:" in
        *:"$ansipath_path":*) return 3 ;;
    esac

    if [ "$ansipath_action" = append ]; then
        if [ -n "${PATH-}" ]; then
            PATH=$PATH:$ansipath_path
        else
            PATH=$ansipath_path
        fi
    else
        if [ -n "${PATH-}" ]; then
            PATH=$ansipath_path:$PATH
        else
            PATH=$ansipath_path
        fi
    fi
    export PATH
    return 0
}

report_pathmunge_status() {
    ansipath_status=$1
    ansipath_path_name=$2
    case "$ansipath_status" in
        0) echo "Success: Added '$ansipath_path_name' to \$PATH." ;;
        2) echo "Error 2: Directory '$ansipath_path_name' does not exist." >&2 ;;
        3) echo "Notice: '$ansipath_path_name' is already present in \$PATH." ;;
        4) echo "Success: Removed '$ansipath_path_name' from \$PATH." ;;
        5) echo "Notice: '$ansipath_path_name' was not found in \$PATH." ;;
        *) echo "Error 1: pathmunge failed for unknown reasons on '$ansipath_path_name'." >&2 ;;
    esac
}
