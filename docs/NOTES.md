# Notes on the haiku/386 toolchain

## What is in the package

- Source: github.com/korli/go, branch `golang-1.26-haiku`, commit `4af74db8`
  (Go 1.26.5), plus `patches/` in the order `build.sh` applies them.
- `GOROOT` is `/boot/system/develop/lib/x86/go`, where HaikuPorts' golang puts
  it (`$developLibDir/go`) for the x86 secondary architecture. `bin/go` and
  `bin/gofmt` are symlinks into it.
- `pkg/tool/haiku_386`: asm, cgo, compile, cover, fix, link, preprofile, vet.
  The whole of `src/` ships, including `src/cmd`, because Go builds the less
  used tools from source on first use.
- `go.env` is upstream's with `GOTOOLCHAIN=local`.

## Patches

The first seven are the runtime port that
[renkuos-rtailscale](https://github.com/rainygirl/renkuos-rtailscale) builds
its 32-bit Tailscale with (`i386/go-runtime/` there, which explains each).
`go-haiku386-toolchain.patch` is what the toolchain itself needed on top:

- **Versioned libroot imports.** The base port imported `sigaction`,
  `sigprocmask`, `sysconf`, `fstat`, `stat`, `lstat`, `getrlimit` and
  `getrusage` without a symbol version, and an unversioned import binds to the
  oldest, BeOS-compatible version. That `sigaction` reads `struct sigaction`
  with a shorter layout, so the signal mask the runtime passed was taken for
  the flags (every flag set, `SA_NOCLDWAIT` among them) and the handler was
  called with the registers inline and no `siginfo`. The imports now name the
  versions korli's amd64 port uses, and `sigtramp` takes
  `(sig, siginfo*, ucontext*)`.
- **64-bit file offsets in `syscall`.** `Seek`, `pread`, `pwrite`, `Truncate`,
  `Ftruncate` and `mmap` passed their `off_t` as one stack word. They are now
  written by hand with both halves, and `Seek` returns both halves. `mmap`
  needs seven words, for which there is a new `sysvicall8`.
- **`cmd/vendor/golang.org/x/sys/unix` for haiku/386**, which `cmd/go` does
  not build without: the amd64 files adapted to the 32-bit layout (checked
  against `sizeof`/`offsetof` from the x86 Haiku headers), with the same
  hand-written 64-bit functions.

Measured on the test machine below, with the first seven patches only and
default settings: `hello` crashed with SIGSEGV in 20 of 40 runs (0 of 40
with `GODEBUG=asyncpreemptoff=1`), and a program that recovered from a nil
dereference, received SIGUSR1 or started a child process crashed every time,
`go build` included. With the eighth patch: 0 of 50, and the checks below
pass.

## Verified

Machine: Haiku nightly hrev60450 `x86_gcc2h` in QEMU (TCG, 4 virtual CPUs, 4
GB), not RenkuOS and not real hardware. The package was installed with
`pkgman install <file>`; `GOCACHE`, `GOPATH` and `TMPDIR` pointed at a second
volume, as the boot volume of that image is too small for the build cache.
Runtime settings were the defaults: no `GOMAXPROCS`, no `GODEBUG`.

- `go version`, `go env`: `go1.26.5 haiku/386`, `GOROOT` found through the
  symlink.
- `tests/run.sh`, all passing. Build times, cold cache first:
  `hello.go` 34 s (includes compiling `runtime` and `fmt`), `syscheck.go` 4 s,
  `conc.go` 1 s, `httploop.go` 46 s (includes `net/http`), `dep` 3 s for
  `go mod tidy` (downloads github.com/google/uuid from proxy.golang.org) and
  5 s to build.
- `go run`, `go vet`, `gofmt -l`, `go test -v`, `go test -cover`,
  `go test -bench`, `go doc fmt.Println`, `go list std` (357 packages).
- `go build std`: exit 0.
- Programs cross-built with the same tree and run repeatedly on that machine:
  `hello` 50 of 50, `conc` 20 of 20, `httploop` 10 of 10, `syscheck` 10 of 10.

## Not verified, known limits

- RenkuOS i386 itself and real 32-bit hardware. A package built this way has
  not been run there.
- The 32-bit Tailscale in renkuos-rtailscale is built with the first seven
  patches only and has not been rebuilt with this tree.
- No cgo, and no external linking.
- Third-party modules that import `golang.org/x/sys/unix`: the published
  module has no Haiku files. The
  `golang_bin` package for arm64 works around this with a local module proxy;
  that has not been done here.
- `Termios` in the vendored `x/sys/unix` follows the R1/beta4 headers; nothing
  in the tests uses it. `syscall.Termios` for 386 has a different layout and
  was left as it was.
- `setsigstack` and the handling of signal handlers installed by non-Go code
  were not exercised.
- Go's own test suite (`run.bash`) was not run.
