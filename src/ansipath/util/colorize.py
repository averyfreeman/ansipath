def colorize_line(line: str, palette: list[str]) -> str:
    """
    Splits a single path line by slashes and
    applies a horizontal rainbow gradient.
    """
    components: list[str] = line.split("/")
    colored_line: list[str] = []

    for i, comp in enumerate(components):
        color: str = palette[i % len(palette)]
        if i == 0 and not comp:
            continue
        colored_line.append(f"{color}/{comp}")

    return "".join(colored_line)
