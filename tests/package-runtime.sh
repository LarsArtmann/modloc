#!/usr/bin/env bash
# Runtime check of the built package. The Nix build runs the specs; this one
# runs modloc the way a host does, through its wrapper and with an empty
# environment, and pins what the host receives: the scc on the wrapped PATH,
# the version stamps, the exit codes, the completion files, and that a
# directory the user cannot read is named and skipped, never fatal.
#
# Usage: package-runtime.sh <package> <version> <commit>
#   <package>  the store path of the built package
#   <version>  the version --version must print
#   <commit>   the commit --version must name; empty prints none
#
# A leg whose precondition this sandbox cannot provide (a directory its build
# user is not allowed to read) reports VACUOUS and does not run.
set -euo pipefail

package=$1
version=$2
commit=$3
modloc=$package/bin/modloc

fail() {
	echo "modloc package check failed: $*" >&2
	exit 1
}

# Any other failing command leaves the same marker, so a broken package can
# never fail this check without saying so.
trap 'echo "modloc package check failed: a command exited non-zero (see the lines above)" >&2' ERR

# empty runs modloc with no environment at all: whatever it needs must be its own.
empty() {
	env -i "$modloc" "$@"
}

# Counting works through the wrapper alone.
fixture=$PWD/repo
mkdir -p "$fixture/internal/x" "$fixture/internal/y/deep"
printf '%s\n' 'module example.com/root' >"$fixture/go.mod"
printf '%s\n' 'package main' >"$fixture/main.go"
printf '%s\n' 'package x' 'func X() {}' >"$fixture/internal/x/x.go"
printf '%s\n' 'module example.com/root/y' >"$fixture/internal/y/go.mod"
printf '%s\n' 'package y' >"$fixture/internal/y/y.go"
printf '%s\n' 'package d' >"$fixture/internal/y/deep/d.go"

empty "$fixture" >go.tsv
printf '%s\n' \
	'repo	module	files	code	comment	blank	lines' \
	"$fixture	.	2	3	0	0	3" \
	"$fixture	internal/y	2	2	0	0	2" >expected
diff -u expected go.tsv || fail "the TSV report differs from the expected one"

# doctor names an scc that runs, and it answers with the same version.
readonly DOCTOR_LINE='^scc  ok  (/nix/store/[^/]+/bin/scc) \(scc version (.+)\)$'
doctor=$(empty doctor)
if [[ ! $doctor =~ $DOCTOR_LINE ]]; then
	fail "doctor printed '$doctor', expected 'scc  ok  <path> (scc version <version>)'"
fi

scc=${BASH_REMATCH[1]}
scc_version=${BASH_REMATCH[2]}
[ "$(env -i "$scc" --version)" = "scc version $scc_version" ] ||
	fail "$scc does not report the version doctor printed ($scc_version)"

# The stamps are the ones the flake linked in. fang abbreviates a commit to
# seven characters.
expected="modloc version $version"
if [ -n "$commit" ]; then
	expected+=" (${commit:0:7})"
fi
actual=$(empty --version)
[ "$actual" = "$expected" ] || fail "--version printed '$actual', expected '$expected'"

# A usage error exits 2, names the problem on stderr and prints nothing on stdout.
code=0
empty --bogus >usage.out 2>usage.err || code=$?
[ "$code" = 2 ] || fail "an unknown flag exited $code, expected 2"
grep -F 'unknown flag: --bogus' usage.err >/dev/null || fail "the unknown flag is not named on stderr"
[ ! -s usage.out ] || fail "a usage error wrote to stdout"

# Help is on stdout and exits 0.
empty --help >help.out
grep -F 'USAGE' help.out >/dev/null || fail "--help prints no usage"

# The completion files are installed.
for completion in bash-completion/completions/modloc.bash fish/vendor_completions.d/modloc.fish zsh/site-functions/_modloc; do
	[ -f "$package/share/$completion" ] || fail "completion file missing: share/$completion"
done

# A directory the user cannot read is named on stderr and skipped; the rest
# of the repository is still counted.
locked=$PWD/locked
mkdir -p "$locked/secret"
printf '%s\n' 'module example.com/locked' >"$locked/go.mod"
printf '%s\n' 'package locked' >"$locked/open.go"
printf '%s\n' 'module example.com/locked/secret' >"$locked/secret/go.mod"
printf '%s\n' 'package secret' >"$locked/secret/hidden.go"
chmod 000 "$locked/secret"
trap 'chmod 755 "$locked/secret"' EXIT

if ls "$locked/secret" >/dev/null 2>&1; then
	echo "modloc package check: VACUOUS: this sandbox lets its build user read a mode-000 directory, so the unreadable-directory leg did not run" >&2
else
	code=0
	empty locked >locked.tsv 2>locked.err || code=$?
	[ "$code" = 0 ] || fail "an unreadable subdirectory made the scan exit $code, expected 0"

	printf '%s\n' \
		'repo	module	files	code	comment	blank	lines' \
		'locked	.	1	1	0	0	1' >expected
	diff -u expected locked.tsv || fail "the report over the readable remainder differs from the expected one"

	grep -F "modloc: skipped 1 unreadable directory under 'locked': secret" locked.err >/dev/null ||
		fail "the unreadable directory is not named on stderr"
fi

echo "modloc package check: passed"
