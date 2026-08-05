# ansipath 🌈

A lightweight, high-visibility diagnostic terminal visualizer for system paths. It breaks down congested, multi-layered `$PATH` maps into beautifully colorized horizontal rainbow gradients, making locations from deeply-nested toolchain management libraries, like `mise`, `asdf`, and `nvm`, _actually readable_ again!

It also serves as a protective utility, explicitly flagging **duplicate path references** and **dead directories** that no longer exist on your platform.

---

## Features

- 🎨 **BBS-Inspired Rainbow Palette**: Smooth 8-bit ANSI gradient tracking mapped to directory horizontal nesting depth. Relive your inner early 90s! _(or customize to your heart's content)_
- 🛠️ **System Diagnostics**: Built-in safe `collections.Counter` analysis provides duplicate path feedback that could previously have been difficult to see.
- ✖️ **Dead Path Detection**: Non-obtrusive dimmed alerts highlight configuration decrepidation.
- 🧹 **Smart Input Sanitization**: Auto-resolves raw environment strings or verbose wrapper outputs (like `which command` errors).
- ⚡ **Astral uv Orchestration**: Pure type-hinted Python 3.10+ package with a flat, portable POSIX `pathmunge` setup engine.

---

## Installation

### Prerequisites

Project uses **uv**-style `pyproject.toml`:

```bash
curl -LsSf https://astral.sh | sh
```

### Run the Installer

Clone your project directory and run the universal POSIX installer script:

```bash
chmod +x install.sh
./install.sh
```

The installer will automatically detect your shell configuration layout (`bash`, `zsh`, or `nushell`), build the standalone wheel via `uv`, and append the safe `pathmunge` wrapper library to your initialization runtime files.

---

## Usage

### 1. View Active Environment Paths

Simply type the command anywhere to get a clear visual breakdown of your active session:

```bash
ansipath
```

### 2. Diagnose Command Lookups (Command Substitution)

To quickly check exactly why a command might be missing or where it sits in the loop hierarchy, pass it right into the analyzer:

```bash
ansipath "$(which missing_command)"
```

---

## Development Pipeline

Because this project enforces strict software patterns, you can execute linting matrices and unit test suites across isolation bubbles via `uv`:

- **Run Type Check Inferences (`mypy`)**:
  ```bash
  uv run mypy src/ tests/
  ```
- **Execute Automated Test Assertions (`pytest`)**:
  ```bash
  uv run pytest
  ```
