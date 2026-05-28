import os
import sys
from unittest.mock import patch

import pytest
from ansipath.main import main
from ansipath.util.colorize import colorize_line
from ansipath.util.diagnostics import check_paths
from ansipath.util.sanitize import which


def test_sanitize_which() -> None:
    """
    Verify that the sanitize module successfully strips `which` wrapper syntax.
    """
    # Test a raw, messy 'which' failure output string
    messy_input = "which: no missingapp in (/usr/bin:/bin)"
    assert which(messy_input) == "/usr/bin:/bin"

    # Test an input that is already perfectly clean
    clean_input = "/usr/bin:/bin"
    assert which(clean_input) == "/usr/bin:/bin"


def test_colorize_line() -> None:
    """
    Verify that path lines are sliced by slashes and individual segments are colored.
    """
    sample_palette: list[str] = ["\033[31m", "\033[32m"]
    sample_line = "/usr/bin"

    result = colorize_line(sample_line, sample_palette)

    # Root slash tracks to index 1, nested bin tracks to index 0
    assert "\033[32m/usr" in result
    assert "\033[31m/bin" in result


def test_diagnostics_check_paths() -> None:
    """
    Verify that diagnostic checkers accurately pull duplicates and dead paths.
    """
    test_lines = ["/usr/bin", "/bin", "/usr/bin", "/fake/ghost/folder"]

    def mock_exists(path: str) -> bool:
        return path in ["/usr/bin", "/bin"]

    with patch("os.path.exists", side_effect=mock_exists):
        duplicates, dead_paths = check_paths(test_lines)

        # Verify duplicate metrics map correctly
        assert "/usr/bin" in duplicates
        assert duplicates["/usr/bin"] == 2

        # Verify ghost tracking isolated the invalid entry
        assert "/fake/ghost/folder" in dead_paths


def test_main_integration_flow(capsys: pytest.CaptureFixture[str]) -> None:
    """Integration test to verify that loading configuration assets and outputs works."""
    test_path = "/usr/bin:/bin"
    
    # Save the original os.path.exists to safely fall back to it
    real_exists = os.path.exists

    # Only mock path validations for our specific mock directories
    def mock_exists(path: str) -> bool:
        if path in ["/usr/bin", "/bin"]:
            return True
        return real_exists(path)  # Let standard package config files lookup normally

    with patch("sys.argv", ["ansipath"]), patch.dict(
        os.environ, {"PATH": test_path}
    ), patch("os.path.exists", side_effect=mock_exists):
        main()
        
        captured = capsys.readouterr()
        output = captured.out

        # Verify critical elements from the loaded JSON palette structure appear
        assert "bin" in output
        assert "\033[0m" in output  # Verify reset code was safely applied
