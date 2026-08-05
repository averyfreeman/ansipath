#!/usr/bin/env python3

css_vars = {
    "--md-default-bg-color": "#030303",
    "--md-default-fg-color": "#6272a4",
    "--md-default-fg-color--light": "#030303",
    "--md-primary-fg-color": "#151515",
    "--md-primary-fg-color--light": "#304052",
    "--md-primary-fg-color--dark": "#f6f6f6",
    "--md-accent-fg-color": "#ff79fa",
    "--md-typeset-a-color": "#8be9fd",
    "--md-code-bg-color": "#44475a",
    "--md-code-fg-color": "#fff830",
    "--md-footer-bg-color": "#21222c",
    "--md-footer-bg-color--dark": "#363636"
}

def hex_to_rgb(hex_code):
    """Convert a hex color string to an RGB tuple."""
    hex_code = hex_code.lstrip('#')
    return tuple(int(hex_code[i:i+2], 16) for i in (0, 2, 4))

def main():
    print("Dracula Material for MkDocs Palette:\n")
    
    for name, hex_val in css_vars.items():
        r, g, b = hex_to_rgb(hex_val)
        
        # ANSI 24-bit background color escape sequence
        ansi_bg = f"\033[48;2;{r};{g};{b}m"
        reset = "\033[0m"
        
        # Print a solid block of spaces with the background color applied
        print(f"{ansi_bg}          {reset} {hex_val} | {name}")

if __name__ == "__main__":
    main()