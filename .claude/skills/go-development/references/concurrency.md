# Concurrency

Go makes concurrency easy to start and easy to get subtly wrong. The recurring theme: every
goroutine needs a known exit, and shared state needs either a channel or a lock — never bare
access.

## Goroutines
- In libraries, prefer letting the caller control concurrency rather than spawning goroutines
  internally. If a library must spawn one, document its lifecycle and cleanup.
- Always define how each goroutine exits. A goroutine with no exit path is a leak.
- Coordinate completion with `sync.WaitGroup` or channels.
- Ensure cancellation and cleanup paths exist (e.g., honor `context.Context` cancellation) so
  goroutines don't outlive their purpose.

## Channels
- Use channels to communicate between goroutines; prefer communicating over channels to
  sharing memory and locking around it.
- Close a channel from the sender side, never the receiver side.
- Use buffered channels only when the capacity is known and intentional.
- Use `select` for non-blocking or multiplexed channel operations.

## Synchronization
- `sync.Mutex` protects shared mutable state — keep critical sections small.
- `sync.RWMutex` when reads dominate and contention matters.
- Choose by purpose: channels for communication, mutexes for protecting state.
- `sync.Once` for one-time initialization.

### WaitGroup pattern by Go version
Pick the form that matches the `go` directive in `go.mod`:

- `go >= 1.25`: use `WaitGroup.Go`, which spawns and tracks in one call:
  ```go
  var wg sync.WaitGroup
  wg.Go(task1)
  wg.Go(task2)
  wg.Wait()
  ```
- `go < 1.25`: use the classic `Add`/`Done` pattern:
  ```go
  var wg sync.WaitGroup
  wg.Add(2)
  go func() { defer wg.Done(); task1() }()
  go func() { defer wg.Done(); task2() }()
  wg.Wait()
  ```
