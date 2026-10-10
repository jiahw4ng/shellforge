# Go Style Refactor Checklist

Use this checklist to bring Shellforge into conformance with
`.beryl/agent/coding-policy.md` through small, behavior-preserving commits.
Google's Go guidance is auxiliary; current Shellforge behavior, tests, package
boundaries, and security constraints remain authoritative.

## Working Rules

- [ ] Select one package and one policy concern per commit.
- [ ] Trace every affected call site with `rg` before editing.
- [ ] Run the focused package tests before and after the change.
- [ ] Keep mechanical, documentation, API, and behavioral changes in separate
      commits unless they cannot compile independently.
- [ ] Do not add dependencies or change lesson, UI, persistence, or sandbox
      behavior as incidental cleanup.
- [ ] Run `make fmt`, the focused checks, `make check`, and
      `./.beryl/scripts/check.sh` after every slice.
- [ ] If a test file changes intentionally, run
      `./.beryl/scripts/update-test-manifest.sh` and commit the manifest with it.
- [ ] Enable a new repository-wide linter only after its existing findings have
      been resolved.
- [ ] Write and ratify a focused mini-plan before any context, lifecycle,
      concurrency, persistence, or cross-package interface change.

## Completed Slices

- [x] Rename `getLessonRenderParams` to `lessonRenderParams` in
      `internal/app/view.go`.
  - Scope: one source file; no behavior or test changes.
  - Evidence: `go test -count=1 ./internal/app` and `make check` passed.

## Phase 1: Documentation And Small Names

Complete each checkbox as a separate commit unless the listed files must change
together to compile.

- [x] Replace the placeholder package comment and document the exported screen
      vocabulary in `internal/config/screens.go`.
- [x] Document `Keys` and the exported help-binding groups in
      `internal/config/keymap.go`.
- [x] Document or make package-private the exported frame constants and style in
      `internal/ui/frame.go`.
- [x] Document the exported colors and styles in `internal/ui/styles.go`.
- [x] Add the missing comment for `DimensionWithFallback` in
      `internal/ui/utils.go`.
- [x] Add the missing `cmd/shellforge` package comment.
- [x] Document `container.ErrContainerUnavailable` and the blank SQLite driver
      import's purpose.
- [x] Replace the stale TODO on terminal error unwrapping with the actual
      classification contract.
- [x] Rename the container argument helpers to remove unnecessary `get` prefixes:
      `execCommandArguments`, `setupCommandArguments`, and
      `createContainerArguments`.
- [x] Rename vague implementation files such as `internal/app/others.go` and
      `internal/container/getarguments.go` only when the new name clearly states
      their responsibility.
- [x] Review comments that narrate syntax or temporary implementation steps and
      retain only contract, ownership, lifecycle, or rationale documentation.

Focused checks:

```bash
go test -count=1 ./internal/config ./internal/ui ./internal/container ./internal/terminal ./internal/app
make check
```

## Phase 2: Export Surface And API Names

- [ ] Make assertion shell scripts package-private; update their production and
      package-test references together.
- [ ] Make screen titles and UI frame details package-private where they have no
      cross-package caller.
- [ ] Make container synchronization fields private; expose behavior through
      methods rather than mutable fields.
- [ ] Rename `terminal.TerminalStartError` to `terminal.StartError` and preserve
      `errors.Is`/`errors.As` behavior.
- [ ] Replace the static invalid-terminal-start error factory with a documented
      sentinel error if callers need classification.
- [ ] Move the completion-store interface into its consuming `internal/app`
      package while retaining `completion.Store` as the concrete SQLite adapter.
- [ ] Review exported declarations package by package and unexport anything with
      no real cross-package consumer.

Focused checks:

```bash
go test -count=1 ./internal/assertion ./internal/container ./internal/terminal ./internal/completion ./internal/app
make check
```

## Phase 3: Function Signatures And Call-Site Clarity

- [ ] Replace the seven-argument `ui.Disclosure` call with one small parameter
      struct.
- [ ] Replace screen-rendering signatures containing several adjacent strings,
      booleans, or collections with focused parameter structs where the call
      sites are otherwise ambiguous.
- [ ] Confirm every `context.Context` parameter is first and is not stored for
      unrelated future work.
- [ ] Confirm pointer receivers are used for mutation, locks, and identity while
      Bubble Tea handlers intentionally returning new state retain value
      receivers.
- [ ] Confirm channels communicate ownership through `<-chan` or `chan<-` where
      only one direction is valid.

Focused checks:

```bash
go test -count=1 ./internal/ui ./internal/screens ./internal/app ./internal/terminal
make check
```

## Phase 4: Contexts, Errors, And Resource Ownership

Treat each item as behavior-sensitive and ratify its mini-plan before editing.

