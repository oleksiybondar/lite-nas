#!/usr/bin/env python3
"""Merge packaged JSON defaults into installed object configurations."""

import json
import os
import shutil
import sys
import tempfile
from pathlib import Path
from typing import Any


def merge_values(installed: Any, packaged: Any, path: str = "") -> Any:
    """Add missing packaged values while preserving installed conflicts."""
    if isinstance(installed, dict) and isinstance(packaged, dict):
        for key, value in packaged.items():
            child_path = f"{path}.{key}" if path else key
            if key not in installed:
                installed[key] = value
            elif isinstance(installed[key], dict) and isinstance(value, dict):
                installed[key] = merge_values(installed[key], value, child_path)
            elif installed[key] != value and sys.stdin.isatty():
                installed[key] = resolve_conflict(child_path, installed[key], value)
        return installed
    return installed


def resolve_conflict(path: str, installed: Any, packaged: Any) -> Any:
    """Ask an interactive operator to keep, replace, or recursively merge a conflict."""
    print(f"\nJSON setting conflict: {path}")
    print(f"  Installed: {json.dumps(installed, ensure_ascii=False)}")
    print(f"  Packaged:  {json.dumps(packaged, ensure_ascii=False)}")
    while True:
        choice = input("Choose [u]se installed, use [p]ackaged, or [m]erge: ").strip().lower()
        if choice in {"u", "p", "m"}:
            break
    if choice == "p":
        return packaged
    if choice == "m" and isinstance(installed, dict) and isinstance(packaged, dict):
        return merge_values(installed, packaged, path)
    return installed


def main() -> int:
    """Merge a packaged JSON map into its installed counterpart atomically."""
    if len(sys.argv) != 3:
        raise ValueError("usage: merge-json-config.py INSTALLED PACKAGED")
    target, source = map(Path, sys.argv[1:])
    installed = json.loads(target.read_text(encoding="utf-8"))
    packaged = json.loads(source.read_text(encoding="utf-8"))
    if not isinstance(installed, dict) or not isinstance(packaged, dict):
        raise ValueError("JSON config merge requires object roots")
    merged = merge_values(installed, packaged)
    if merged == json.loads(target.read_text(encoding="utf-8")):
        return 0
    metadata = target.stat()
    backup = target.with_name(target.name + ".pre-merge")
    if not backup.exists():
        shutil.copy2(target, backup)
    descriptor, temporary = tempfile.mkstemp(prefix=".json-config-", dir=target.parent)
    try:
        with os.fdopen(descriptor, "w", encoding="utf-8") as stream:
            json.dump(merged, stream, indent=2, ensure_ascii=False)
            stream.write("\n")
            stream.flush()
            os.fsync(stream.fileno())
        os.chown(temporary, metadata.st_uid, metadata.st_gid)
        os.chmod(temporary, metadata.st_mode)
        os.replace(temporary, target)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
