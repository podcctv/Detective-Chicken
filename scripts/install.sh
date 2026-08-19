#!/bin/sh
set -eu

SERVER_URL=""
ENROLL_TOKEN="${ENROLL_TOKEN:-}"

while [ "$#" -gt 0 ]; do
  case "$1" in
    --server) SERVER_URL="$2"; shift 2 ;;
    --enroll) ENROLL_TOKEN="$2"; shift 2 ;;
    *) echo "Unknown argument: $1" >&2; exit 2 ;;
  esac
done

[ -n "$SERVER_URL" ] || { echo "--server is required" >&2; exit 2; }
[ -n "$ENROLL_TOKEN" ] || { echo "--enroll or ENROLL_TOKEN is required" >&2; exit 2; }

installer="${SERVER_URL%/}/api/v1/install/${ENROLL_TOKEN}.sh"
if command -v curl >/dev/null 2>&1; then
  curl -fsSL --proto '=https' --tlsv1.2 "$installer" | sh
elif command -v wget >/dev/null 2>&1; then
  wget -qO- "$installer" | sh
else
  echo "curl or wget is required to download the generated installer" >&2
  exit 1
fi
