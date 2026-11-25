---
name: majsoul-via-amae-koromo
description: Owner's data-scope decision — all public Tenhou games; Majsoul via Amae Koromo for registered users unless full logs are needed
metadata:
  type: project
---
Decided 2026-11-25 (see decisions.md D-001): hybrid data scope. Ingest ALL public Tenhou games. For Mahjong Soul,
try to piggyback on Amae Koromo first instead of connecting to Majsoul ourselves; if Amae Koromo can't provide full
game logs, query it only for players registered on anyRiichi. Implement our own Majsoul connection right away if necessary.

**Why:** reduce effort/risk of running a Majsoul crawler (ToS, account bans, protocol churn).
**Status:** owner chose "AK + own fetcher" (D-006): ask SAPikachu for a key; own fetchGameRecord for registered users only.
**How to apply:** research found Amae Koromo serves aggregate stats openly but game lists are CAPTCHA-gated and
full logs are not served at all → full-log stats for Majsoul need our own `fetchGameRecord` client. See
docs/research/amae-koromo.md and [[owner-working-style]].
