#!/usr/bin/env python3

import os
import sys


def main():
    # Allow filtering or passing paths via args; default to system $PATH
    path_source = sys.argv[1] if len(sys.argv) > 1 else os.environ.get("PATH", "")

    if not path_source:
        print("Error: No path string found in $PATH or arguments.", file=sys.stderr)
        sys.exit(1)

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

    # Break into separate lines by the colon separator
    lines = path_source.split(":")

    output = []
    for line in lines:
        if not line:
            continue
        # Split each individual line by its directory slashes
        components = line.split("/")
        colored_line = []

        for i, comp in enumerate(components):
            # i tracks horizontal depth to keep columns aligned
            color = palette[i % len(palette)]
            if i == 0 and not comp:
                # This handles the leading root slash gracefully
                continue
            colored_line.append(f"{color}/{comp}")

        output.append("".join(colored_line) + reset)

    print("\n".join(output))


if __name__ == "__main__":
    main()
