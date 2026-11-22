// An HTTP server and client talking over loopback.
package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("listen:", err)
		os.Exit(1)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "n=%s", r.URL.Query().Get("n"))
	})
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)

	client := &http.Client{Timeout: 10 * time.Second}
	url := "http://" + ln.Addr().String() + "/echo?n="
	const clients, each = 8, 25
	var wg sync.WaitGroup
	var mu sync.Mutex
	good := 0
	for c := 0; c < clients; c++ {
		wg.Add(1)
		go func(c int) {
			defer wg.Done()
			for i := 0; i < each; i++ {
				n := fmt.Sprint(c*1000 + i)
				resp, err := client.Get(url + n)
				if err != nil {
					fmt.Println("get:", err)
					return
				}
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				if string(body) == "n="+n {
					mu.Lock()
					good++
					mu.Unlock()
				}
			}
		}(c)
	}
	wg.Wait()
	srv.Close()
	fmt.Printf("%d of %d requests answered correctly\n", good, clients*each)
	if good != clients*each {
		fmt.Println("FAIL")
		os.Exit(1)
	}
	fmt.Println("ok")
}
