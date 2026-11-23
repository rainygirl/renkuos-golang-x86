#!/bin/sh
# run.sh - smoke test for an installed golang_x86, run on the 32-bit machine.
#
#   sh tests/run.sh            # everything
#   sh tests/run.sh offline    # skip the test that downloads a module
#
# Builds and runs each program in this directory with the installed go and
# prints how long each build took. Exits non-zero at the first failure.
set -eu

here=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
work=${TMPDIR:-/tmp}/golang-x86-test.$$
mkdir -p "$work"
trap 'rm -rf "$work"' EXIT

timed() {
	label=$1
	shift
	start=$(date +%s)
	"$@"
	echo "    $label: $(( $(date +%s) - start )) s"
}

go version
for prog in hello syscheck conc preempt httploop localhost; do
	echo "== $prog"
	timed "go build" go build -o "$work/$prog" "$here/$prog.go"
	"$work/$prog"
done

if [ "${1:-}" != offline ]; then
	echo "== dep (downloads github.com/google/uuid)"
	cp -R "$here/dep" "$work/dep"
	(
		cd "$work/dep"
		timed "go mod tidy" go mod tidy
		timed "go build" go build -o "$work/depbin" .
	)
	"$work/depbin"
fi
echo "all passed"
