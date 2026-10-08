# Coding Policy

This policy is mandatory for Shellforge changes. It adapts the
[Google Go Style Best Practices](https://google.github.io/styleguide/go/best-practices)
to this repository. That guide describes itself as auxiliary rather than
canonical, so established Shellforge code, tests, package boundaries, and
configured tooling take precedence when guidance conflicts.

## Before Editing

1. Read the task workflow selected by `.beryl/agent/task-routing.md` and the
   smallest relevant canonical project files.
2. Trace behavior from the exact entry point, message, lesson definition, or
   failing test before generalizing.
3. Do not implement a feature until the user ratifies a plan. State success
   checks and commit boundaries before implementation.
4. Preserve unrelated worktree changes. Do not edit generated shims directly;
   edit `.beryl/agent/tool-instruction-template.md` and synchronize them.
5. Do not add or upgrade dependencies, install a new toolchain, or introduce a
   JavaScript package manager without explicit user approval.

## Formatting And Imports

- Let repository tooling decide formatting: run `make fmt`, which applies
  `gofumpt` and `goimports` through golangci-lint v2. Do not hand-format around
  those tools.
- Keep imports in formatter-defined groups and avoid aliases unless they resolve
  a real collision or clarify a conventional package role.
- Use Go modules and the Makefile commands. Do not edit `go.sum` manually.
- Keep lines and files focused enough that maintainers can find behavior by
  responsibility. Go does not require one type per file; avoid both giant files
  and fragmentation into trivial files.

## Names And Package Shape

- Choose package names for what they provide at the call site. Do not create
  vague `util`, `helper`, or `common` packages; add behavior to an existing
  cohesive package or name a new responsibility precisely.
- Do not repeat the package, receiver, parameter, pointer, or return type in a
  function name unless needed to disambiguate related operations.
- Use noun-like names for queries and verb-like names for actions. Avoid `Get`
  when a concise noun communicates the lookup.
- Prefer whole words and established project vocabulary: lesson, page, guide,
  sandbox, terminal session, progress check, assertion, and completion.
- Avoid shadowing standard packages and outer variables in longer scopes. Use
  `=` rather than `:=` when a nested block must update an existing variable.
- Keep packages cohesive. Types that clients must use together or that share
  tightly coupled unexported implementation normally belong together; split a
  package only for a concept that stands on its own.

## Functions, Receivers, And Interfaces

- Keep signatures short and call sites readable. Use a small options struct only
  when several related parameters make calls ambiguous; do not introduce
  functional options without enough optional complexity to justify them.
- Pass `context.Context` explicitly as the first parameter for cancelable work;
  never place it in an options struct or store it for unrelated future calls.
- Specify channel direction wherever ownership is one-way.
- Use pointer receivers when a method mutates state, owns a lock or other
  non-copyable field, or must preserve object identity. Value receivers are
  appropriate for Shellforge message handlers that intentionally return a new
  Bubble Tea model.
- Avoid speculative interfaces. Define the smallest interface in the consuming
  package when a real substitution or boundary exists. Accept interfaces and
  return concrete types by default.
- Keep exported APIs small. Do not export test backdoors or abstractions that
  have only one hypothetical implementation.

## Declarations, Collections, And Strings

- Use `:=` for new non-zero local values and `var` when the meaningful initial
  state is the zero value or a later decode fills the value.
- Use composite literals when initial members are known. Initialize maps before
  writing; nil slices and maps are valid read-only empty values.
- Preallocate only when the final size is known or measurement shows value. Do
  not add capacity hints as decoration.
- Use `+` for a few simple string pieces, `fmt.Sprintf` for formatting, and
  `strings.Builder` for incremental construction.
- Treat package state with scrutiny. Logically constant style/key definitions and
  behavior-invisible caches are acceptable; mutable service locators, registries,
  and replaceable global clients are not. Do not mutate `config.Keys` or shared
  styles in tests.

## Documentation And Comments

- Document every exported package, type, function, method, constant group, and
  sentinel error that is not self-evident to Go documentation tools.
- Start an exported declaration's comment with its exact name. For methods, put
  a third-person verb immediately after the method name, for example,
  `Close closes`, `Send sends`, or `Render returns`.
- Explain contracts, ownership, lifecycle, cleanup, surprising concurrency,
  security boundaries, error semantics, and why a choice exists. Do not narrate
  obvious syntax or enumerate every parameter mechanically.
- Use a blank comment line between Godoc paragraphs. Prefer runnable examples
  when usage is easier to demonstrate than describe.
- Document cleanup obligations and any concurrency behavior that differs from
  normal Go expectations. Call attention to unusual conditions such as
  `err == nil` when they are easy to misread.
- Remove stale TODOs when completing them. New TODOs must identify a concrete
  missing decision or follow-up; do not use them as a substitute for handling an
  error or finishing the requested slice.

## Errors, Logging, And Cleanup

- Handle every error deliberately. Return it, translate it at a boundary, or log
  it when no caller can act; do not silently discard it unless cleanup is already
  best-effort and the choice is explicit.
- Add concise operation context without repeating details already supplied by
  the underlying error. Error text is lowercase and has no trailing punctuation.
- Wrap with `%w` only when callers should use `errors.Is` or `errors.As`; use
  `%v` or a new boundary error when the underlying implementation must not become
  API. Place ordinary `%w` wrapping at the end as `operation: %w`.
- Use `ErrName` sentinels or typed errors for conditions callers must classify.
  Never branch on error strings.
- Usually log or return an error at one layer, not both. Use structured `slog`
  fields, keep messages actionable, and never log secrets or sensitive learner
  data.
- Return errors rather than panic for runtime failures. Panic is reserved for a
  truly unreachable invariant or tightly contained internal mechanism whose
  public boundary converts it back to an error.
- Acquire and release resources in the same ownership layer. Register cleanup
  immediately after successful acquisition. Preserve idempotent container
  removal and close files, databases, PTYs, and cancellation functions.

## Bubble Tea And Concurrency

- Keep `State.Update` a thin event dispatcher. Put concrete message handling in
  `internal/app/handlemessages.go`, screen-specific key routing in
  `internal/app/keypress.go`, and blocking work in `tea.Cmd` functions.
- Never block the Bubble Tea event loop with Docker, SQLite, assertion, PTY, or
  filesystem work. Apply explicit timeouts at external-operation boundaries.
- Tag all asynchronous terminal-scoped results with the terminal generation and
  ignore stale messages after reset or replacement.
- A `TermSession` owns its emulator, sandbox, and exit signal. Preserve the
  receive-only exit channel and one-time notification/removal semantics.
- Shellforge-reserved lesson keys are consumed by the app; all other terminal
  input is forwarded to Bubbleterm. Standalone sandbox keys are not silently
  repurposed as lesson controls.
- Keep screen and renderer packages free of external side effects. They receive
  state and return presentation output.

## Docker And Lesson Safety

- Construct Docker commands as argument slices in `internal/container`; never
  concatenate learner-controlled input into a host shell command.
- Preserve no network, no host mounts, no Docker socket, no privileged mode,
  exact generated container names, resource limits, and forced exact-session
  cleanup.
- Lesson setup is trusted repository content and may run as container root.
  Interactive work and progress checks run as `student` unless an approved
  design explicitly changes the boundary.
- Keep assertion paths and semantics container-local. History, working-directory,
  and environment snapshot files are learner-editable evidence, not security
  controls.
- When editing lessons, keep YAML and referenced Markdown aligned and run
  `make lessonlint`. Update loader/app expectations when lesson IDs, ordering, or
  page structure changes.

## Tests

- Test observable behavior and lifecycle boundaries, not private implementation
  trivia. Cover a normal path and a meaningful edge, failure, or stale-event path.
- Keep checks and failure decisions in the `Test` function. Use `t.Helper` for
  setup helpers; do not build assertion-helper frameworks.
- Use `t.Error` to keep checking independent expectations. Use `t.Fatal` only
  when the current test or subtest cannot continue, and never call it from a
  goroutine other than the test goroutine.
- Prefer table-driven subtests for genuinely shared behavior. Use keyed struct
  literals when field identity matters.
- Use `t.TempDir`, `t.Cleanup`, small consumer-owned fakes, and real parsers,
  renderers, SQLite, Bash scripts, and transport paths where practical.
- Do not access the user's real state directory or require Docker for unit tests.
  Keep the real Docker/Bubbleterm path in the integration test and report skips.
- Never weaken an existing test to pass a change. Intentional test edits require
  `./.beryl/scripts/update-test-manifest.sh`.

## Required Verification

After edits, run the formatter, the smallest focused checks from
`.beryl/agent/testing-policy.md`, `make check`, and
`./.beryl/scripts/check.sh`. Build the binary or sandbox image when the changed
boundary requires it. Report exact failures, skips, and unavailable runtime
evidence.
