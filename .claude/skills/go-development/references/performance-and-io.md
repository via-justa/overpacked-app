# Performance and I/O

Profile before you optimize. Most Go is fast enough; spend effort where measurement says it
matters, and prefer algorithmic wins over micro-optimizations.

## Memory management
- Minimize allocations in hot paths.
- Reuse objects where it pays off (`sync.Pool` for short-lived, frequently-allocated objects).
- Use value receivers for small structs.
- Preallocate slices when the size is known (`make([]T, 0, n)`).
- Avoid unnecessary `string`/`[]byte` conversions.

## Profiling
- Use `pprof`; benchmark critical paths with `testing.B`.
- Profile first, optimize second, and focus on algorithmic improvements before fiddling.

## Readers and buffers
Most `io.Reader` streams are consume-once — reading advances internal state, so you can't
re-read without doing something deliberate.

- To read data more than once, buffer it once: `io.ReadAll` (or a limited read) into `[]byte`,
  then make fresh readers per use via `bytes.NewReader(buf)` / `bytes.NewBuffer(buf)`.
- For strings, `strings.NewReader(s)`; a `*bytes.Reader` can rewind with
  `Seek(0, io.SeekStart)`.
- To duplicate a stream while reading it, use `io.TeeReader`, or fan out writes with
  `io.MultiWriter`.
- Reattach a `*bufio.Reader` to a new source with `Reset(r)` — it won't rewind unless the
  source supports seeking.
- For large payloads, avoid unbounded buffering: stream, use `io.LimitReader`, or spill to a
  temp file.

## HTTP request bodies
A consumed `req.Body` can't be reused, which breaks redirects and retries.

- Keep the payload as `[]byte` and set `req.Body = io.NopCloser(bytes.NewReader(buf))` before
  each send.
- Better, set `req.GetBody` so the transport can recreate the body itself:
  ```go
  req.GetBody = func() (io.ReadCloser, error) {
      return io.NopCloser(bytes.NewReader(buf)), nil
  }
  ```

## Streaming with io.Pipe
Use `io.Pipe` to stream without buffering the whole payload: write to the `*io.PipeWriter` in a
separate goroutine while the reader consumes. Always close the writer; on failure use
`CloseWithError(err)`. `io.Pipe` is for streaming, not for rewinding or making a reader
reusable.

**Warning — ordering:** with `io.Pipe`, especially feeding a `multipart.Writer`, all writes
must happen in strict sequential order. Concurrent or out-of-order writes corrupt multipart
boundaries and the stream.

### Streaming multipart/form-data
```go
pr, pw := io.Pipe()
mw := multipart.NewWriter(pw)
// use pr as the request body; set Content-Type to mw.FormDataContentType()
go func() {
    // write all parts to mw in order
    // on error: pw.CloseWithError(err)
    // on success: mw.Close() then pw.Close()
}()
```
Don't store request/in-flight form state on a long-lived client — build it per call. Streamed
bodies aren't rewindable, so for retries/redirects buffer small payloads or provide `GetBody`.
