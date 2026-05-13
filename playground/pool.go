package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, jobs chan int, results chan int, wg *sync.WaitGroup) {
	defer wg.Done() // called when this goroutine returns, no matter what
	for job := range jobs {
		// range over channel — blocks until a job arrives, exits when channel is closed
		fmt.Printf("worker %d picked up job %d\n", id, job)
		time.Sleep(500 * time.Millisecond) // simulate work
		results <- job * 2
	}
}

func main() {
	jobs := make(chan int, 10)    // buffered channel — can hold 10 jobs
	results := make(chan int, 10) // buffered so workers don't block writing results
	var wg sync.WaitGroup
	// spawn 3 workers
	for i := 1; i <= 3; i++ {
		wg.Add(1) // one more goroutine starting
		go worker(i, jobs, results, &wg)
	}
	//Notice &wg — we pass a pointer to the WaitGroup, not a copy.
	//All 3 workers share the same WaitGroup counter.
	//If you passed wg without &, each worker would get its own copy and Done()
	// would decrement the wrong counter.

	// send 10 jobs
	for j := 1; j <= 10; j++ {
		jobs <- j
	}
	close(jobs) // tell workers no more jobs are coming

	// close results channel once all workers are done
	// this must be in a separate goroutine — wg.Wait() blocks,
	// and we need to range over results below on the main goroutine
	go func() { // define a function with no name
		wg.Wait()
		close(results) // safe to close now — no more workers writing to it
	}() // function definition ends here  // () = immediately call it

	//time.Sleep(3 * time.Second) // dirty wait — we'll fix this with WaitGroup
	//wg.Wait() // block here until all 3 workers call wg.Done()
	// collect all results
	for result := range results {
		fmt.Println("result:", result)
	}
	fmt.Println("all jobs done")
}

//wg.Wait()           → blocks until all workers finish
//range results       → blocks until results channel is closed
//If main calls wg.Wait() first, it freezes. But workers are trying to push into results — and nobody is reading from results. Since results is buffered with capacity 10, once it fills up, workers block trying to write. Now everyone is stuck:
//main     → blocked on wg.Wait()
// worker 1 → blocked on results <- (channel full)
// worker 2 → blocked on results <- (channel full)
// worker 3 → blocked on results <- (channel full)
//This is a deadlock — everyone waiting for everyone else, program hangs forever.
