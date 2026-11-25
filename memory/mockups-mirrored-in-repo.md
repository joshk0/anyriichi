---
name: mockups-mirrored-in-repo
description: Design-canvas mockups must be mirrored into docs/design/web/mockups for account-migration safety
metadata:
  type: feedback
---
UI mockups live on a claude.ai Design canvas (private to the current Claude account) AND are mirrored in
`docs/design/web/mockups/` (canvas.json, .dc.html artboards, asset PNGs, README with blob→file map and
re-publish steps).

**Why:** owner (2026-11-25) expects a possible Claude account migration; artifacts don't move with the repo.
**How to apply:** after every canvas change, copy the changed files back into the repo mirror and commit. Record
decisions made in canvas comment threads in docs/design/decisions.md (comments are not mirrored). See [[owner-working-style]].
