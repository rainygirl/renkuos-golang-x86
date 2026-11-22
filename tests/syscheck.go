// The parts of the runtime and syscall packages that the haiku/386 port had
// to get right by hand: signals, child processes, stat, 64-bit file offsets,
// mmap. Prints one line per check and exits non-zero if any failed.
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var failed bool

func check(name string, ok bool, detail ...any) {
	state := "ok  "
	if !ok {
		state = "FAIL"
		failed = true
	}
	fmt.Println(state, name, fmt.Sprint(detail...))
}

func nilDeref() (msg string) {
	defer func() { msg = fmt.Sprint(recover()) }()
	var p *int
	return fmt.Sprint(*p)
}

func divide(a, b int) (msg string) {
	defer func() { msg = fmt.Sprint(recover()) }()
	return fmt.Sprint(a / b)
}

func main() {
	dir, err := os.MkdirTemp("", "syscheck")
	if err != nil {
		fmt.Println("FAIL MkdirTemp", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	// Signals.
	check("nil dereference panics", nilDeref() == "runtime error: invalid memory address or nil pointer dereference", nilDeref())
	check("integer divide by zero panics", divide(1, 0) == "runtime error: integer divide by zero", divide(1, 0))
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	syscall.Kill(os.Getpid(), syscall.SIGUSR1)
	select {
	case s := <-c:
		check("signal.Notify", s == syscall.SIGUSR1, s)
	case <-time.After(3 * time.Second):
		check("signal.Notify", false, "timeout")
	}

	// Child processes.
	for i := 0; i < 5; i++ {
		out, err := exec.Command("/bin/sh", "-c", "echo child; exit 3").Output()
		ee, _ := err.(*exec.ExitError)
		if string(out) != "child\n" || ee == nil || ee.ExitCode() != 3 {
			check("exec, output and exit status", false, string(out), err)
			break
		}
		if i == 4 {
			check("exec, output and exit status", true)
		}
	}

	// stat.
	name := filepath.Join(dir, "f")
	data := bytes.Repeat([]byte("0123456789abcdef"), 4096) // 64 KiB
	os.WriteFile(name, data, 0o640)
	when := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	os.Chtimes(name, when, when)
	st, err := os.Stat(name)
	if err != nil {
		check("stat", false, err)
	} else {
		check("stat size", st.Size() == int64(len(data)), st.Size())
		check("stat mode", st.Mode() == 0o640, st.Mode())
		check("stat mtime", st.ModTime().Equal(when), st.ModTime().UTC())
	}
	f, _ := os.OpenFile(name, os.O_RDWR, 0)
	if fst, err := f.Stat(); err != nil {
		check("fstat", false, err)
	} else {
		check("fstat mtime and size", fst.ModTime().Equal(when) && fst.Size() == int64(len(data)), fst.ModTime().UTC(), " ", fst.Size())
	}
	lst, err := os.Lstat(dir)
	check("lstat directory", err == nil && lst.IsDir())
	ents, err := os.ReadDir(dir)
	check("readdir", err == nil && len(ents) == 1 && ents[0].Name() == "f")

	// 64-bit offsets.
	end, err := f.Seek(0, io.SeekEnd)
	check("seek to end", err == nil && end == int64(len(data)), end, err)
	back, err := f.Seek(-16, io.SeekCurrent)
	check("seek backwards", err == nil && back == int64(len(data))-16, back, err)
	buf := make([]byte, 16)
	n, err := f.ReadAt(buf, 4096+3)
	check("pread", err == nil && n == 16 && string(buf) == "3456789abcdef012", string(buf), err)
	_, err = f.WriteAt([]byte("XY"), 10)
	f.ReadAt(buf[:4], 9)
	check("pwrite", err == nil && string(buf[:4]) == "9XYc", string(buf[:4]), err)
	const big = 5 << 30 // past 4 GiB: needs both halves of the offset
	bigPos, err := f.Seek(big, io.SeekStart)
	check("seek past 4 GiB", err == nil && bigPos == big, bigPos, err)
	err = f.Truncate(1000)
	fst, _ := f.Stat()
	check("ftruncate", err == nil && fst.Size() == 1000, fst.Size(), err)
	err = os.Truncate(name, 100)
	st, _ = os.Stat(name)
	check("truncate", err == nil && st.Size() == 100, st.Size(), err)

	// mmap, at a file offset.
	os.WriteFile(name, data, 0o640)
	m, err := syscall.Mmap(int(f.Fd()), 8192, 4096, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		check("mmap", false, err)
	} else {
		check("mmap at an offset", bytes.Equal(m, data[8192:8192+4096]))
		syscall.Munmap(m)
	}
	f.Close()

	var rl syscall.Rlimit
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rl)
	check("getrlimit", err == nil && rl.Cur > 0 && rl.Cur <= rl.Max, rl.Cur, " ", rl.Max, " ", err)
	wd, err := os.Getwd()
	check("getwd", err == nil && wd != "", wd)
	host, err := os.Hostname()
	check("hostname", err == nil && host != "", host)

	if failed {
		os.Exit(1)
	}
	fmt.Println("all checks passed")
}
