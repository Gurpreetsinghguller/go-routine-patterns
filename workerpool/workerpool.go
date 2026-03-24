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
