---
name: owner-working-style
description: How the project owner wants agents to work — confirmation before design, questions, repo-local memory
metadata:
  type: feedback
---
Never design or build something without confirmation or a proposed mockup first. Ask as many questions as needed;
record unanswered ones in docs/design/open-questions.md and resolve them when they become relevant.
Keep memory in this repo, not the Claude account (account may change).

**Why:** owner stated this in the original spec (docs/design/00-original-spec.md) and in the kickoff session (2026-11-25).
**How to apply:** design docs start as `Status: Proposed`; implement only `Accepted`. Use AskUserQuestion / PR review for sign-off.
