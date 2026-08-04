#!/usr/bin/env python3

import os
import sys

"""
    usage: 
    $ chmod +x ansipath_minimal.py
    $ ./ansipath_minimal.py
    -- or -- 
    $ python3 -m ansipath_minimal
"""

def main():
    """
        --------- first rendition of this little utility ---------
        This version can actually filter other path strings if given
        another arg string at runtime. No-arg default: $PATH
        Does not include dead path link tests like full version.
    """
    path_source = sys.argv[1] if len(sys.argv) > 1 else os.environ.get("PATH", "")

    if not path_source:
        print("Error: No path string found in $PATH or arguments.", file=sys.stderr)
        sys.exit(1)

    print("Note: Only full version checks for non-existent paths.")
    palette = [
        "\033[38;5;160m",  # Red
        "\033[38;5;208m",  # Orange
        "\033[38;5;220m",  # Yellow
        "\033[38;5;40m",  # Green
        "\033[38;5;51m",  # Cyan
        "\033[38;5;33m",  # Blue
        "\033[38;5;93m",  # Indigo
        "\033[38;5;127m",  # Violet
    ]
    reset = "\033[0m"

    """
       This loop 1. splits the $PATH string by colon (":"),
       2. splits each line by forward slash ("/") for coloring 
       3. tracks horizontal depth to keep columns aligned
       and 4. handles the first ("root") slash gracefully
       Lastly, it 5. joins the output with newline escapes ("\n"")
    """
    lines = path_source.split(":")
    output = []
    for line in lines:
        if not line:
            continue
        components = line.split("/")
        colored_line = []

        for i, comp in enumerate(components):
            color = palette[i % len(palette)]
            if i == 0 and not comp:
                continue
            colored_line.append(f"{color}/{comp}")

        output.append("".join(colored_line) + reset)

    print("\n".join(output))


if __name__ == "__main__":
    main()
