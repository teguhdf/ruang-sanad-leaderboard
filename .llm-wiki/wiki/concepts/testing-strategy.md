---
type: concept
status: stub
---

# Testing Strategy

Laju Go uses in-memory Go tests for all business logic and database access.

## Go Unit/Integration Tests

- **Scope**: Services, queries, handlers, cache
- **Command**: `go test ./...`
- **Setup**: In-memory SQLite, no external dependencies
- **Exception**: Test files (`*_test.go`) may call queries directly for test data setup (bypassing three-tier rule)