- [ ] Make completion-store initialization context-aware and use
      `PingContext`/`ExecContext` for SQLite startup work.
- [ ] Use `exec.CommandContext` in assertion script tests and other bounded
      subprocess work.
- [ ] Give the interactive terminal process an explicitly owned lifetime rather
      than inheriting the short startup timeout.
- [ ] Audit every `%w`: retain wrapping only when callers classify the cause;
      otherwise translate the error at the package boundary.
- [ ] Replace `fmt.Errorf` with `errors.New` for static errors.
- [ ] Remove duplicate log-and-return chains so each failure is normally logged
      or returned at one owning layer.
- [ ] Handle file, database, row, PTY, and container cleanup results explicitly;
      document the few cleanup paths that are intentionally best-effort.
- [ ] Preserve idempotent removal of the exact generated container and report
      the first cleanup failure without broadening removal scope.

Focused checks:

```bash
go test -count=1 ./internal/completion ./internal/logging ./internal/assertion ./internal/container ./internal/terminal ./internal/app
make check
```

## Phase 5: Bubble Tea And Terminal Lifecycle

This is the highest-risk phase. Keep generation guards and Docker isolation
unchanged.

- [ ] Identify every Docker, SQLite, assertion, PTY, and filesystem operation
      reachable synchronously from `State.Update`.
- [ ] Move blocking terminal cleanup out of the Bubble Tea event loop and into a
      timed `tea.Cmd`.
- [ ] Ensure lesson reset closes the prior session before installing its
      replacement without accepting stale start, exit, or assertion messages.
- [ ] Preserve the receive-only exit signal and one-time exit notification.
- [ ] Ensure every startup failure after container creation removes that exact
      container.
- [ ] Add focused tests proving cleanup is scheduled asynchronously, remains
      idempotent, and cannot let an old generation modify a new session.
- [ ] Run the real Docker/Bubbleterm integration test and distinguish a runtime
      skip from a passing end-to-end path.

Focused checks:

```bash
go test -count=1 ./internal/container ./internal/terminal ./internal/app
go test -race ./internal/app ./internal/terminal
make check
```

## Phase 6: File Focus And Test Organization

- [ ] Split `internal/app/model_test.go` by navigation, lessons, completion, and
      terminal lifecycle while preserving every assertion.
- [ ] Split `internal/screens/screens_test.go` by lesson, menu/settings, and
      terminal rendering while preserving every assertion.
- [ ] Keep shared test helpers small, call `t.Helper`, and leave failure
      decisions in the test or subtest goroutine.
- [ ] Review large production files for mixed responsibilities; split only when
      each resulting file has a clear cohesive purpose.
- [ ] Regenerate and commit `tests/.manifest.sha256` with every intentional test
      reorganization.

Focused checks:

```bash
go test -count=1 ./internal/app ./internal/screens
make check
```

## Phase 7: Formatting And Permanent Enforcement

- [ ] Change the `goimports` local prefix in `.golangci.yml` to the actual module
      path, `shellforge`, and isolate the resulting import-only diff.
- [ ] Clear all applicable `revive` findings, then enable `revive` in the normal
      lint configuration.
- [ ] Clear and enable `contextcheck` and `noctx` after the context/lifecycle
      phases are complete.
- [ ] Clear and enable `errorlint`, `rowserrcheck`, and `sqlclosecheck` after the
      error/resource phase is complete.
- [ ] Do not enable subjective rules that are not required by
      `.beryl/agent/coding-policy.md` merely to increase the linter count.

Verification:

```bash
make fmt
make fmt-check
make lint
go vet ./...
git diff --check
```

## Final Completion Audit

- [ ] Review every heading in `.beryl/agent/coding-policy.md` against current
      production code and tests; do not infer compliance from a green linter
      alone.
- [ ] Confirm no new dependencies, host access, Docker privileges, networking,
      or broader cleanup targets were introduced.
- [ ] Confirm visible TUI text, key mappings, lessons, SQLite schema, and
      completion behavior remain unchanged unless separately approved.
- [ ] Run the full project gate with writable caches when required:

  ```bash
  env GOCACHE=/tmp/shellforge-gocache \
    GOTMPDIR=/tmp/shellforge-gotmp \
    GOLANGCI_LINT_CACHE=/tmp/shellforge-lint-cache \
    make check
  ```

- [ ] Run `make build` and the real terminal integration test.
- [ ] Run `./.beryl/scripts/check-tests-unchanged.sh`.
- [ ] Run `./.beryl/scripts/check.sh` and report any preserved external-contract
      failure separately; do not overwrite `AGENTS.md` or hand-edit the Beryl
      lock merely to make the gate green.
- [ ] Update `.beryl/agent/architecture.md`, `design-tree.md`,
      `ubiquitous-language.md`, or an ADR only when a durable boundary or design
      decision actually changed.
