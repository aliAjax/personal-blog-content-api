# Bug Reproduction

## Bug

Updating an article without a `tag_ids` field removes every existing tag relation.

## Trigger

Run `go test ./internal/article -run '^TestUpdateKeepsTagsWhenTagIDsAreOmitted$' -count=1`.

## Error

The repository unexpectedly executes `DELETE FROM article_tags WHERE article_id = ?` and the test fails before commit.
