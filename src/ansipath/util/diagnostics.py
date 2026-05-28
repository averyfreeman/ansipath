import os
from collections import Counter


def check_paths(lines: list[str]) -> tuple[dict[str, int], list[str]]:
    """Analyzes paths to isolate duplicates and locate missing directories on disk."""
    path_counts: Counter[str] = Counter(lines)
    duplicates: dict[str, int] = {
        path: count for path, count in path_counts.items() if count > 1
    }

    dead_paths: list[str] = [line for line in lines if not os.path.exists(line)]
    return duplicates, dead_paths


def print_warnings(
    duplicates: dict[str, int], dead_paths: list[str], colors: dict[str, str]
) -> None:
    """Prints human-readable terminal alerts for path optimization opportunities."""
    reset: str = colors["reset"]

    if duplicates:
        print(
            f"\n{colors['yellow_bold']}⚠️  Warning: Duplicate entries detected in your $PATH:{reset}"
        )
        for path, count in duplicates.items():
            print(f"  • {path} ({count} times)")

    if dead_paths:
        unique_dead: list[str] = sorted(list(set(dead_paths)))
        print(
            f"\n{colors['dim_red']}✖  Alert: Broken paths found (directories do not exist):{reset}"
        )
        for path in unique_dead:
            print(f"  • {path}")
