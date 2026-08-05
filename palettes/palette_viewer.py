#!/usr/bin/env python3
import json
import re
import sys


def extract_strings(data):
    """Recursively extract all strings from a JSON object in order."""
    if isinstance(data, dict):
        for value in data.values():
            yield from extract_strings(value)
    elif isinstance(data, list):
        for item in data:
            yield from extract_strings(item)
    elif isinstance(data, str):
        yield data


def main():
    if len(sys.argv) < 2:
        print(f"Usage: {sys.argv[0]} <path_to_json_file>", file=sys.stderr)
        sys.exit(1)

    # Regex to match ANSI escape codes (CSI sequences like \x1b[...m)
    ansi_regex = re.compile(r"(\x1b|\\033|\\u001b)\[([0-9;]*m)")

    try:
        with open(sys.argv[1], "r", encoding="utf-8") as f:
            file_data = json.load(f)
    except Exception as e:
        print(f"Error reading JSON file: {e}", file=sys.stderr)
        sys.exit(1)

    print("Found ANSI Escape Sequences (Top-Left to Bottom-Right):\n")

    count = 0
    extracted_codes = []
    
    for text in extract_strings(file_data):
        """
            Process every string found in the JSON hierarchy
            Store extracted codes in order for the palette,
            normalize the display representation, and create a safe, 
            self-contained colored preview string. and print
        """
        for match in ansi_regex.finditer(text):
            count += 1
            code_body = match.group(2)  # The parameters and 'm'
            repr_code = f"\\033[{code_body}"
            color_preview = f"\x1b[{code_body}■■■ PREVIEW ■■■\x1b[0m"

            print(
                f"{count:02d}. Code: {repr_code:<12} | Visual: {color_preview}"
            )
            
            extracted_codes.append(code_body)  # for horizontal bar

    if count == 0:
        print("No ANSI escape codes found in the JSON file.")
    else:
        """
            Build and display the bottom left-right array:
            We use a list comprehension to wrap our block characters in each saved ANSI code, 
            making sure to reset (\x1b[0m) at the end of every block to prevent color bleeding.
        """
        print("\nHorizontal Color Palette:")
        horizontal_blocks = " ".join([f"\x1b[{code}■■■■■\x1b[0m" for code in extracted_codes])
        
        print(horizontal_blocks)
        print() # Just a little extra padding at the bottom for terminal cleanliness


if __name__ == "__main__":
    main()