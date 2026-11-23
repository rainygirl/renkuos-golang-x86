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

`build.sh` applies them in this order; each builds on the ones before it.

| Patch | From |
| --- | --- |
| `go-haiku386.patch` (the port), `-netpoll`, `-preempt`, `-sockets`, `-semacreate`, `-osyield`, `-mmap` | renkuos-rtailscale `i386/go-runtime/`, which explains each |
| `go-haiku386-toolchain.patch` | this repository, revision 1 |
| `go-haiku386-syscallargs.patch` | renkuos-gotools, merged in revision 2 |
| `go-haiku386-hosts.patch` | renkuos-golang (arm64) `src/net/hook_haiku.go`, revision 3 |

This is the one haiku/386 patch series: the 32-bit rclone, Syncthing and gh
of [renkuos-gotools](https://github.com/rainygirl/renkuos-gotools) are built
with the cross toolchain that `build.sh` leaves in `.work/go`.

The first seven are the runtime port that
[renkuos-rtailscale](https://github.com/rainygirl/renkuos-rtailscale) builds
its 32-bit Tailscale with. `go-haiku386-toolchain.patch` is what the
toolchain itself needed on top:

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

`go-haiku386-syscallargs.patch` makes `syscall_sysvicall6` and
`syscall_rawsysvicall6` copy their six arguments into an eight-word array
whose last two words are zero. `go-haiku386-mmap.patch` had widened the
window `asmsysvicall6` copies to eight words, but these two still pointed it
at `a1` on their own frame, where words seven and eight are the result
slots. renkuos-gotools' `golang.org/x/sys/unix` for haiku/386 calls `mmap`
through `sysvicall6` with six words and needs the seventh, the high half of
the offset, to be zero.

`go-haiku386-hosts.patch` adds `src/net/hook_haiku.go`, the same file as in
renkuos-golang (arm64) revision 10. Haiku has no `/etc/hosts`; its hosts
file is `/boot/system/settings/network/hosts`, next to the `resolv.conf`
that korli's tree already reads. Without it the pure Go resolver sent
`localhost` to the DNS server and failed (`lookup localhost on 10.0.2.3:53:
no such host`), so `net.Listen("tcp", "localhost:0")` and Caddy's default
admin address did not work.

### What was merged from renkuos-gotools

renkuos-gotools built its 32-bit tools with the seven-patch tree of
renkuos-rtailscale and its own `patches/go/go-haiku386-fixes.patch` on top,
which did three things:

| Fix in go-haiku386-fixes.patch | Here |
| --- | --- |
| async preemption off by default on haiku/386, because `hello` crashed with SIGSEGV | dropped: a workaround for the unversioned `sigaction` import, fixed by `go-haiku386-toolchain.patch`. With it, async preemption stays on and `tests/preempt.go` (which hangs with `GODEBUG=asyncpreemptoff=1`) passes |
| two-word `off_t` in `Seek`, `pread`, `pwrite`, `Truncate`, `Ftruncate` | dropped: the same change is in `go-haiku386-toolchain.patch`, which also does `mmap` with `sysvicall8` |
| a zero-padded eight-word window in `syscall_sysvicall6` | kept, as `go-haiku386-syscallargs.patch` |

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

Revision 2, same machine and settings (2026-11-22):

- `tests/run.sh`, all passing, `preempt.go` included (5.7 s; with
  `GODEBUG=asyncpreemptoff=1` it still runs after 40 s, so it does test
  async preemption). Built natively and run repeatedly: `hello` 50 of 50,
  `preempt` 10 of 10, a stress of 2000 goroutines with channels, timers,
  tickers and a mutex 10 of 10.
- `os/exec`: 20 concurrent children with exit codes, and a stdin pipe.
- `net/http` client: `http://example.com/`, `http://neverssl.com/` and the
  QEMU host, all 200.
- Signals: `signal.Notify` for SIGINT and SIGTERM delivers both; without it,
  each terminates the program with that signal.
- `go test -short strconv sort os/signal time sync`: all pass except
  `os/signal`'s `TestNotifyContextStringer`, which expects the SIGTERM name
  `terminated` where Haiku's signal table (the same for amd64) says
  `termination requested`; the rest of `os/signal` passes with that test
  skipped.
- rclone, Syncthing and gh from renkuos-gotools, cross-built with
  `.work/go/bin/go` of this tree: see the verification there.

Revision 3, same machine and settings (2026-11-23), the package installed
with `pkgman install <file>`:

- `go version`: `go1.26.5 haiku/386`.
- `tests/run.sh`, all passing, including the new `localhost.go`:
  `net.LookupHost("localhost")` returns `[127.0.0.1 ::1]`, and an HTTP server
  listening on `localhost:0` answers a GET to `http://localhost:<port>/`.
- renkuos-gotools' eight 32-bit tools, cross-built with `.work/go/bin/go` of
  this tree: see the verification there.

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
