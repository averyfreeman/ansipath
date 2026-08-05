#!/usr/bin/env bash
set -euo pipefail

# The package embeds this file, so update it before uv builds the wheel.
cp -f palettes/palette.json src/ansipath/config/palette.json
uv tool install . --force
