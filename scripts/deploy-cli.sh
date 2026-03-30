#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
#  deploy-cli.sh  — Push a local folder to the Deploy Platform
#
#  Usage:
#    ./scripts/deploy-cli.sh <app-name> <path-to-project-folder>
#
#  Example:
#    ./scripts/deploy-cli.sh my-app ./sample-app
#
#  The folder must contain a Dockerfile. The script zips it and uploads
#  it to the local API server, then polls until the deployment is live.
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

API="${DEPLOY_API_URL:-http://localhost:8080}"
APP_NAME="${1:-}"
SRC_DIR="${2:-}"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; BLUE='\033[0;34m'; NC='\033[0m'

usage() {
  echo -e "Usage: $0 <app-name> <source-directory>"
  echo -e "  app-name   : lowercase letters, numbers, hyphens (3-30 chars)"
  echo -e "  source-dir : directory containing your Dockerfile"
  echo ""
  echo -e "Example: $0 hello-world ./my-project"
  exit 1
}

[[ -z "$APP_NAME" || -z "$SRC_DIR" ]] && usage
[[ ! -d "$SRC_DIR" ]] && { echo -e "${RED}ERROR:${NC} Directory '$SRC_DIR' not found"; exit 1; }
[[ ! -f "$SRC_DIR/Dockerfile" ]] && { echo -e "${YELLOW}WARN:${NC} No Dockerfile found in '$SRC_DIR' — deploy may fail"; }

# Validate app name
if ! echo "$APP_NAME" | grep -qE '^[a-z0-9][a-z0-9-]{1,28}[a-z0-9]$'; then
  echo -e "${RED}ERROR:${NC} Invalid app name '$APP_NAME'"
  echo "  Must be 3-30 chars, lowercase letters/numbers/hyphens, not start/end with hyphen"
  exit 1
fi

# Check API is reachable
if ! curl -sf "$API/health" > /dev/null 2>&1; then
  echo -e "${RED}ERROR:${NC} API not reachable at $API"
  echo "  Make sure the platform is running: ./scripts/start.sh"
  exit 1
fi

# Create zip (exclude common junk)
ZIP_FILE="/tmp/${APP_NAME}-$(date +%s).zip"
echo -e "${CYAN}[1/4]${NC} Packaging ${SRC_DIR}..."
(cd "$SRC_DIR" && zip -r "$ZIP_FILE" . \
  -x "*.git*" \
  -x "node_modules/*" \
  -x "__pycache__/*" \
  -x "*.pyc" \
  -x ".env" \
  -x "dist/*" \
  -x "build/*" \
  -x "*.DS_Store" \
) > /dev/null
ZIP_SIZE=$(du -sh "$ZIP_FILE" | cut -f1)
echo -e "   Packed ${ZIP_SIZE} → $(basename $ZIP_FILE)"

# Upload
echo -e "${CYAN}[2/4]${NC} Uploading to $API ..."
RESPONSE=$(curl -sf -X POST "$API/deploy" \
  -F "app_name=$APP_NAME" \
  -F "source=@$ZIP_FILE" \
  2>&1) || { echo -e "${RED}ERROR:${NC} Upload failed: $RESPONSE"; rm -f "$ZIP_FILE"; exit 1; }

rm -f "$ZIP_FILE"

DEPLOY_ID=$(echo "$RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['deploy_id'])" 2>/dev/null) \
  || { echo -e "${RED}ERROR:${NC} Could not parse deploy ID from: $RESPONSE"; exit 1; }

echo -e "   Deploy ID: ${BLUE}${DEPLOY_ID:0:8}${NC}"

# Poll status
echo -e "${CYAN}[3/4]${NC} Building and deploying..."
LAST_LOG_COUNT=0
DOTS=0

while true; do
  STATUS_JSON=$(curl -sf "$API/deploy/$DEPLOY_ID" 2>/dev/null) || { sleep 3; continue; }
  STATUS=$(echo "$STATUS_JSON" | python3 -c "import sys,json; print(json.load(sys.stdin)['status'])" 2>/dev/null)
  LOGS=$(echo "$STATUS_JSON" | python3 -c "import sys,json; [print(l) for l in json.load(sys.stdin).get('logs',[])]" 2>/dev/null)
  LOG_COUNT=$(echo "$LOGS" | wc -l)

  # Print new log lines
  if [ "$LOG_COUNT" -gt "$LAST_LOG_COUNT" ]; then
    echo "$LOGS" | tail -n $((LOG_COUNT - LAST_LOG_COUNT)) | while IFS= read -r line; do
      if [[ "$line" == ✓* ]]; then echo -e "   ${GREEN}$line${NC}"
      elif [[ "$line" == ✗* ]]; then echo -e "   ${RED}$line${NC}"
      elif [[ "$line" == ▶* ]]; then echo -e "   ${CYAN}$line${NC}"
      else echo "   $line"
      fi
    done
    LAST_LOG_COUNT=$LOG_COUNT
  fi

  case "$STATUS" in
    live)
      URL=$(echo "$STATUS_JSON" | python3 -c "import sys,json; print(json.load(sys.stdin).get('url',''))" 2>/dev/null)
      echo ""
      echo -e "${CYAN}[4/4]${NC} Done!"
      echo ""
      echo -e "  ${GREEN}✓ Deployment live!${NC}"
      echo -e "  URL: ${BLUE}${URL}${NC}"
      echo -e "  Dashboard: http://localhost:5173"
      echo ""
      exit 0
      ;;
    failed)
      echo ""
      echo -e "  ${RED}✗ Deployment failed.${NC}"
      echo "  Check the dashboard for full logs: http://localhost:5173"
      exit 1
      ;;
  esac
  sleep 3
done
