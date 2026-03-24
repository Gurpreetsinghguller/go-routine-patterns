package fanInFanOut

import (
	"sync"
)

func worker(dTP []int, results chan []int, wg *sync.WaitGroup) {
	defer wg.Done()
	result := make([]int, len(dTP))
	for i, v := range dTP {
		result[i] = v * 2
	}
	results <- result

}

// 1, 2, 3, 4, 5, 6, 7, 8, 9, 10
// 0, 1, 2, 3, 4, 5, 6, 7, 8, 9
func Manager() []int {
	dataToBeProcessed := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	numberOfWorkers := 3
	results := make(chan []int, numberOfWorkers)
	resultingData := []int{}
	var wg sync.WaitGroup
	chunkSize := len(dataToBeProcessed) / numberOfWorkers
	for i := 0; i < numberOfWorkers; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == numberOfWorkers-1 { // if it's the last worker, it should process all the remaining data
			end = len(dataToBeProcessed)
		}
		wg.Add(1)
		go worker(dataToBeProcessed[start:end], results, &wg)
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	for r := range results {
		resultingData = append(resultingData, r...)
	}

	return resultingData
}
