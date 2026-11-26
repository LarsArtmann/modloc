#!/usr/bin/env bash
# Truth gate: the spec count claimed in FEATURES.md must equal the number of
# Ginkgo specs in the sources, so the inventory cannot drift from the tests.
set -euo pipefail
cd "$(dirname "$0")/.."

actual=$(grep -rhoE '\b(It|Entry)\(' internal cmd --include='*_test.go' | wc -l | tr -d ' ')
claimed=$(grep -oE '[0-9]+ specs' FEATURES.md | head -n 1 | grep -oE '^[0-9]+' || true)

if [[ -z $claimed ]]; then
	echo "FAIL: FEATURES.md does not claim a spec count (expected the phrase 'N specs')" >&2
	exit 1
fi

if [[ $actual != "$claimed" ]]; then
	echo "FAIL: FEATURES.md claims $claimed specs but the sources define $actual" >&2
	echo "      run scripts/test-count.sh after adding specs and update FEATURES.md" >&2
	exit 1
fi

echo "ok: FEATURES.md spec count truthful ($actual specs)"
