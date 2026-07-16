#!/usr/bin/env python3
"""Render a plain-text file as a terminal-style HTML page for headless-Chrome
screenshot capture.

Usage: python3 render_terminal_html.py <input.txt> <output.html> "<title>"
"""
import html
import sys
from pathlib import Path


def main() -> int:
    if len(sys.argv) != 4:
        print("usage: render_terminal_html.py <input.txt> <output.html> <title>",
              file=sys.stderr)
        return 2

    txt = Path(sys.argv[1]).read_text(encoding="utf-8", errors="replace")
    out = Path(sys.argv[2])
    title = sys.argv[3]

    lines = []
    for raw in txt.splitlines():
        escaped = html.escape(raw)
        css_class = ""
        if escaped.startswith("FAIL"):
            css_class = "fail"
        elif escaped.startswith("OK"):
            css_class = "ok"
        elif escaped.startswith("..."):
            css_class = "note"
        elif "passed" in escaped.lower() and "All" in escaped:
            css_class = "ok"
        elif "failed" in escaped.lower() and "check" in escaped.lower():
            css_class = "fail"
        elif escaped.startswith("[R"):
            css_class = "section"

        if css_class:
            lines.append(f'<span class="{css_class}">{escaped}</span>')
        else:
            lines.append(escaped)

    body = "\n".join(lines) or "&nbsp;"

    html_doc = f"""<!doctype html>
<html><head><meta charset="utf-8"><title>{html.escape(title)}</title>
<style>
  body {{ margin: 0; padding: 24px;
         background: #0d1117; color: #c9d1d9;
         font-family: 'SF Mono', 'Menlo', 'Monaco', 'Consolas', 'DejaVu Sans Mono', monospace;
         font-size: 14px; line-height: 1.45;
         white-space: pre; }}
  h1 {{ font-family: -apple-system, system-ui, sans-serif;
        font-size: 13px; color: #8b949e; font-weight: 600;
        margin: 0 0 18px 0; padding-bottom: 10px;
        border-bottom: 1px solid #30363d; white-space: normal; }}
  .fail    {{ color: #ff7b72; }}
  .ok      {{ color: #56d364; }}
  .note    {{ color: #d29922; }}
  .section {{ color: #79c0ff; font-weight: bold; }}
</style></head>
<body><h1>{html.escape(title)}</h1>
{body}</body></html>"""
    out.write_text(html_doc, encoding="utf-8")
    return 0


if __name__ == "__main__":
    sys.exit(main())
