# Bug Reproduction

## Bug

An HTTP panic after a partial response leaves the original success status and corrupts the JSON body.

## Trigger

Run `go test ./pkg/middleware ./pkg/response -run '^(TestRecoverDiscardsResponseWrittenBeforePanic|TestJSONDoesNotCommitSuccessBeforeEncoding)$' -count=1`.

## Error

The recorder keeps status `201` with a partial JSON prefix, while direct encoding failure keeps `200` and returns encoder internals.
