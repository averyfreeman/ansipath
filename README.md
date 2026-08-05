# ansipath 🌈

Discover `$PATH` legibility: print each path on a new line in a rainbow of colors, so you can actually read it.

```
leg•i•bil•i-ty | leja'bilade | noun
the quality of being clear enough to read: we've increased the type size for
greater legibility; design and layout clearly affect the legibility of the text.
Syns: READABILITY, clarity, readableness, ease of reading
```

## Features

- Colorizes each `$PATH` entry by directory depth.
- Flags duplicate entries and directories that no longer exist.
- Includes `which` and `pathmunge` shell helpers for Bash and Zsh.

## Installation

Install [uv](https://docs.astral.sh/uv/getting-started/installation/), then run these commands from the repository root:

The default build uses `catppuccin_mocha`. To choose another palette, copy it to
`palettes/palette.json` before building.

```bash
cp -v palettes/catppuccin_mocha.json palettes/palette.json
./scripts/build.sh
bash install_helpers.sh
```

The helper installer supports Bash and Zsh. Restart your shell, or source the profile file reported by the installer.

## Usage

1. **`ansipath` command**

   Print the current `$PATH`, one entry per line.

   ```bash
   ansipath
   ```

   ![ansipath output](./screenshots/ansipath_example.png)

2. **`which` helper**

   Run `which` as usual. If a command is missing, the installed helper prints the lookup error and runs `ansipath` automatically.

   ```bash
   which missing_command
   ```

   ![which helper output](./screenshots/whichfunc_example.png)

3. **`pathmunge` helper**

   Prepend, append, or remove a directory from `$PATH` in the current shell.

   ```bash
   pathmunge prepend ~/.local/bin
   ```

   ![pathmunge helper output](./screenshots/pathmunge_example.png)

## Development

```bash
uv run mypy src/ tests/
uv run pytest
```

## Palette file viewer

 ```bash
 cd palettes
 chmod +x palette_viewer.py
 bash palette_viewer.py stonewall.json 
 ```

 ![pathmunge helper output](./screenshots/palette_viewer_example.png)

