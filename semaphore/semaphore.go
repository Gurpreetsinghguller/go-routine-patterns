package semaphore

import "sync"

type Semaphore struct {
	ch       chan struct{}
	maxLimit int
}

func NewSemaphore(limit int) *Semaphore {
	return &Semaphore{
		ch:       make(chan struct{}, limit),
		maxLimit: limit,
	}
}

func (s *Semaphore) Acquire() {
	s.ch <- struct{}{} // This waits if the channel is full, effectively blocking until a slot is available
}

func (s *Semaphore) Available() bool {
	return len(s.ch) < s.maxLimit
}

func (s *Semaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *Semaphore) Release() {
	<-s.ch // This releases a slot in the semaphore, allowing another goroutine to acquire it
}

func worker(data int, result chan<- int) {

	result <- data * 2

}

func Manager() []int {
	dataToBeProcessed := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	processedData := []int{}

	result := make(chan int)
	sem := NewSemaphore(3)
	var wg sync.WaitGroup
	var collecterWG sync.WaitGroup
	collecterWG.Add(1)
	go func() {
		defer collecterWG.Done()
		for r := range result {
			processedData = append(processedData, r)
		}
	}()

	for i := 0; i < len(dataToBeProcessed); i++ {
		sem.Acquire() // Acquire a slot in the semaphore, blocking if necessary
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			defer sem.Release()
			worker(val, result)
		}(dataToBeProcessed[i])
	}

	wg.Wait()
	close(result)
	collecterWG.Wait()

	return processedData
}
