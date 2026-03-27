package workerpool

import (
	"fmt"
	"sync"
)

func worker(id int, jobs <-chan int, results chan<- int) {
	fmt.Println("Worker", id, "started")
	for r := range jobs {
		fmt.Println("worker", id, "processing", r)
		results <- r * 2
	}
	fmt.Println("Worker", id, "finished")
}

func Manager() []int {
	const numWorkers = 3
	const numJobs = 10
	jobs := make(chan int, numJobs)
	wresults := make(chan int, numJobs)
	var finalResults []int
	wg := sync.WaitGroup{}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(id, jobs, wresults)
		}(i)
	}

	for j := 0; j < numJobs; j++ {
		jobs <- j
	}
	close(jobs)
	go func() {
		wg.Wait()
		close(wresults)
	}()
	for r := range wresults {
		finalResults = append(finalResults, r)
	}
	return finalResults
}

func Manager2() []int {
	numberOfWorkers := 3
	dataToBeProcessed := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	jobs := make(chan int, numberOfWorkers)
	results := make(chan int, numberOfWorkers)
	var finalResults []int
	wg := sync.WaitGroup{}
	controlerWG := sync.WaitGroup{}
	controlerWG.Add(1)
	go func() {
		defer controlerWG.Done()
		for r := range results {
			finalResults = append(finalResults, r)
		}
	}()

	for i := 0; i < numberOfWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(id, jobs, results)
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	for _, job := range dataToBeProcessed {
		jobs <- job
	}
	close(jobs)
	controlerWG.Wait()
	return finalResults
}

// This version is more concise and efficient, as it eliminates the need for an additional goroutine to close the results channel after all workers are done.
//
//	Instead, we can directly close the results channel in the main goroutine after waiting for all workers to finish.
//	This approach simplifies the code and reduces the overhead of managing an extra goroutine.
//
// Also we do not have to share the finalResults slice between goroutines, as we can safely append results in the main goroutine without any race conditions.
func Manager3() []int {
	const numWorkers = 3
	const numJobs = 10

	jobs := make(chan int, numWorkers)
	wresults := make(chan int, numWorkers)

	var wg sync.WaitGroup

	// start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(id, jobs, wresults)
		}(i)
	}

	// close results when all workers done
	go func() {
		wg.Wait()
		close(wresults)
	}()

	// send jobs
	go func() {
		for j := 0; j < numJobs; j++ {
			jobs <- j
		}
		close(jobs)
	}()

	// collect results in main goroutine — no sharing, no race
	var finalResults []int
	for r := range wresults {
		finalResults = append(finalResults, r)
	}

	return finalResults
}
