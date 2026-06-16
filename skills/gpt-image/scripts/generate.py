#!/usr/bin/env python3
"""Skill launcher for the shared gpt-image CLI.

Resolution order:
1. Prebuilt binary inside the skill folder (scripts/, skill root, or cmd/).
2. A `gpt-image` executable on PATH.

Agent installers should download the appropriate prebuilt binary from the
GitHub Release and place it next to this script or in the skill root.
"""
from __future__ import annotations

import shutil
import subprocess
import sys
from pathlib import Path


def _find_skill_binary(script_path: Path) -> Path | None:
    """Look for a prebuilt gpt-image binary inside the skill folder."""
    scripts_dir = script_path.parent
    skill_root = scripts_dir.parent
    candidates = [
        scripts_dir / "gpt-image",
        scripts_dir / "gpt-image.exe",
        skill_root / "gpt-image",
        skill_root / "gpt-image.exe",
        skill_root / "cmd" / "gpt-image",
        skill_root / "cmd" / "gpt-image.exe",
    ]
    for c in candidates:
        if c.is_file():
            return c
    return None


def _delegate(command: list[str]) -> int:
    """Run another CLI process with the original argv and return its exit code."""
    completed = subprocess.run(command + sys.argv[1:], check=False)
    return completed.returncode


def main() -> int:
    script_path = Path(__file__).resolve()

    binary = _find_skill_binary(script_path)
    if binary is not None:
        return _delegate([str(binary)])

    executable = shutil.which("gpt-image")
    if executable:
        return _delegate([executable])

    print(
        "error: could not find the gpt-image CLI backend. Install a prebuilt binary\n"
        "into this skill folder (next to generate.py or in the skill root), put it\n"
        "on PATH, then retry the same command.\n"
        "Download: https://github.com/ZacharyJia/gpt-image2-cli/releases/latest",
        file=sys.stderr,
    )
    return 2


if __name__ == "__main__":
    sys.exit(main())
