// Goroutines, channels, timers, a mutex and the garbage collector.
package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

func main() {
	const workers, jobs = 16, 20000
	in, out := make(chan int), make(chan int, 64)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := range in {
				buf := make([]byte, 512) // garbage, so the collector runs
				out <- n*2 + len(buf) - 512
			}
		}()
	}
	go func() {
		for i := 1; i <= jobs; i++ {
			in <- i
		}
		close(in)
		wg.Wait()
		close(out)
	}()
	sum := 0
	for v := range out {
		sum += v
	}

	start := time.Now()
	ticks := 0
	tick := time.NewTicker(20 * time.Millisecond)
	done := time.After(500 * time.Millisecond)
loop:
	for {
		select {
		case <-tick.C:
			ticks++
		case <-done:
			break loop
		}
	}
	tick.Stop()
	runtime.GC()
	elapsed := time.Since(start)

	want := jobs * (jobs + 1)
	fmt.Printf("sum=%d want=%d ticks=%d elapsed=%s\n", sum, want, ticks, elapsed.Round(time.Millisecond))
	if sum != want || ticks < 15 || elapsed > 2*time.Second {
		fmt.Println("FAIL")
		os.Exit(1)
	}
	fmt.Println("ok")
}
