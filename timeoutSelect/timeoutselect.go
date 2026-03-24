package timeoutselect

import (
	"context"
	"sync"
)

func worker(ctx context.Context, job <-chan int, result chan<- int) {
	for {
		select {
		case data, ok := <-job:
			if !ok {
				close(result)
				return
			}
			result <- data * 2
		case <-ctx.Done():
			close(result)
			return
		}
	}
}

func Manager() []int {
	dataToBeProcessed := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	processedData := []int{}
	job := make(chan int)
	result := make(chan int)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		worker(ctx, job, result)
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for r := range result {
			processedData = append(processedData, r)
		}
	}()

	for i := 0; i < len(dataToBeProcessed); i++ {
		job <- dataToBeProcessed[i]
	}
	close(job)
	wg.Wait()
	cancel()
	return processedData
}
