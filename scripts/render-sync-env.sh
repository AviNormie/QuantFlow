#!/usr/bin/env bash
# Sync manual Render env vars and trigger deploys for StockFlow services.
# Prerequisites: render CLI logged in, workspace set, blueprint deployed.
#
# Usage:
#   cp deploy/render.manual.env.example deploy/render.manual.env
#   # edit deploy/render.manual.env
#   ./scripts/render-sync-env.sh

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="${ROOT}/deploy/render.manual.env"
RENDER_CONFIG="${HOME}/.render/cli.yaml"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Missing $ENV_FILE"
  echo "Copy deploy/render.manual.env.example and set FINNHUB_API_KEY and ALLOWED_ORIGINS."
  exit 1
fi

# shellcheck disable=SC1090
source "$ENV_FILE"

if [[ -z "${FINNHUB_API_KEY:-}" ]]; then
  echo "FINNHUB_API_KEY is required in deploy/render.manual.env"
  exit 1
fi

if [[ -z "${ALLOWED_ORIGINS:-}" ]]; then
  echo "ALLOWED_ORIGINS is required in deploy/render.manual.env (your Vercel URL)"
  exit 1
fi

API_KEY="$(awk '/^    key:/{print $2}' "$RENDER_CONFIG")"
if [[ -z "$API_KEY" ]]; then
  echo "Render API key not found. Run: render login"
  exit 1
fi

render workspace set tea-ct3m35jtq21c738tajq0 --confirm >/dev/null 2>&1 || true

put_env() {
  local service_id="$1"
  local key="$2"
  local value="$3"
  local payload
  payload="$(python3 -c 'import json,sys; print(json.dumps({"value": sys.argv[1]}))' "$value")"
  curl -sf -X PUT \
    -H "Authorization: Bearer ${API_KEY}" \
    -H "Content-Type: application/json" \
    -d "$payload" \
    "https://api.render.com/v1/services/${service_id}/env-vars/${key}"
}

deploy_service() {
  local service_id="$1"
  curl -sf -X POST \
    -H "Authorization: Bearer ${API_KEY}" \
    -H "Content-Type: application/json" \
    -d '{}' \
    "https://api.render.com/v1/services/${service_id}/deploys"
}

find_service_id() {
  local name="$1"
  render services --output json --confirm | python3 -c "
import json,sys
data=json.load(sys.stdin)
for item in data:
    s=item.get('service',item)
    if s.get('name')=='${name}':
        print(s['id'])
        break
"
}

SERVICES=(
  stockflow-market
  stockflow-api-gateway
  stockflow-auth
  stockflow-websocket
)

echo "Looking for StockFlow services on Render..."
MISSING=0
for name in "${SERVICES[@]}"; do
  id="$(find_service_id "$name")"
  if [[ -z "$id" ]]; then
    echo "  MISSING: $name"
    MISSING=1
  else
    echo "  FOUND:   $name ($id)"
  fi
done

if [[ "$MISSING" -eq 1 ]]; then
  echo ""
  echo "StockFlow is not deployed yet. Launch the blueprint first:"
  echo "  1. https://dashboard.render.com/blueprints"
  echo "  2. New Blueprint Instance → connect AviNormie/QuantFlow → branch main"
  echo "  3. When prompted, enter FINNHUB_API_KEY and ALLOWED_ORIGINS"
  echo "  4. Re-run: ./scripts/render-sync-env.sh"
  exit 1
fi

MARKET_ID="$(find_service_id stockflow-market)"
GATEWAY_ID="$(find_service_id stockflow-api-gateway)"

echo ""
echo "Pushing environment variables..."

put_env "$MARKET_ID" "FINNHUB_API_KEY" "$FINNHUB_API_KEY"
put_env "$GATEWAY_ID" "ALLOWED_ORIGINS" "$ALLOWED_ORIGINS"

if [[ -n "${SENTRY_DSN:-}" ]]; then
  for name in stockflow-api-gateway stockflow-auth stockflow-market stockflow-websocket; do
    id="$(find_service_id "$name")"
    put_env "$id" "SENTRY_DSN" "$SENTRY_DSN"
    put_env "$id" "SENTRY_ENVIRONMENT" "${SENTRY_ENVIRONMENT:-production}"
    put_env "$id" "SENTRY_RELEASE" "${SENTRY_RELEASE:-stockflow@production}"
    put_env "$id" "SENTRY_TRACES_SAMPLE_RATE" "${SENTRY_TRACES_SAMPLE_RATE:-0.2}"
  done
fi

if [[ -n "${POSTHOG_API_KEY:-}" ]]; then
  for name in stockflow-api-gateway stockflow-auth stockflow-market stockflow-websocket; do
    id="$(find_service_id "$name")"
    put_env "$id" "POSTHOG_API_KEY" "$POSTHOG_API_KEY"
    put_env "$id" "POSTHOG_HOST" "${POSTHOG_HOST:-https://us.i.posthog.com}"
  done
fi

echo "Triggering deploys..."
for name in "${SERVICES[@]}"; do
  id="$(find_service_id "$name")"
  deploy_service "$id" >/dev/null
  echo "  deployed: $name"
done

GATEWAY_URL="$(render services --output json --confirm | python3 -c "
import json,sys
data=json.load(sys.stdin)
for item in data:
    s=item.get('service',item)
    if s.get('name')=='stockflow-api-gateway':
        print(s.get('serviceDetails',{}).get('url',''))
")"
WS_URL="$(render services --output json --confirm | python3 -c "
import json,sys
data=json.load(sys.stdin)
for item in data:
    s=item.get('service',item)
    if s.get('name')=='stockflow-websocket':
        print(s.get('serviceDetails',{}).get('url',''))
")"

echo ""
echo "Backend URLs:"
echo "  API:       ${GATEWAY_URL}"
echo "  WebSocket: ${WS_URL}"
echo ""
echo "Set on Vercel (then redeploy frontend):"
echo "  NEXT_PUBLIC_API_URL=${GATEWAY_URL}"
echo "  NEXT_PUBLIC_WEBSOCKET_URL=${WS_URL}"
echo ""
echo "Verify:"
echo "  curl ${GATEWAY_URL}/ready"
