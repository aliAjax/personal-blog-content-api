# Bug Reproduction

## Bug

A reply can reference a parent comment owned by a different article.

## Trigger

Run `go test ./internal/comment -run '^TestCreateRejectsParentFromAnotherArticle$' -count=1`.

## Error

The service returns no error and persists the invalid reply; the test reports `expected a cross-article parent to be rejected`.
