import json
import os
import sys
from typing import Any
from importlib.resources import files

from .util.colorize import colorize_line
from .util.diagnostics import check_paths, print_warnings
from .util.sanitize import which


def load_config() -> dict[str, Any]:
    """Safely resolves and parses our asset configuration JSON file."""
    # importlib.resources is the modern standard for safely opening package assets
    config_path = files("ansipath.config").joinpath("palette.json")
    with config_path.open("r", encoding="utf-8") as f:
        return json.load(f)


def main() -> None:
    # 1. Grab raw arguments if passed, otherwise default to active $PATH
    raw_source: str = (
        " ".join(sys.argv[1:]) if len(sys.argv) > 1 else os.environ.get("PATH", "")
    )

    # 2. Sanitize and clean potential 'which' error formatting wrappers
    path_source: str = which(raw_source)

    if not path_source:
        print("Error: No valid path tracking string found.", file=sys.stderr)
        sys.exit(1)

    # 3. Load style configurations from JSON
    colors: dict[str, Any] = load_config()
    palette: list[str] = colors["palette"]
    reset: str = colors["reset"]
    dim_red: str = colors["dim_red"]

    # 4. Process split configurations
    lines: list[str] = [line for line in path_source.split(":") if line]
    duplicates, dead_paths = check_paths(lines)

    # 5. Build and display the rainbow output layout
    output: list[str] = []
    for line in lines:
        exists: bool = os.path.exists(line)
        line_output: str = colorize_line(line, palette) + reset

        if not exists:
            line_output += f" {dim_red}[DEAD PATH]{reset}"
        output.append(line_output)

    print("\n".join(output))

    # 6. Execute reporting warnings block
    print_warnings(duplicates, dead_paths, colors)


if __name__ == "__main__":
    main()
