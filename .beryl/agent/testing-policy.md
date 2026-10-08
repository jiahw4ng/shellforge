# Testing Policy

## Command Matrix

| Check | Command | Status | Purpose |
| --- | --- | --- | --- |
| Format | `make fmt` | required after Go edits | Applies `gofumpt` and `goimports` through golangci-lint v2 |
| Format check | `make fmt-check` | required | Prints a formatting diff without editing |
| Static analysis | `make lint` | required | Runs the configured standard linters and all enabled staticcheck checks except `QF1008` |
| Lesson validation | `make lessonlint` | required for content/schema changes | Parses, validates, and hydrates all embedded YAML and Markdown lesson material |
| Unit and integration tests | `make test` | required | Runs `go test -count=1 ./...`; the Docker integration test skips only when the runtime is unavailable |
| Project quality gate | `make check` | required | Runs format check, lint, lesson validation, and all Go tests |
| Build | `make build` | required for entrypoint/build changes | Writes the ignored `bin/shellforge` artifact |
| Sandbox image | `make sandbox-image` | required for Dockerfile/Bashrc/runtime-image changes | Builds `shellforge-sandbox:0.1.0` for `linux/amd64` |
| Beryl aggregate gate | `./.beryl/scripts/check.sh` | required after repository edits | Runs agent readiness, Markdown, component, secret, manifest, and configured project checks |
| Test manifest immutability | `./.beryl/scripts/check-tests-unchanged.sh` | required | Detects unauthorized changes in the Go test scope |

`make check` is the authoritative project gate and matches the main CI verify
job. CI also runs `make build`, while a separate job runs `make sandbox-image`.

## Test Layers

### Pure Unit Tests

- Package-local tests cover rendering, state transitions, key routing, lesson
  parsing and validation, assertion command construction, Docker argument
  construction, structured errors, and disclosure behavior.
- Keep failure decisions in the `Test` function. Helpers may call `t.Fatal` only
  for setup or cleanup preconditions and must call `t.Helper`.
- Prefer explicit `got`/`want` messages and table-driven subtests when inputs
  share one behavior. Do not introduce assertion-helper frameworks.

### Filesystem And SQLite Tests

- Use `t.TempDir`; never write tests against the real
  `~/.local/state/shellforge` directory.
- Verify observable persistence across close/reopen, idempotency, reset behavior,
  and private permissions where the operating system supports them.

### Application State Tests

- Drive `State.Update` with concrete Bubble Tea messages and inspect the returned
  model and command.
- Cover both UI text and state/lifecycle invariants when behavior spans the
  state machine and renderer.
- Protect terminal generation guards, async loading failure behavior, completion
  persistence, sandbox reset semantics, and reserved-key routing.

### Docker Integration Test

`internal/app/terminal_integration_test.go` exercises the real path through
Docker, Bubbleterm, the PTY, and the outer state machine. It requires a reachable
Docker engine and the sandbox image. A missing runtime is an explicit skip;
other startup or behavior failures fail the test.

Run `make sandbox-image` before relying on this test as end-to-end evidence.
Never weaken its assertions or turn unexpected failures into skips.

### Lesson Tests

- Every YAML file and referenced Markdown page ships inside the binary.
- When adding, renumbering, or changing lessons, update YAML, Markdown, loader
  expectations, and relevant UI/app tests together.
- `cmd/lessonlint` owns authoring-policy validation; runtime loading intentionally
  limits itself to parsing and hydration.
- Validate unique IDs/numbers, page references, non-empty content, success
  messages, and assertion-specific fields.

## Change-To-Check Map

| Changed area | Focused checks before the full gate |
| --- | --- |
| `internal/app` | `go test -count=1 ./internal/app` |
| `internal/screens`, `internal/ui`, or `internal/lessonrender` | `go test -count=1 ./internal/screens ./internal/ui ./internal/lessonrender` |
| `internal/assertion` | `go test -count=1 ./internal/assertion ./cmd/lessonlint` |
| `internal/container` or `internal/terminal` | `go test -count=1 ./internal/container ./internal/terminal ./internal/app` |
| `internal/completion` or `internal/logging` | `go test -count=1 ./internal/completion ./internal/logging` |
| `internal/lessons`, `lessons/`, or `cmd/lessonlint` | `make lessonlint` and `go test -count=1 ./internal/lessons ./cmd/lessonlint ./internal/app` |
| `docker/` | `go test -count=1 ./internal/container ./internal/app`, then `make sandbox-image` |
| Build, dependency, linter, or CI files | `make check` and `make build`; add `make sandbox-image` when image behavior is affected |
| `.beryl/agent/*.md` only | `./.beryl/scripts/check-md.sh` and `./.beryl/agent/scripts/agent-doctor.sh`, then the Beryl aggregate gate |

After focused checks, run `make check` directly when it has not already run, then
run `./.beryl/scripts/check.sh`.

## Test Modification Rule

Never weaken tests to make implementation pass. An intentional test change is
allowed only when the requested behavior or ratified design changes. After any
intentional test edit:

1. Run `./.beryl/scripts/update-test-manifest.sh`.
2. Commit `tests/.manifest.sha256` with the test change.
3. Explain the behavior change and manifest update in the final response.

The manifest detects changes; branch protection and review provide the stronger
immutability boundary.

## Mocking And Fakes

- Use small consumer-owned interfaces only at real boundaries, as with
  `completion.CompletionStore` and `assertion.Executor`.
- Prefer real parsing, rendering, SQLite, Bash scripts, and Docker argument
  construction in tests.
- Fake external effects when Docker or persistence is not the behavior under
  test. Do not mock domain logic in the same package.
- Use production transport/lifecycle code in integration tests rather than a
  hand-written imitation.

## Generated And Runtime Evidence

- `bin/shellforge` is generated and ignored; do not commit it.
- The Docker image is external build output. Record whether it was rebuilt and
  whether the real integration test ran or skipped.
- For visible TUI changes, assert stable text/layout contracts in unit tests and,
  when practical, manually run `go run ./cmd/shellforge` in a real terminal.
- If a prerequisite is unavailable, report the exact skipped command or test and
  do not claim that layer was verified.
