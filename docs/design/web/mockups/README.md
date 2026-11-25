# UI mockups (Claude design canvas mirror)

These files mirror the **"anyRiichi UI mockups"** design canvas, a claude.ai Artifact of type *Design*:
https://claude.ai/artifact/7Xck63BrfEFUchViwVWgRv (private to the owner's Claude account).

The repo copy is the **source of truth for migration**: if the project moves to another Claude account,
the artifact link won't follow, so re-create the canvas from these files.

## Contents
| Path | What |
|---|---|
| `project/canvas.json` | Canvas index: artboard positions/sizes/titles, notes |
| `project/Main.dc.html` | Player page — desktop (interactive: platform filter, stat tabs) |
| `project/PlayerMobile.dc.html` | Player page — phone |
| `project/Accounts.dc.html` | Settings — linked accounts, Tenhou verification, log privacy, sign-in |
| `project/Upload.dc.html` | Submit game logs (Tenhou + Mahjong Soul), missing-log list, extension |
| `project/SignIn.dc.html` | Sign in / create account, onboarding (link Tenhou name, log privacy) — proposed 2026-11-25 |
| `assets/` | Uploaded image assets (platform icons) referenced by the artboards |

### Asset map
Artboards reference uploaded assets by artifact-local blob URL. After re-uploading to a new canvas,
replace these URLs in the `.dc.html` files with the new ones the upload returns:

| Blob URL in current canvas | Repo file |
|---|---|
| `/_blob/9b1188e6dd8efb5c77be8b86131b3356` | `assets/mahjongsoul-icon.png` (official Mahjong Soul app icon) |
| `/_blob/457615ed98121fa99d4278456336ae89` | `assets/tenhou-icon.png` (Tenhou favicon) |

## Re-creating the canvas in another account (ask Claude to do this)
1. `Artifact` quickstart with intent `design` → publish with the Design `type_url`, title "anyRiichi UI mockups".
2. Upload `assets/*.png` to the new artifact (`asset: true`), note the returned `/_blob/...` URLs, update the `.dc.html` files.
3. Publish `project/canvas.json` + all `project/*.dc.html` to the new artifact URL (`root` = this folder).
4. Update the link at the top of this README and in `docs/design/web/user-interface.md`.

## Keeping in sync
Whenever the canvas is edited (by Claude or by hand in the browser), pull the latest files back here
(`Artifact read` with `path: project/<file>`) and commit them. Comment threads on the canvas are **not**
mirrored — decisions from them are recorded in `docs/design/decisions.md`.
