def which(raw_input: str) -> str:
    """Strips away 'which' error prefixes and matching parentheses to extract clean paths."""
    text: str = raw_input.strip()

    # Check if the string is wrapped in 'which: no command in (...)'
    if "no " in text and " in (" in text:
        # Extract everything strictly between the first '(' and the trailing ')'
        start_idx: int = text.find("(") + 1
        end_idx: int = text.rfind(")")
        if start_idx > 0 and end_idx > start_idx:
            text = text[start_idx:end_idx]

    # Fallback cleanup for any miscellaneous parenthetical encapsulation
    return text.strip("()")
