#!/usr/bin/env bash
# tests/negative/run.sh: the bite-proof of the specs and the flake checks.
#
# A check that cannot fail is decoration. Every case copies the git-tracked
# tree, applies ONE contract break, runs the suite that should catch it and
# requires the failure to carry the case's marker (the spec it expects to
# see failing). An unrelated failure does not count, and neither does a
# tree that no longer compiles: the check did not bite. The unmutated copy
# is the green control: everything passes there before anything is broken.
#
# Usage:
#   bash tests/negative/run.sh                 # every fast case
#   CASES=cli,render bash tests/negative/run.sh # the cases of these groups only
#   SLOW=1 bash tests/negative/run.sh          # also the cases that build the package
#   KEEP=1 bash tests/negative/run.sh          # keep the copies under ~/.cache for triage
#   bash tests/negative/run.sh --help
#
# Needs go and scc on PATH (nix develop --command bash tests/negative/run.sh
# inside the dev shell, plus the scc of the host).
#
# Exit codes: 0 every case bit, 1 a case did not bite or the control is red,
# 2 the environment cannot run the harness.
set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SYSTEM="$(uname -m)-$(uname -s | tr '[:upper:]' '[:lower:]')"
WORK="$(mktemp -d "${XDG_CACHE_HOME:-$HOME/.cache}/modloc-negative.XXXXXX")"
[ "${KEEP:-0}" = "1" ] || trap 'rm -rf "$WORK"' EXIT

FILTER="${CASES:-}"

if [ "${1:-}" = "--help" ] || [ "${1:-}" = "-h" ]; then
	sed -n 's/^# \?//p' "$0" | sed -n '2,24p'
	exit 0
fi

for tool in go scc git python3 rsync; do
	command -v "$tool" >/dev/null 2>&1 || {
		echo "cannot run the harness: '$tool' is not on PATH" >&2
		exit 2
	}
done

# Every flake check that needs no package build: the green control runs them
# all, and a nix case names the one it breaks.
NIX_CHECKS="nixos-module home-manager-module stack-drift secrets"

BIT=0
MISSED=0

# make_copy copies the tracked tree, with its working-tree contents, into a
# fresh directory under $WORK.
make_copy() {
	local dir="$WORK/$1"
	mkdir -p "$dir"
	(cd "$REPO_ROOT" && git ls-files -z | rsync -a --files-from=- --from0 . "$dir/")
	printf '%s' "$dir"
}

