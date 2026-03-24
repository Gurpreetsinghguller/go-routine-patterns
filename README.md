# go-routine-patterns# Go Goroutine Patterns
> Concurrency & Channel Patterns Reference for Go Developers

---

## Table of Contents
1. [Fan-Out / Fan-In](#1-fan-out--fan-in)
2. [Worker Pool](#2-worker-pool)
3. [Pipeline](#3-pipeline)
4. [Timeout with Select](#4-timeout-with-select)
5. [Semaphore Pattern](#5-semaphore-pattern)
6. [Rate Limiter](#6-rate-limiter)
7. [Quick Reference](#7-quick-reference)
8. [Golden Rules](#8-golden-rules)

---

## 1. Fan-Out / Fan-In

Distribute work across N goroutines **(Fan-Out)**, then collect all results into a single channel **(Fan-In)**.

```
Input: [1, 2, 3, 4, 5, 6]
              |
        Split into N chunks
              |
    ┌─────────┼─────────┐
    ▼         ▼         ▼
 Worker1   Worker2   Worker3   ← Fan-OUT
 [1,2]     [3,4]     [5,6]
    │         │         │
    └─────────┼─────────┘
              ▼
    Single results channel     ← Fan-IN
              |
     [2, 4, 6, 8, 10, 12]
```

> **Rule:** Workers never close the shared channel. Use `sync.WaitGroup` to track all senders. The coordinator closes the channel only after all workers are done.


> **Pitfall:** Never call `wg.Wait()` before ranging the channel in the same goroutine — it causes a **deadlock**.

```go
// DEADLOCK
wg.Wait()               // blocks here forever
close(results)          // never reached
for res := range results { ... }  // never reached

// CORRECT — runs concurrently alongside the range
go func() {
    wg.Wait()
    close(results)
}()
for res := range results { ... }
```

> **Tip:** Results arrive in non-deterministic order. Tag results with worker ID if ordering matters.

---

## 2. Worker Pool

A fixed pool of N goroutines that continuously pull jobs from a shared jobs channel. More efficient than spawning a new goroutine per job.

```
 Jobs Channel (buffered)
 [job1, job2, job3, job4, job5 ...]
           |
    ┌──────┼──────┐
    ▼      ▼      ▼
  W1     W2     W3      ← Fixed pool (N=3)
(busy) (busy) (busy)
    │      │      │
    └──────┼──────┘
           ▼
   results channel
```

> **Rule:** Send all jobs then close the jobs channel. Workers `range` over it and exit cleanly when it closes.


> **Fan-Out vs Worker Pool:** Fan-Out pre-splits data. Worker Pool uses a shared jobs channel — workers pull dynamically, so faster workers naturally do more work.

---

## 3. Pipeline

Chain multiple stages together. Each stage reads from an input channel, transforms data, and writes to an output channel.

```
generate() ──► square() ──► print()
     │              │
  chan int        chan int

Each stage: func stage(ctx, in <-chan int) <-chan int
```

> **Rule:** Each stage owns its output channel and closes it when done. Use `context.WithCancel` for clean cancellation across all stages.

> 💡 **Tip:** Cancelling the context propagates through all stages immediately via the `select` statement.

---

## 4. Timeout with Select

Race between a result channel and a timer using `select`. Whichever fires first wins.

> **Rule:** Always pair a long-running goroutine with a timeout to prevent goroutine leaks.

```go
// ⚠️ Basic — has goroutine leak on timeout
func doWork() <-chan string {
    result := make(chan string, 1)
    go func() {
        time.Sleep(2 * time.Second) // simulate slow work
        result <- "done"
    }()
    return result
}

func main() {
    select {
    case res := <-doWork():
        fmt.Println("Got result:", res)
    case <-time.After(1 * time.Second):
        fmt.Println("Timed out!")
        // goroutine inside doWork() is still running!
    }
}
```

```go
// ✅ Better — use context.WithTimeout to cancel properly
func doWork(ctx context.Context) <-chan string {
    result := make(chan string, 1)
    go func() {
        select {
        case <-time.After(2 * time.Second):
            result <- "done"
        case <-ctx.Done():
            return // properly cancelled, no leak
        }
    }()
    return result
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    select {
    case res := <-doWork(ctx):
        fmt.Println("Got result:", res)
    case <-ctx.Done():
        fmt.Println("Timed out!")
    }
}
```

---

## 5. Semaphore Pattern

Limit the number of goroutines running concurrently using a buffered channel as a semaphore.

```
sem := make(chan struct{}, 3)  // max 3 concurrent

Goroutine 1: sem <- struct{}{}   ✅  (1/3 slots used)
Goroutine 2: sem <- struct{}{}   ✅  (2/3 slots used)
Goroutine 3: sem <- struct{}{}   ✅  (3/3 slots used)
Goroutine 4: sem <- struct{}{}   ⏳  BLOCKS until one finishes

When G1 finishes:  <-sem          🔓  G4 unblocks
```

> **Rule:** Acquire (send) before starting work, release (receive) when done. Buffer size = max concurrency.

```go
func main() {
    sem := make(chan struct{}, 3) // allow max 3 concurrent goroutines
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            sem <- struct{}{}        // acquire: blocks if 3 already running
            defer func() { <-sem }() // release on exit

            fmt.Printf("Worker %d running\n", id)
            time.Sleep(500 * time.Millisecond)
        }(i)
    }

    wg.Wait()
}
```

> 💡 **Tip:** `struct{}` uses zero bytes. Always prefer `chan struct{}` over `chan bool` or `chan int` for semaphores and signals.

---

## 6. Rate Limiter

Throttle operations to a fixed rate using `time.Ticker`.

```go
func main() {
    requests := []int{1, 2, 3, 4, 5}

    // Allow 2 requests per second
    limiter := time.NewTicker(500 * time.Millisecond)
    defer limiter.Stop()

    for _, req := range requests {
        <-limiter.C // wait for next tick
        fmt.Println("Processing request", req, "at", time.Now().Format("15:04:05.000"))
    }
}
```

For bursty rate limiting (allow N requests immediately, then throttle):

```go
// Burst of 3, then 1 per second
burstyLimiter := make(chan time.Time, 3)

// Pre-fill burst slots
for i := 0; i < 3; i++ {
    burstyLimiter <- time.Now()
}

// Refill at 1/sec
go func() {
    for t := range time.NewTicker(1 * time.Second).C {
        burstyLimiter <- t
    }
}()

for _, req := range requests {
    <-burstyLimiter
    fmt.Println("Processing request", req)
}
```

---

## 7. Quick Reference

| Pattern | Use When | Key Construct |
|---|---|---|
| **Fan-Out / Fan-In** | Parallelize chunked work, merge results | `sync.WaitGroup` + `chan` |
| **Worker Pool** | Fixed concurrency, dynamic job queue | `jobs chan` + `range` |
| **Pipeline** | Multi-stage data transformation | Chained channels |
| **Timeout / Select** | Limit wait time on slow operations | `select` + `time.After` |
| **Semaphore** | Cap max concurrent goroutines | `buffered chan struct{}` |
| **Rate Limiter** | Throttle request throughput | `time.Ticker` |

---

## 8. Golden Rules

1. **Channel Ownership** — The goroutine that creates a channel is responsible for closing it.

2. **Multiple Senders** — Never close from a sender when multiple senders exist. Use `WaitGroup` + a coordinator goroutine.

3. **Goroutine Leaks** — Every goroutine needs an exit path. Use `context.WithCancel` or a `done` channel.

4. **Buffered vs Unbuffered** — Unbuffered = synchronous handoff. Buffered = decouples sender and receiver speed.

5. **`select` Default** — Adding a `default` case makes `select` non-blocking. Omit it if you want to block and wait.

6. **Empty Struct** — Use `chan struct{}` for signals and semaphores — zero memory cost.