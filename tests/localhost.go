// "localhost" resolves through Haiku's hosts file, and an HTTP server
// listening on localhost answers a client that dials it by name.
package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"slices"
	"time"
)

func fail(f string, a ...any) {
	fmt.Printf(f+"\nFAIL\n", a...)
	os.Exit(1)
}

func main() {
	addrs, err := net.LookupHost("localhost")
	if err != nil {
		fail("LookupHost(localhost): %v", err)
	}
	fmt.Println("localhost:", addrs)
	if !slices.Contains(addrs, "127.0.0.1") {
		fail("127.0.0.1 missing")
	}

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		fail("listen: %v", err)
	}
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "hello")
	})}
	go srv.Serve(ln)
	defer srv.Close()

	_, port, _ := net.SplitHostPort(ln.Addr().String())
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("http://localhost:" + port + "/")
	if err != nil {
		fail("get: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(body) != "hello" {
		fail("body %q", body)
	}
	fmt.Println("ok")
}
