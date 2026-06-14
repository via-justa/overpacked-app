# Error Handling

Errors are values in Go, and handling them well is most of what separates robust Go from
fragile Go. The guiding idea: check errors where they happen, add context as they travel up,
and let callers decide what to do.

## Basics
- Check an error immediately after the call that produced it.
- Don't ignore errors with `_` unless you have a real reason — and document that reason.
- Error is the last return value. Name the variable `err`.
- Keep error message text lowercase and without trailing punctuation, because messages get
  wrapped and concatenated: `failed to open file` reads well inside a larger chain.

## Creating errors
- `errors.New` for simple static errors.
- `fmt.Errorf` for dynamic messages.
- Custom error types when callers need to inspect domain-specific detail.
- Export sentinel error variables (`var ErrNotFound = errors.New("not found")`) when callers
  need to compare against a known error.

## Wrapping and checking
- Add context as you propagate, using `fmt.Errorf` with the `%w` verb so the original error
  stays unwrappable: `fmt.Errorf("loading config: %w", err)`.
- Check wrapped errors with `errors.Is` (sentinel comparison) and `errors.As` (type
  extraction) rather than string matching or `==`.

## Propagation discipline
- Handle each error at the level that can actually do something about it.
- Don't both log *and* return an error — pick one, or you'll get duplicate noise up the stack.
  Generally, return it and let the top level log once.
- Consider structured errors (custom types carrying fields) when richer debugging context
  helps.
