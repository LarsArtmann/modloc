#!/usr/bin/env bash
# Coverage floor: the statement coverage of all packages must reach the
# given percentage (default 80). BuildFlow runs the tests and reports the
# number; the floor is this project's own gate.
set -euo pipefail
cd "$(dirname "$0")/.."

floor="${1:-80}"
profile="$(mktemp)"
trap 'rm -f "$profile"' EXIT

go test -count=1 -coverprofile="$profile" ./... > /dev/null
total=$(go tool cover -func="$profile" | awk '/^total:/ { sub(/%/, "", $3); print $3 }')
echo "total coverage: ${total}% (floor ${floor}%)"

if ! awk -v got="$total" -v want="$floor" 'BEGIN { exit !(got + 0 >= want + 0) }'; then
	echo "FAIL: coverage ${total}% is below the floor of ${floor}%" >&2
	exit 1
fi
