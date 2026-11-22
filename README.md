# Go for 32-bit RenkuOS

English · [Italiano](README.it.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

The Go toolchain for 32-bit x86 RenkuOS (Haiku, `x86_gcc2`), running on the
machine itself: `go`, `gofmt`, the standard library and the compiler, linker,
assembler and vet. Go 1.26.5, from [korli's Haiku Go tree](https://github.com/korli/go)
with an experimental `haiku/386` port on top. Upstream Go and HaikuPorts
support Haiku on x86_64 only; this is unofficial.

## Install with pkgman

```sh
pkgman add-repo https://pkgman.rainygirl.com/x86_gcc2
pkgman install golang_x86
```

If `pkgman add-repo` fails with `Operation not supported`, the network kit of
that image has no TLS; use `http://` instead of `https://` in the address.

The package installs `go` and `gofmt`; `GOROOT` is
`/boot/system/develop/lib/x86/go`. It takes about 60 MB in
`/boot/system/packages`.

## Install from a file

```sh
pkgman install ./golang_x86-1.26.5-2-x86_gcc2.hpkg
```

## Use

```sh
go version                    # go version go1.26.5 haiku/386
go build ./...
```

- No cgo: `CGO_ENABLED` is 0 and cannot be turned on.
- `GOTOOLCHAIN` is `local`. There is no `haiku/386` toolchain to download, so
  a `go.mod` that asks for a newer Go fails instead of fetching one.
- Modules that import `golang.org/x/sys/unix` do not build: the published
  module has no Haiku files for any architecture.
- The first build compiles the standard library into the build cache
  (`~/.cache/go-build`). Point `GOCACHE`, `GOPATH` and `TMPDIR` at another
  volume if the boot volume is small.

## Uninstall

```sh
pkgman uninstall golang_x86
```

## Build the package

On Linux or macOS, with a Go to bootstrap from and Haiku's `package` tool:

```sh
./build.sh                    # writes dist/golang_x86-<version>-2-x86_gcc2.hpkg
```

`tests/run.sh` is the smoke test, to be run on the 32-bit machine after
installing. What the patches change, what was tested and what was not is in
[docs/NOTES.md](docs/NOTES.md). `patches/` is the one haiku/386 patch series: the
32-bit tools of [renkuos-gotools](https://github.com/rainygirl/renkuos-gotools)
are built with this toolchain too.

## AI disclosure

Parts of this port were developed with the assistance of AI coding tools
(Anthropic's Claude). All code has been reviewed and tested by the author.
