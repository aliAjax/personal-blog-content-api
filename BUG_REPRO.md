# Bug Reproduction

## Bug

Concurrent application startup can abort when two instances seed the same unique row.

## Trigger

Run `go test ./internal/seed -run '^TestSecondStartupSurvivesConcurrentCategoryInsert$' -count=1`.

## Error

The second startup returns `Error 1062: Duplicate entry 'golang'` after another instance inserts between its lookup and insert.
