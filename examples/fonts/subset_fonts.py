#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Subset one or more OpenType/TrueType fonts with a user-provided charset.

Examples:

    python subset_fonts.py --text-file charset.txt font/*.otf
    python subset_fonts.py --text-file cjk.txt --text-file latin.txt \
        --output-dir font/subset font/*.ttf

Install fonttools before running:

    python -m pip install fonttools

The input fonts and the character list are intentionally supplied by the
caller. This tool does not assume a particular font family or language.
"""

from __future__ import annotations

import argparse
import glob
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Create smaller font files from a user-provided character list."
    )
    parser.add_argument(
        "fonts",
        nargs="+",
        help="font paths or glob patterns, such as font/*.otf",
    )
    parser.add_argument(
        "--text-file",
        required=True,
        type=Path,
        action="append",
        dest="text_files",
        help="UTF-8 character list supplied by the user; repeat to merge files",
    )
    parser.add_argument(
        "--output-dir",
        type=Path,
        help="directory for generated files (defaults to each input directory)",
    )
    parser.add_argument(
        "--keep-hinting",
        action="store_true",
        help="preserve hinting tables instead of removing them",
    )
    return parser.parse_args()


def subset_command() -> list[str]:
    executable = shutil.which("pyftsubset")
    if executable:
        return [executable]

    module_check = subprocess.run(
        [sys.executable, "-m", "fontTools.subset", "--help"],
        capture_output=True,
        text=True,
    )
    if module_check.returncode == 0:
        return [sys.executable, "-m", "fontTools.subset"]

    raise RuntimeError(
        "pyftsubset was not found. Install FontTools with: "
        "python -m pip install fonttools"
    )


def expand_fonts(patterns: list[str]) -> list[Path]:
    result: list[Path] = []
    seen: set[Path] = set()

    for pattern in patterns:
        matches = glob.glob(pattern, recursive=True)
        if not matches and Path(pattern).is_file():
            matches = [pattern]
        if not matches:
            raise FileNotFoundError(f"No font file or match found: {pattern}")

        for match in matches:
            path = Path(match)
            if not path.is_file():
                continue
            resolved = path.resolve()
            if resolved not in seen:
                seen.add(resolved)
                result.append(path)

    if not result:
        raise FileNotFoundError("No font files were found to process")
    return result


def output_path(input_path: Path, output_dir: Path | None) -> Path:
    target_dir = output_dir if output_dir is not None else input_path.parent
    return target_dir / f"{input_path.stem}-Subset{input_path.suffix}"


def read_charset(paths: list[Path]) -> str:
    """Read and merge user-provided character files in first-seen order."""
    characters: dict[str, None] = {}
    for path in paths:
        charset_path = path.resolve()
        if not charset_path.is_file():
            raise FileNotFoundError(f"Character file not found: {charset_path}")
        try:
            text = charset_path.read_text(encoding="utf-8")
        except UnicodeDecodeError as error:
            raise ValueError(f"Character file must be UTF-8: {charset_path}") from error

        # Character files are commonly line-oriented. Treat line breaks as
        # separators while preserving spaces, tabs, and every other character.
        for character in text.lstrip("\ufeff"):
            if character not in "\r\n":
                characters.setdefault(character, None)

    if not characters:
        raise ValueError("Character files do not contain any characters")
    return "".join(characters)


def subset_font(
    command: list[str],
    input_path: Path,
    output_path: Path,
    charset_path: Path,
    keep_hinting: bool,
) -> None:
    output_path.parent.mkdir(parents=True, exist_ok=True)
    args = command + [
        str(input_path),
        f"--text-file={charset_path}",
        f"--output-file={output_path}",
        "--layout-features=*",
    ]
    if not keep_hinting:
        args.append("--no-hinting")

    try:
        subprocess.run(args, check=True, capture_output=True, text=True)
    except subprocess.CalledProcessError as error:
        message = (error.stderr or error.stdout or "Unknown error").strip()
        raise RuntimeError(f"Subsetting failed for {input_path}: {message}") from error

    original_size = input_path.stat().st_size
    subset_size = output_path.stat().st_size
    reduction = (1 - subset_size / original_size) * 100
    print(
        f"{input_path} -> {output_path} "
        f"({subset_size / 1024 / 1024:.2f} MiB, {reduction:.1f}% smaller)"
    )


def main() -> int:
    args = parse_args()
    charset_text = read_charset(args.text_files)
    command = subset_command()
    fonts = expand_fonts(args.fonts)

    # FontTools accepts one --text-file argument. Write the merged input to a
    # temporary UTF-8 file so callers can keep language-specific lists separate.
    with tempfile.TemporaryDirectory(prefix="flowui-font-subset-") as temp_dir:
        charset_path = Path(temp_dir) / "charset.txt"
        charset_path.write_text(charset_text, encoding="utf-8")
        print(f"Using {len(charset_text)} unique characters from {len(args.text_files)} file(s)")
        for input_path in fonts:
            subset_font(
                command,
                input_path,
                output_path(input_path, args.output_dir),
                charset_path,
                args.keep_hinting,
            )
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (FileNotFoundError, RuntimeError, ValueError) as error:
        print(f"Error: {error}", file=sys.stderr)
        raise SystemExit(1)
