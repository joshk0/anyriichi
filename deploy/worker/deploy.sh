#!/usr/bin/env bash
# Build the worker for linux/amd64 and install it on the GCE VM over IAP SSH.
# Usage: deploy/worker/deploy.sh [--env]   (--env also uploads worker env from local .env)
# Requires gcloud signed in to the owner's PERSONAL account with project "anyriichi".
set -euo pipefail
PROJECT=${GCP_PROJECT:-anyriichi}
ZONE=${GCP_ZONE:-us-east1-b}
VM=${GCP_VM:-anyriichi-worker}
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
TMP=$(mktemp -d); trap 'rm -rf "$TMP"' EXIT

acct=$(gcloud config get account 2>/dev/null)
case "$acct" in *@ternary.app) echo "refusing: gcloud is on a work account ($acct)" >&2; exit 1;; esac

(cd "$ROOT/worker" && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o "$TMP/worker" ./cmd/worker)
cp "$ROOT/deploy/worker/anyriichi-worker.service" "$TMP/"

if [[ "${1:-}" == "--env" ]]; then
  # Only worker-relevant keys; never the Supabase CLI token.
  grep -E '^(DATABASE_URL|HTTP_ADDR|LOG_LEVEL|TENHOU_POLL_INTERVAL|TENHOU_BACKFILL|TENHOU_FETCH_HOUOU_LOGS|USER_AGENT|MAJSOUL_ACCOUNT|MAJSOUL_PASSWORD)=' "$ROOT/.env" > "$TMP/worker.env"
fi

G=(--project "$PROJECT" --zone "$ZONE" --tunnel-through-iap)
gcloud compute scp "${G[@]}" "$TMP"/* "$VM":/tmp/
gcloud compute ssh "${G[@]}" "$VM" --command 'set -e
sudo id anyriichi >/dev/null 2>&1 || sudo useradd --system --no-create-home --shell /usr/sbin/nologin anyriichi
sudo install -d -m 755 /opt/anyriichi
sudo install -d -m 750 -g anyriichi /etc/anyriichi
sudo install -m 755 /tmp/worker /opt/anyriichi/worker
sudo install -m 644 /tmp/anyriichi-worker.service /etc/systemd/system/
if [ -f /tmp/worker.env ]; then sudo install -m 640 -g anyriichi /tmp/worker.env /etc/anyriichi/worker.env; fi
sudo test -f /etc/anyriichi/worker.env || sudo install -m 640 -g anyriichi /dev/null /etc/anyriichi/worker.env
rm -f /tmp/worker /tmp/worker.env /tmp/anyriichi-worker.service
sudo systemctl daemon-reload
sudo systemctl enable --now anyriichi-worker
sudo systemctl restart anyriichi-worker
sleep 2
systemctl is-active anyriichi-worker
curl -fsS localhost:8080/healthz && echo'
