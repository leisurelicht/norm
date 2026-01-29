---
name: code-review
description: Review Go code for correctness, security, performance, and maintainability. Use when reviewing pull requests, examining code changes, or when the user asks for a code review.
---

# Code Review

## Checklist

- [ ] Logic correct, edge cases handled
- [ ] No SQL injection, proper parameterized queries
- [ ] Error handling complete (no silently ignored errors)
- [ ] Context propagation correct
- [ ] No data races in concurrent code
- [ ] Tests cover changes
- [ ] Follows gofmt/goimports-reviser formatting

## Go-Specific Checks

### Error Handling
- Every error checked or explicitly ignored with `_`
- Errors wrapped with context using `fmt.Errorf("...: %w", err)`
- No `panic` in library code

### Interfaces
- Accept interfaces, return concrete types
- Interface segregation (small, focused interfaces)
- Verify interface compliance with `var _ Interface = (*Impl)(nil)`

### Concurrency
- Goroutines have clear ownership and lifecycle
- Channels properly closed by sender
- Mutex usage correct (defer unlock, no nested locks)

### Performance
- Avoid allocations in hot paths
- Pre-allocate slices when size known
- Consider real performance gain

## Project-Specific Checks

### Controller Pattern
- Method chaining returns `Controller` interface
- `setCalled` tracks method calls
- `preCheck` validates operation compatibility
- Reset clears state properly

### Database Operations
- Use parameterized queries (no string interpolation)
- Validate column names against `fieldNameMap`
- Handle `ErrNotFound` and `ErrDuplicateKey`

## Feedback Format

- 🔴 **Critical**: Must fix
- 🟡 **Suggestion**: Should consider
- 🟢 **Nitpick**: Optional
