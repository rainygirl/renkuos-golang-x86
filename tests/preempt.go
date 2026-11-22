// Asynchronous preemption: with one P, goroutines spinning in loops that make
// no function calls can only be interrupted by a signal. If async preemption
// does not work, the sleep, the timer and the GC below never return and the
// program hangs; if the signal path is broken, it crashes.
package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

var sink [4]uint64

func spin(i int) {
	for {
		sink[i]++
	}
}

func main() {
	runtime.GOMAXPROCS(1)
	for i := range sink {
		go spin(i)
	}
	start := time.Now()
	for round := 0; round < 20; round++ {
		time.Sleep(5 * time.Millisecond)
		<-time.After(time.Millisecond)
		runtime.GC()
	}
	if sink[0] == 0 {
		fmt.Println("FAIL preempt: the spinning goroutines never ran")
		os.Exit(1)
	}
	fmt.Println("ok   async preemption, 20 rounds of sleep+timer+GC in", time.Since(start).Round(time.Millisecond))
}
