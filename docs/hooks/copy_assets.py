"""MkDocs hook: copy img/ into docs/wiki/assets before each build."""

from __future__ import annotations

import shutil
from pathlib import Path


def on_pre_build(config, **kwargs) -> None:
    root = Path(config.config_file_path).parent
    src = root / "img"
    dest = Path(config.docs_dir) / "assets"
    dest.mkdir(parents=True, exist_ok=True)

    logo = src / "Logo.svg"
    if logo.is_file():
        shutil.copy2(logo, dest / "Logo.svg")

    panel_src = src / "panel"
    panel_dest = dest / "panel"
    if panel_src.is_dir():
        shutil.copytree(panel_src, panel_dest, dirs_exist_ok=True)
