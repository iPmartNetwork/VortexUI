#!/usr/bin/env python3
"""Build GitHub Pages site: MkDocs wiki + React landing at /."""

from __future__ import annotations

import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
LANDING = ROOT / "docs" / "landing"
SITE = ROOT / "site"
WIKI_HOME = SITE / "wiki-home"


def run(cmd: list[str], cwd: Path | None = None) -> None:
    print("+", " ".join(cmd), flush=True)
    subprocess.run(cmd, cwd=cwd, check=True)


def main() -> int:
    if SITE.exists():
        shutil.rmtree(SITE)

    run([sys.executable, "-m", "mkdocs", "build", "-f", str(ROOT / "mkdocs.yml")])

    wiki_index = SITE / "index.html"
    if wiki_index.is_file():
        WIKI_HOME.mkdir(parents=True, exist_ok=True)
        shutil.move(str(wiki_index), WIKI_HOME / "index.html")

    npm = shutil.which("npm")
    if not npm:
        print("npm not found; skipping landing page", file=sys.stderr)
        return 1

    if not (LANDING / "node_modules").is_dir():
        run([npm, "ci"], cwd=LANDING)
    run([npm, "run", "build"], cwd=LANDING)

    landing_index = LANDING / "dist" / "index.html"
    if not landing_index.is_file():
        print(f"missing {landing_index}", file=sys.stderr)
        return 1

    shutil.copy2(landing_index, SITE / "index.html")
    (SITE / ".nojekyll").write_text("", encoding="utf-8")
    print(f"built {SITE} ({(SITE / 'index.html').stat().st_size} bytes landing)", flush=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
