---
name: git-workflow-pre-prototype
description: Until a workable prototype exists, push accumulated work directly to main instead of opening many small PRs
metadata:
  type: feedback
---
Until anyRiichi reaches a workable prototype: work in the local tree, let changes accumulate, then commit and push
directly to `main`. No small PRs. After the prototype exists, switch to targeted, readable PRs.

**Why:** owner (2026-11-25) found many small doc PRs unnecessary; also a follow-up commit was pushed to an
already-merged PR branch (#64) and had to be re-PR'd (#65).
**How to apply:** batch related doc/code changes, push to main at sensible checkpoints. When PRs resume, never push
to a branch whose PR is merged — check `gh pr view` first.