# mutate replaces old with new in file, exactly once, inside a copy. It never
# touches the repository itself. Any other count is a harness problem (the
# anchor went stale), never a verdict about the specs.
mutate() { # mutate <file> <old> <new>
	case $PWD/ in
	"$WORK"/*) ;;
	*)
		echo "MUTATION-ANCHOR: mutate ran outside the copies ($PWD)" >&2
		exit 2
		;;
	esac

	python3 - "$1" "$2" "$3" <<'EOF'
import sys

path, old, new = sys.argv[1:4]
text = open(path, encoding="utf-8").read()
found = text.count(old)
if found != 1:
    sys.exit(f"MUTATION-ANCHOR: {path}: expected exactly one occurrence of {old!r}, found {found}")
open(path, "w", encoding="utf-8").write(text.replace(old, new))
EOF
}

# as_flake makes a copy a git repository: the flake's source filter only
# passes tracked files.
as_flake() {
	git init -q
	git add -A
}

# breaks describes a case: the group it belongs to, the suite that must
# catch it (go: a package; nix-eval: a check that fails the evaluation;
# nix-build: a check whose build must fail), the marker the failure must
# carry, and the mutations (file, old, new) that break the contract.
breaks() {
	GROUP=$1
	KIND=$2
	TARGET=$3
	MARKER=$4
	shift 4
	MUTATIONS=("$@")
}

# ---------------------------------------------------------------------------
# The cases. A function named case_<name> that ends in breaks.
# ---------------------------------------------------------------------------

case_root_without_panic_recovery() {
	breaks cli go internal/cli \
		'reports a crash in the counter as a failure, not as a panic' \
		internal/cli/cli.go \
		'recoverPanics := v5.RecoveryMiddleware[Config]()' \
		$'recoverPanics := func(_ context.Context, _ *Config, _ v5.CommandMetadata, next func() error) error {\n\t\t\treturn next()\n\t\t}'
}

case_allowed_values_tag() {
	breaks cli go internal/cli \
		'pins the whole contract: every flag, its default and values, and every command' \
		internal/cli/cli.go \
		'values:"go,rust"' \
		'values:"go,rust,python"'
}

case_paths_checked_one_by_one() {
	breaks cli go internal/cli \
		'checks every path before it scans any of them' \
		internal/cli/scan.go \
		$'\terr = validateRepositories(repos)\n\tif err != nil {\n\t\treturn err\n\t}\n' \
		'' \
		$'\tfor _, repo := range repos {\n\t\treport, scanErr :=' \
		$'\tfor _, repo := range repos {\n\t\tif err := validateRepositories([]string{repo}); err != nil {\n\t\t\treturn err\n\t\t}\n\n\t\treport, scanErr :='
}

case_column_width_ignores_the_header() {
	breaks render go internal/render \
		'writes a table with grouped, right-aligned numbers that ends in a TOTAL row' \
		internal/render/render.go \
		'widest := width' \
		'widest := width * 0'
}

case_nested_modules_count_toward_their_parent() {
	breaks loc go internal/loc \
		"never counts a nested module's lines toward the module that encloses it" \
		internal/loc/attribute.go \
		$'\t\t\treturn dir\n' \
		$'\t\t\treturn RootModule\n'
}

case_node_modules_becomes_a_module() {
	breaks loc go internal/loc \
		'does not descend into node_modules' \
		internal/loc/markers.go \
		'case "vendor", "target", "node_modules", ".git":' \
		'case "vendor", "target", ".git":'
}

case_scc_code_read_from_comments() {
	breaks scc go internal/scc \
		'returns every file scc reports, relative to the repository' \
		internal/scc/scc.go \
		'Code:    job.Code,' \
		'Code:    job.Comment,'
}

case_nixos_module_ignores_enable() {
	breaks nix nix-build nixos-module \
		'modloc module check failed: nixos-module' \
		deploy/nixos/modloc.nix \
		'config = lib.mkIf cfg.enable {' \
		'config = {'
}

case_home_manager_module_ignores_enable() {
	breaks nix nix-build home-manager-module \
		'modloc module check failed: home-manager-module' \
		deploy/home-manager/modloc.nix \
		'config = lib.mkIf cfg.enable {' \
		'config = {'
}

case_nixos_module_installs_nothing() {
	breaks nix nix-build nixos-module \
		'modloc module check failed: nixos-module' \
		deploy/nixos/modloc.nix \
		'environment.systemPackages = [ cfg.package ];' \
		'environment.systemPackages = [ ];'
}

case_first_party_dependency_without_a_stack_row() {
	breaks nix nix-eval stack-drift \
		'modloc stack drift: required by go.mod but named nowhere in the Chosen foundation section: go-ndjson' \
		go.mod \
		'github.com/larsartmann/go-output v0.38.7' \
		$'github.com/larsartmann/go-output v0.38.7\n\tgithub.com/larsartmann/go-ndjson v0.0.1'
}

case_a_secret_in_the_tree() {
	breaks nix nix-build secrets \
		'modloc secrets check failed' \
		README.md \
		'# modloc' \
		$'# modloc\n\ngithub_token = ghp_''''_zK9mQ2xR7tW4bN8cV1dK6fH3jL0pS5yGuHi6'
}

# The slow cases build the package from the mutated tree.
case_package_without_the_scc_wrapper() {
	# The wrapper line, as it stands in the Nix string.
	# shellcheck disable=SC2016
	local wrapper='wrapProgram $out/bin/modloc --prefix PATH : ${lib.makeBinPath [ pkgs.scc ]}'
	breaks slow nix-build package-runtime \
		'modloc package check failed' \
		flake.nix \
		"$wrapper" \
		'true'
}

case_package_without_its_version_stamp() {
	breaks slow nix-build package-runtime \
		'modloc package check failed' \
		flake.nix \
		$'\t\t\t"-X main.version=${version}"\n' \
		''
}

# ---------------------------------------------------------------------------
# The run.
# ---------------------------------------------------------------------------

# apply breaks the contract in the copy, or reports a stale anchor.
apply() { # apply <dir>
	local index
	(
		cd "$1" || exit 2
		for ((index = 0; index < ${#MUTATIONS[@]}; index += 3)); do
			mutate "${MUTATIONS[index]}" "${MUTATIONS[index + 1]}" "${MUTATIONS[index + 2]}"
		done
	)
}

# judged runs one command in the copy and classifies its outcome.
judged() { # judged <dir> <label> <log> <cmd...>
	local dir=$1 label=$2 log=$3
	shift 3
	local code=0
	(cd "$dir" && "$@") >"$log" 2>&1 || code=$?

	if [ "$code" = 0 ]; then
		verdict NO_BITE
	elif grep -q 'MUTATION-ANCHOR' "$log"; then
		verdict ANCHOR
	elif grep -q '\[build failed\]\|\[setup failed\]' "$log"; then
		verdict COMPILE
	elif grep -qF "$MARKER" "$log"; then
		verdict BIT
	else
		verdict OTHER_FAILURE
	fi
}

verdict() {
	local outcome=$1
	case $outcome in
	BIT) BIT=$((BIT + 1)) ;;
	*) MISSED=$((MISSED + 1)) ;;
	esac
	printf '%-14s %s/%s\n' "$outcome" "${GROUP:-control}" "${NAME:-$label}"
	if [ "$outcome" != BIT ]; then
		printf '              log: %s\n' "$log"
	fi
}

# The control: nothing is broken, everything passes.
control=$(make_copy control)
flake_checks=()
for check in $NIX_CHECKS; do
	flake_checks+=(".#checks.$SYSTEM.$check")
done
log="$WORK/control.log"
code=0
(
	cd "$control" || exit 2
	as_flake || exit 1
	go test -count=1 ./... || exit 1
	nix build --no-link -L "${flake_checks[@]}"
) >"$log" 2>&1 || code=$?
if [ "$code" = 0 ]; then
	printf '%-14s %s\n' PASS control
else
	printf '%-14s %s\n' RED control
	printf '              the tree fails before anything is mutated; fix that first\n'
	printf '              log: %s\n' "$log"
	exit 1
fi

for case_function in $(compgen -A function | grep '^case_' | sort); do
	NAME=${case_function#case_}
	"$case_function"

	if [ -n "$FILTER" ] && [[ ",$FILTER," != *",$GROUP,"* ]]; then
		continue
	fi
	if [ "$GROUP" = slow ] && [ "${SLOW:-0}" != 1 ]; then
		continue
	fi

	dir=$(make_copy "$GROUP-$NAME")
	log="$WORK/$GROUP-$NAME.log"
	apply "$dir" 2>"$log" || {
		verdict ANCHOR
		continue
	}

	case $KIND in
	go) judged "$dir" go "$log" go test -count=1 "./$TARGET/" ;;
	nix-eval)
		(cd "$dir" && as_flake)
		judged "$dir" eval "$log" nix eval --raw ".#checks.$SYSTEM.$TARGET.drvPath"
		;;
	nix-build)
		(cd "$dir" && as_flake)
		judged "$dir" build "$log" nix build --no-link -L ".#checks.$SYSTEM.$TARGET"
		;;
	esac
done

echo
if [ "$MISSED" = 0 ]; then
	echo "all $BIT mutations bit"
else
	echo "$BIT mutations bit, $MISSED did not (NO_BITE: the suite passes with the contract broken; ANCHOR: the harness went stale; COMPILE: the mutation broke the build; OTHER_FAILURE: it failed for another reason)"
	[ "${KEEP:-0}" = 1 ] || echo "run again with KEEP=1 to keep the logs"
fi
[ "$MISSED" = 0 ]
