# Bug Reproduction

## Bug

Creating an article with an existing slug returns HTTP 400 and exposes MySQL duplicate-key details.

## Trigger

Run `go test ./internal/article -run '^TestCreateDuplicateSlugReturnsConflictWithoutDatabaseDetails$' -count=1`.

## Error

The response is `400` with `Error 1062: Duplicate entry 'same-slug' for key 'uk_articles_slug'` instead of a sanitized conflict response.
