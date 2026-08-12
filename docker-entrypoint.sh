#!/bin/sh

set -eu

config_file="${FRONTEND_DIST:-/app/frontend/dist}/config.js"

{
  printf 'window.configs = '
  jq -n \
    --arg posthog_api_key "${VUE_APP_POSTHOG_API_KEY:-}" \
    --arg google_client_id "${VUE_APP_GOOGLE_CLIENT_ID:-${CLIENT_ID:-}}" \
    --arg microsoft_client_id "${VUE_APP_MICROSOFT_CLIENT_ID:-${MICROSOFT_CLIENT_ID:-}}" \
    '{
      VUE_APP_POSTHOG_API_KEY: $posthog_api_key,
      VUE_APP_GOOGLE_CLIENT_ID: $google_client_id,
      VUE_APP_MICROSOFT_CLIENT_ID: $microsoft_client_id
    }'
  printf ';\n'
} > "$config_file"

exec "$@"
