---
title: "fix: README demo video plays wide and full-width"
type: fix
status: active
date: 2026-05-14
---

# fix: README demo video plays wide and full-width

## Summary

The README's demo video does not play and renders narrow because GitHub's HTML sanitizer strips most attributes from `<video>` tags rendered in `.md` files. Replace the `<video>` element with a bare user-attachments URL on its own line so GitHub auto-embeds it as its native wide, responsive video player with controls.

---

## Problem Frame

`README.md:25` uses `<video src="https://github.com/user-attachments/assets/7b6bf18a-2bab-482b-a72d-fac9ab7436c2" width="100%" autoplay loop muted playsinline controls></video>`. GitHub's markdown sanitizer strips `autoplay`, `loop`, `muted`, `playsinline`, and `width` from `<video>` tags in README rendering. The result: the video renders at its intrinsic pixel dimensions (often small/narrow) and does not auto-load with the rich GitHub player experience. GitHub user-attachments URLs auto-embed into a full-width responsive `<video>` player when written as a bare URL on its own line — this is the documented, sanitizer-safe path.

---

## Requirements

- R1. The demo video plays from the README when viewed on github.com.
- R2. The demo video renders wide / full-content-width within GitHub's README column.
- R3. The fix is markdown-only — no asset re-upload, no new tooling.
- R4. The header layout above and `---` divider below remain visually correct.

---

## Scope Boundaries

- Not re-encoding, re-uploading, or relocating the video asset.
- Not introducing a thumbnail-link pattern (e.g., linking to YouTube).
- Not touching unrelated README sections.

---

## Key Technical Decisions

- **Bare URL on its own line, not `<video>` HTML**: GitHub strips video attributes but rich-embeds bare user-attachments URLs as wide responsive video players. The only reliable "wide + plays" path without re-hosting.
- **Drop autoplay/loop/muted**: GitHub never honors these on README videos regardless of attribute presence.

---

## Implementation Units

- U1. **Replace `<video>` tag with bare user-attachments URL**

**Goal:** Make the demo video play and render wide on the GitHub-rendered README.

**Requirements:** R1, R2, R3, R4

**Dependencies:** None

**Files:**
- Modify: `README.md`

**Approach:**
- Replace the single line at `README.md:25` containing the `<video>` element with the bare URL `https://github.com/user-attachments/assets/7b6bf18a-2bab-482b-a72d-fac9ab7436c2` on its own line, preserving the surrounding blank lines and the `---` divider below.

**Test scenarios:**
- Happy path: View the README on github.com on the PR branch — the demo video appears as a wide, full-content-width player with a working play button.
- Edge case: Local markdown preview (e.g., VS Code) shows the URL as a link rather than an embedded video — acceptable; the contract is the GitHub-rendered view.

**Verification:**
- `README.md:25` no longer contains a `<video>` HTML element.
- The bare URL renders as a wide playable video on the GitHub PR preview.

---

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Bare URL renders as a plain link in non-GitHub markdown viewers | Accepted — the README's primary surface is github.com. |
| User-attachments asset is private/expired | Verify in PR preview before merging. |

---

## Sources & References

- Related code: `README.md:25`
