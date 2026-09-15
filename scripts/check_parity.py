#!/usr/bin/env python3
"""Compare every calendar day and boundary in the shared Go/Rust input range."""
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[1]
with tempfile.TemporaryDirectory(prefix="trading-hour-parity-") as directory:
    outputs = []
    for language, command in [
        ("go", ["go", "run", "./internal/parity"]),
        ("rust", ["cargo", "run", "--locked", "--quiet", "--example", "parity"]),
    ]:
        output = Path(directory) / language
        with output.open("wb") as stream:
            subprocess.run(command, cwd=root, stdout=stream, check=True)
        outputs.append(output)
    with outputs[0].open() as go, outputs[1].open() as rust:
        lines = 0
        while True:
            expected, actual = go.readline(), rust.readline()
            if expected != actual:
                raise SystemExit(f"Go/Rust mismatch at line {lines + 1}:\nGo:   {expected}Rust: {actual}")
            if not expected:
                break
            lines += 1
    print(f"Go/Rust calendar parity passed: {lines} identical records")
