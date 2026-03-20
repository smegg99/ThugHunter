#!/usr/bin/env python3
"""
Generates platform-specific tray icon variants from source SVGs.
For each theme (dark and light), generates one "no badge" icon and four "badge" icons with different badge colors. Badge is a path element in the SVG with inkscape:label="badge".
"""

from __future__ import annotations

import copy
import re
from pathlib import Path
from typing import Any, cast

from lxml import etree as _etree  # type: ignore[import-untyped]

try:
    import cairosvg as _cairosvg  # type: ignore[import-untyped]
except ImportError:
    raise SystemExit("cairosvg is required: pip install cairosvg lxml")

etree: Any = cast(Any, _etree)
cairosvg: Any = cast(Any, _cairosvg)

SCRIPT_DIR = Path(__file__).parent
ICON_PREFIX = "thughunter-tray"

SOURCES = {
    "dark": {
        "badge": SCRIPT_DIR / "thughunter-logo-tray-dark-badge.svg",
        "no-badge": SCRIPT_DIR / "thughunter-logo-tray-dark-no-badge.svg",
    },
    "light": {
        "badge": SCRIPT_DIR / "thughunter-logo-tray-light-badge.svg",
        "no-badge": SCRIPT_DIR / "thughunter-logo-tray-light-no-badge.svg",
    },
}

BADGE_COLORS = {
    "dark": {
        "default": "#FFFFFF",  # Default one, nothings running in the background.
        "color1":  "#42A5F5",  # Scraper run initializing.
        "color2":  "#66BB6A",  # Scraper run in progress.
        "color3":  "#FFA726",  # Scraper run stopping.
    },
    "light": {
        "default": "#FFFFFF",  # Red
        "color1":  "#1E88E5",  # Scraper run initializing.
        "color2":  "#43A047",  # Scraper run in progress.
        "color3":  "#FB8C00",  # Scraper run stopping.
    },
}

INKSCAPE_LABEL = "{http://www.inkscape.org/namespaces/inkscape}label"
SVG_PATH_TAG = "{http://www.w3.org/2000/svg}path"

def parse_svg(path: Path) -> Any:
    return etree.parse(str(path), etree.XMLParser(remove_blank_text=False))


def find_badge(tree: Any) -> Any | None:
    """Return the <path> element whose inkscape:label is 'badge', or None."""
    for el in tree.iter(SVG_PATH_TAG):
        if el.get(INKSCAPE_LABEL) == "badge":
            return el
    return None

def recolor_badge(tree: Any, color: str) -> Any:
    """Deep-copy *tree* and set the badge path's fill to *color*."""
    tree = copy.deepcopy(tree)
    badge = find_badge(tree)
    if badge is None:
        raise ValueError("No <path> with inkscape:label='badge' found in SVG")
    style: str = badge.get("style", "")
    style = re.sub(r"fill:\s*[^;]+", f"fill:{color}", style)
    badge.set("style", style)
    return tree

def svg_bytes(tree: Any) -> bytes:
    return etree.tostring(
        tree, xml_declaration=True, encoding="UTF-8", pretty_print=True
    )


def write_svg(tree: Any, dest: Path) -> None:
    dest.parent.mkdir(parents=True, exist_ok=True)
    dest.write_bytes(svg_bytes(tree))
    print(f"  SVG  {dest.relative_to(SCRIPT_DIR)}")

def write_png(tree: Any, dest: Path, size: int = 32) -> None:
    dest.parent.mkdir(parents=True, exist_ok=True)
    cairosvg.svg2png(
        bytestring=svg_bytes(tree),
        write_to=str(dest),
        output_width=size,
        output_height=size,
    )
    print(f"  PNG  {dest.relative_to(SCRIPT_DIR)}")

def emit(tree: Any, name: str, theme_dir: Path) -> None:
    """Write one icon variant to all three platform output directories."""
    write_png(tree, theme_dir / "windows" / f"{name}.png")
    write_svg(tree, theme_dir / "linux" / f"{name}.svg")
    write_png(tree, theme_dir / "linux-png-fallback" / f"{name}.png")

def main():
    for theme in ("dark", "light"):
        print(f"\n{'=' * 50}\n  Theme: {theme}\n{'=' * 50}")
        src = SOURCES[theme]
        theme_dir = SCRIPT_DIR / theme

        no_badge = parse_svg(src["no-badge"])
        name = f"{ICON_PREFIX}-{theme}"
        print(f"\n[{name}]")
        emit(no_badge, name, theme_dir)

        badge_base = parse_svg(src["badge"])
        for color_key, hex_color in BADGE_COLORS[theme].items():
            suffix = "" if color_key == "default" else f"-{color_key}"
            name = f"{ICON_PREFIX}-{theme}-badge{suffix}"
            print(f"\n[{name}]  badge = {hex_color}")
            colored = recolor_badge(badge_base, hex_color)
            emit(colored, name, theme_dir)

    print(f"\n{'=' * 50}")
    print("Done.  Generated icons under dark/ and light/.\n")
    print("Linux freedesktop icon-theme install paths:")
    print("  SVGs -> /usr/share/icons/hicolor/scalable/status/")
    print("  PNGs -> /usr/share/icons/hicolor/32x32/status/")
    print("  Then:   gtk-update-icon-cache /usr/share/icons/hicolor/")

if __name__ == "__main__":
    main()