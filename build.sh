#!/bin/sh
# build.sh - cross-build the Go toolchain for 32-bit RenkuOS/Haiku and
# package it as golang_x86.
#
#   ./build.sh
#
# Runs on Linux or macOS (not on Haiku: a 32-bit Haiku machine has no Go to
# bootstrap from). It
#   1. clones korli's Haiku Go tree at the pinned commit,
#   2. applies patches/ (the haiku/386 port and its runtime fixes),
#   3. runs make.bash with GOOS=haiku GOARCH=386, which leaves the native
#      toolchain in bin/haiku_386 and pkg/tool/haiku_386,
#   4. stages that as a package and, when a package tool is at hand, writes
#      dist/golang_x86-<version>-<revision>-x86_gcc2.hpkg.
#
# Environment:
#   GOROOT_BOOTSTRAP  a Go >= 1.24 for the build machine (default: "go" in PATH)
#   GO_SRC            use this already patched Go tree instead of cloning
#   PACKAGE           Haiku's package tool built for the build machine
#                     (default: "package" in PATH). Without one the staged
#                     tree is left in .work/stage: on Haiku,
#                     "cd .work/stage && package create ../../dist/<name>.hpkg".
#   REVISION          package revision (default 3)
#   VENDOR            vendor field; has to match the repository the package
#                     goes into (default "rainygirl")
#   WORK              scratch directory (default: ./.work)
set -eu

GO_REPOSITORY=${GO_REPOSITORY:-https://github.com/korli/go.git}
GO_BRANCH=golang-1.26-haiku
GO_COMMIT=4af74db8618f13d5bc0f3447fa6a7adb906fa709
REVISION=${REVISION:-3}
NAME=golang_x86
# Where HaikuPorts' golang puts GOROOT ($developLibDir/go), for the x86
# secondary architecture.
PREFIX=develop/lib/x86/go

die() { echo "error: $*" >&2; exit 1; }
note() { echo "==> $*"; }

here=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
work=${WORK:-$here/.work}
src=${GO_SRC:-$work/go}
stage=$work/stage
mkdir -p "$work" "$here/dist"

[ "$(uname -s)" != Haiku ] || die "cross-build this on Linux or macOS"

if [ -n "${GOROOT_BOOTSTRAP:-}" ]; then
	bootstrap=$GOROOT_BOOTSTRAP
else
	command -v go >/dev/null 2>&1 || die "no Go to bootstrap from: set GOROOT_BOOTSTRAP"
	bootstrap=$(go env GOROOT)
fi

# --- source ------------------------------------------------------------------

if [ -z "${GO_SRC:-}" ] && [ ! -f "$src/.patched" ]; then
	rm -rf "$src"
	note "cloning $GO_REPOSITORY at $GO_COMMIT"
	git clone --quiet --depth 1 --single-branch --branch "$GO_BRANCH" "$GO_REPOSITORY" "$src"
	git -C "$src" cat-file -e "$GO_COMMIT^{commit}" 2>/dev/null \
		|| git -C "$src" fetch --quiet --depth 1 origin "$GO_COMMIT"
	git -C "$src" checkout --quiet --detach "$GO_COMMIT"
	# The order matters: each patch builds on the ones before it.
	git -C "$src" apply --unidiff-zero "$here/patches/go-haiku386.patch"
	for p in netpoll preempt sockets semacreate osyield mmap toolchain syscallargs hosts; do
		git -C "$src" apply "$here/patches/go-haiku386-$p.patch"
	done
	: > "$src/.patched"
	note "applied patches/"
fi
[ -f "$src/src/runtime/syscall2_haiku.go" ] || die "$src is not a patched haiku/386 Go tree"
version=$(sed -n '1s/^go//p' "$src/VERSION")

# --- build -------------------------------------------------------------------

if [ ! -x "$src/bin/haiku_386/go" ]; then
	note "building Go $version for haiku/386 with $("$bootstrap/bin/go" version)"
	(cd "$src/src" && GOOS=haiku GOARCH=386 CGO_ENABLED=0 \
		GOROOT_BOOTSTRAP=$bootstrap ./make.bash)
fi

# --- stage -------------------------------------------------------------------

rm -rf "$stage"
root=$stage/$PREFIX
mkdir -p "$root/bin" "$root/pkg/tool" "$stage/bin" \
	"$stage/documentation/packages/$NAME"

install -m 0755 "$src/bin/haiku_386/go" "$src/bin/haiku_386/gofmt" "$root/bin/"
cp -R "$src/pkg/tool/haiku_386" "$root/pkg/tool/haiku_386"
# asm is always run with -I $GOROOT/pkg/include.
cp -R "$src/pkg/include" "$root/pkg/include"
# Go builds the standard library from source into the build cache, so the
# source is part of the toolchain; so is src/cmd, as the less used tools
# (doc, covdata, pprof, ...) are built from it on first use.
cp -R "$src/src" "$root/src"
cp -R "$src/lib" "$root/lib"
cp "$src/VERSION" "$root/VERSION"
# go.env as upstream, except that there is no haiku/386 toolchain to download.
sed 's/^GOTOOLCHAIN=auto$/GOTOOLCHAIN=local/' "$src/go.env" > "$root/go.env"
grep -q '^GOTOOLCHAIN=local$' "$root/go.env" || die "go.env: GOTOOLCHAIN line not found"

# go and gofmt find GOROOT from their own resolved path.
ln -s "../$PREFIX/bin/go" "$stage/bin/go"
ln -s "../$PREFIX/bin/gofmt" "$stage/bin/gofmt"

cp "$src/LICENSE" "$stage/documentation/packages/$NAME/LICENSE.go"
cp "$src/PATENTS" "$stage/documentation/packages/$NAME/PATENTS.go"
cp "$here/LICENSE" "$here"/README*.md "$stage/documentation/packages/$NAME/"

cat > "$stage/.PackageInfo" <<INFO
name		$NAME
version		$version-$REVISION
architecture	x86_gcc2
summary		"The Go compiler, standard library and tools for 32-bit x86"
description	"Go $version for 32-bit RenkuOS/Haiku: go, gofmt, the standard library and the compile, link, asm and vet tools, running natively. Built from korli's Haiku Go tree with an experimental haiku/386 port on top; upstream Go and HaikuPorts support Haiku on x86_64 only. No cgo. GOROOT is /boot/system/$PREFIX."
packager	"${PACKAGER:-rainygirl}"
vendor		"${VENDOR:-rainygirl}"
copyrights {
	"2009-2026 The Go Authors"
	"2026 rainygirl"
}
licenses {
	"BSD (3-clause)"
	"MIT"
}
urls {
	"https://github.com/korli/go"
}
provides {
	$NAME = $version
	cmd:go = $version
	cmd:gofmt = $version
}
requires {
	haiku
	haiku_x86
}
INFO

# --- package -----------------------------------------------------------------

package=${PACKAGE:-package}
out=$here/dist/$NAME-$version-$REVISION-x86_gcc2.hpkg
if command -v "$package" >/dev/null 2>&1; then
	rm -f "$out"
	(cd "$stage" && "$package" create -q "$out")
	note "$out"
else
	note "no package tool: the staged tree is in $stage"
fi
