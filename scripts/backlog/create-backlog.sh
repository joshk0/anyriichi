#!/usr/bin/env bash
# Idempotently create labels, milestones, issues (and optionally a GitHub Project) from backlog.json.
# Usage: scripts/backlog/create-backlog.sh [owner/repo]   (needs gh; project creation needs `gh auth refresh -s project`)
set -euo pipefail
REPO="${1:-joshk0/anyriichi}"
DIR="$(cd "$(dirname "$0")" && pwd)"
J="$DIR/backlog.json"

LABELS="area:web=1d76db area:worker=0e8a16 area:db=5319e7 area:extension=fbca04 area:infra=c5def5
area:shared=bfdadc platform:tenhou=006b75 platform:majsoul=d93f0b platform:riichicity=e99695
type:research=d4c5f9 needs-design=f9d0c4 blocked:owner=b60205 i18n=c2e0c6"
for pair in $LABELS; do gh label create "${pair%%=*}" -R "$REPO" --color "${pair##*=}" --force >/dev/null; done

existing_ms=$(gh api "repos/$REPO/milestones?state=all&per_page=100" --jq '.[].title')
jq -r '.milestones[]' "$J" | while read -r m; do
  grep -qxF "$m" <<<"$existing_ms" || gh api "repos/$REPO/milestones" -f title="$m" >/dev/null
done

existing=$(gh issue list -R "$REPO" --state all --limit 1000 --json title --jq '.[].title')
jq -c '.issues[]' "$J" | while read -r it; do
  t=$(jq -r .title <<<"$it")
  grep -qxF "$t" <<<"$existing" && { echo "skip: $t"; continue; }
  args=(-R "$REPO" --title "$t" --body "$(jq -r .body <<<"$it")" --milestone "$(jq -r .milestone <<<"$it")")
  while read -r l; do args+=(--label "$l"); done < <(jq -r '.labels[]' <<<"$it")
  gh issue create "${args[@]}"
done

if gh auth status 2>&1 | grep -q "'project'"; then
  num=$(gh project list --owner "${REPO%%/*}" --format json --jq '.projects[] | select(.title=="anyRiichi backlog") | .number')
  [ -z "$num" ] && num=$(gh project create --owner "${REPO%%/*}" --title "anyRiichi backlog" --format json --jq .number)
  gh project link "$num" --owner "${REPO%%/*}" --repo "$REPO" || true
  gh issue list -R "$REPO" --state open --limit 1000 --json url --jq '.[].url' | while read -r u; do
    gh project item-add "$num" --owner "${REPO%%/*}" --url "$u" >/dev/null || true
  done
  echo "Project #$num populated."
else
  echo "Skipping project: run 'gh auth refresh -s project' first."
fi
