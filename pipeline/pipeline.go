package pipeline

import (
	"context"
	"fmt"
	"time"
)

func generate(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			// Simulate work
			case <-ctx.Done(): // Listen for cancellation
				return
			default:
				out <- n
			}
		}
	}()

	return out
}

func square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {

			select {
			case out <- n * n:
				// Simulate work
			case <-ctx.Done(): // Listen for cancellation
				return
			}
		}
	}()

	return out
}

// Stage 3: Print numbers
func printStage(ctx context.Context, in <-chan int) {
	for n := range in {
		select {
		case <-ctx.Done(): // Listen for cancellation
			fmt.Println("Print stage cancelled")
			return
		default:
			fmt.Printf("Result: %d\n", n)
			time.Sleep(500 * time.Millisecond) // Simulate work
		}
	}
}

// The Pipeline pattern is a concurrency pattern used to divide a
// complex task into stages that process data streaming and concurrently.
//
//	Each stage performs a specific transformation on every piece of data.
func PipeLine() {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Build the pipeline: generate -> square -> print
	numsChan := generate(ctx, 1, 2, 3, 4, 5)
	squares := square(ctx, numsChan)

	// Trigger cancellation after 1 second
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("Main: cancelling pipeline...")
		cancel() // Signal all stages to stop
	}()

	printStage(ctx, squares)
	fmt.Println("Pipeline stopped.")
}
