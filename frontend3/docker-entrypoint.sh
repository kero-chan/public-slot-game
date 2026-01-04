#!/bin/sh

set -e

HTML_DIR=/usr/share/nginx/html

VITE_API_URL_VALUE="${VITE_API_URL:-/v1}"
VITE_GAME_ID_VALUE="${VITE_GAME_ID:-00000000-0000-0000-0000-000000000002}"
VITE_TEST_WIN_ANIMATIONS_VALUE="${VITE_TEST_WIN_ANIMATIONS:-false}"

echo "Replacing environment variables in built files..."
echo "  VITE_API_URL: ${VITE_API_URL_VALUE}"
echo "  VITE_GAME_ID: ${VITE_GAME_ID_VALUE}"
echo "  VITE_TEST_WIN_ANIMATIONS: ${VITE_TEST_WIN_ANIMATIONS_VALUE}"

find ${HTML_DIR}/assets -type f -name "*.js" ! -name "*.gz" | while read -r file; do
  sed -i "s|REPLACE_ENV_VITE_API_URL|${VITE_API_URL_VALUE}|g" "$file"
  sed -i "s|REPLACE_ENV_VITE_GAME_ID|${VITE_GAME_ID_VALUE}|g" "$file"
  sed -i "s|REPLACE_ENV_VITE_TEST_WIN_ANIMATIONS|${VITE_TEST_WIN_ANIMATIONS_VALUE}|g" "$file"
done

echo "Environment variable replacement complete"

exec nginx -g 'daemon off;'

