#!/usr/bin/env python3
"""Skill launcher for the shared gpt-image CLI.

Resolution order:
1. Repo checkout / full plugin install: run the compiled gpt-image binary.
2. Shell has a gpt-image executable: delegate to it.
3. Go toolchain is available: run `go run ./cmd/gpt-image` from the repo root.

This keeps `skills/gpt-image` usable when copied as a standalone skill folder
while preserving one canonical implementation for the Go CLI package.
"""
from __future__ import annotations

import shutil
import subprocess
import sys
from pathlib import Path


def _find_repo_binary(script_path: Path) -> Path | None:
    """Look for a prebuilt gpt-image binary relative to the skill folder."""
    candidates = [
        # Full repo layout: <repo>/skills/gpt-image/scripts/generate.py
        script_path.parents[3] / "gpt-image",
        script_path.parents[3] / "gpt-image.exe",
        # Or a cmd/gpt-image build artifact
        script_path.parents[3] / "cmd" / "gpt-image" / "gpt-image",
        script_path.parents[3] / "cmd" / "gpt-image" / "gpt-image.exe",
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

    binary = _find_repo_binary(script_path)
    if binary is not None:
        return _delegate([str(binary)])

    executable = shutil.which("gpt-image")
    if executable:
        return _delegate([executable])

    go = shutil.which("go")
    if go and len(script_path.parents) > 3:
        repo_root = script_path.parents[3]
        if (repo_root / "go.mod").is_file() and (repo_root / "cmd" / "gpt-image").is_dir():
            return _delegate([go, "run", "./cmd/gpt-image"])

    print(
        "error: could not find the gpt-image CLI backend. Build it first with:\n"
        "  go build -o gpt-image ./cmd/gpt-image\n"
        "or install a prebuilt binary onto PATH, then retry the same command.",
        file=sys.stderr,
    )
    return 2


if __name__ == "__main__":
    sys.exit(main())
