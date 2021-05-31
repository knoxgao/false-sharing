package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Node present the data
type Node struct {
	Fields []int64
}

var (
	proc = flag.Int("p", 12, "number of threads,max > 0")
	n    = flag.Int("n", 100*1000, "number of access operation per thread max > 0")
	b    = flag.Bool("b", false, "batch test with threads from 1 to 12")
)

func main() {
	flag.Parse()
	if *proc <= 0 || *n <= 0 {
		panic(flag.ErrHelp)
	}

	noPadNode := &Node{
		Fields: make([]int64, *proc),
	}
	padNode := &Node{
		Fields: make([]int64, *proc*8),
	}

	runtime.GOMAXPROCS(*proc)

	runtimes := *n

	if !*b {
		doTest("test no pad", *proc, runtimes, func(i int) {
			atomic.AddInt64(&noPadNode.Fields[i], 1)
		})

		doTest("test pad", *proc, runtimes, func(i int) {
			atomic.AddInt64(&padNode.Fields[i*8], 1)
		})
	} else {
		for i := 1; i <= 12; i++ {
			runtime.GOMAXPROCS(i)
			doTest("test with no pad", i, runtimes, func(i int) {
				atomic.AddInt64(&noPadNode.Fields[i], 1)
			})

			doTest("test with pad", i, runtimes, func(i int) {
				atomic.AddInt64(&padNode.Fields[i*8], 1)
			})
		}
	}
}

func doTest(name string, procs, runTimes int, fn func(int)) {
	waitStart := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(procs)
	for i := 0; i < procs; i++ {
		go func(i int, n int) {
			<-waitStart
			for n > 0 {
				fn(i)
				n--
			}
			wg.Done()
		}(i, runTimes)
	}
	startTime := time.Now()
	close(waitStart)
	wg.Wait()
	dur := time.Since(startTime)
	fmt.Printf("[%s]:\tthreads %d, runtimes %d, cost %dus\n", name, procs, runTimes, dur.Microseconds())
}
