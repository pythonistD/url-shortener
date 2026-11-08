#!/usr/bin/env bash
set -euo pipefail

if [ -z "${BASE_URL:-}" ] && [ -n "${SERVER_PORT:-}" ]; then
  BASE_URL="http://localhost:${SERVER_PORT}"
fi
BASE_URL="${1:-${BASE_URL:-http://localhost}}"

echo "Running smoke tests against ${BASE_URL}..."

# 1. Shorten valid URL
RES=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/url" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com/golang/go"}')
STATUS=$(echo "${RES}" | tail -n 1)
BODY=$(echo "${RES}" | sed '$d')
test "${STATUS}" -eq 201
SHORT=$(echo "${BODY}" | grep -o '"short_url":"[^"]*' | cut -d'"' -f4)
echo "PASS: Shorten URL (status: ${STATUS}, short: ${SHORT})"

# 2. Idempotency test (same URL returns same short code)
RES=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/url" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com/golang/go"}')
STATUS=$(echo "${RES}" | tail -n 1)
BODY=$(echo "${RES}" | sed '$d')
DUP_SHORT=$(echo "${BODY}" | grep -o '"short_url":"[^"]*' | cut -d'"' -f4)
test "${STATUS}" -eq 201
test "${SHORT}" = "${DUP_SHORT}"
echo "PASS: Idempotency (status: ${STATUS}, matched: ${DUP_SHORT})"

# 3. Redirect to original URL
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/${SHORT}")
test "${STATUS}" -eq 302
echo "PASS: Redirect (status: ${STATUS})"

# 4. Non-existent short code
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/nonexistent")
test "${STATUS}" -eq 404
echo "PASS: Not found (status: ${STATUS})"

# 5. Invalid URL format
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/url" \
  -H "Content-Type: application/json" \
  -d '{"url":"invalid-url"}')
test "${STATUS}" -eq 400
echo "PASS: Invalid URL validation (status: ${STATUS})"

# 6. Invalid TTL (ttl=0)
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/url" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","ttl":0}')
test "${STATUS}" -eq 400
echo "PASS: Invalid TTL validation (status: ${STATUS})"

echo "All smoke tests passed successfully!"

