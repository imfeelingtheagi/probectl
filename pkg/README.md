# pkg/

Shared, public Go libraries — packages intended to be importable by external
consumers without pulling in control-plane internals. The split is enforced by
the Go toolchain itself: the compiler refuses any import of an `internal/`
package from outside this module, so `pkg/` is the one deliberate shelf of
exported code — putting something here is a publishing decision, not a file
move.

## Status

Empty. It is populated only as stable, reusable surfaces emerge — for example a
thin client library or published SDK types. (Note that the generated protobuf
Go currently lives in `internal/gen/`, private on purpose; it would move or be
re-exported here only as a deliberate publishing decision.) Keep this directory
free of control-plane business logic.
