# Architecture

## System Shape

Shellforge is a single Go binary with an Elm-style Bubble Tea state machine. The
binary embeds lesson content, starts one disposable Docker container for each
terminal session, connects a host PTY or ConPTY to an interactive Bash process, and runs
progress assertions through separate non-interactive Docker commands. Completion
and diagnostic data stay in private per-user files outside the repository.

## Bounded Contexts

| Context | Owns | Does Not Own | Public Entry Point |
| --- | --- | --- | --- |
| Application | Bubble Tea state, navigation, command scheduling, message handling, and cross-context orchestration | Docker commands, SQLite details, lesson parsing, terminal rendering primitives | `internal/app.New`, `internal/app.NewWithCompletionStore`, `app.State` |
| Learning content | Lesson schema, embedded YAML/Markdown loading, and authoring validation | TUI state, sandbox execution, progress persistence | `internal/lessons`, `lessondata.Files`, `cmd/lessonlint` |
| Progress assertions | Assertion vocabulary, validation, execution, and results | Container lifecycle and UI rendering | `assertion.Assertion.Validate`, `assertion.Evaluate` |
| Terminal session | Cross-platform PTY, virtual-terminal emulation, exit notification, and assertion access to the active sandbox | Docker policy and application navigation | `internal/terminal.Start`, `terminal.TermSession` |
| Sandbox runtime | Docker CLI arguments, availability checks, container limits, trusted setup, exact-session removal, and non-interactive execution | TUI state and lesson semantics | `internal/container.CreateAndStart`, `container.Container` |
| Storage | SQLite schema, completed-lesson operations, and learner settings such as cursor style and theme | Navigation and presentation | `internal/storage.Open`, `storage.Store`, `storage.Storage` composed of consumer-facing `storage.CompletionStore` and `storage.SettingsStore` |
| Presentation | Screen rendering, shared styles, learner-selectable themes and cursor styles (`internal/ui`), disclosure panels, key vocabulary, and Markdown-to-ANSI rendering | Blocking I/O and lifecycle orchestration | `internal/screens`, `internal/ui`, `internal/lessonrender`, `internal/config` |
| Diagnostics | Private file-backed structured logging | User-facing status and recovery decisions | `internal/logging.Configure` |

## Dependency Direction

```text
cmd/shellforge
  -> internal/app
       -> config, screens, lessons, terminal, assertion, storage, ui, logging
internal/screens -> config, lessons, assertion, lessonrender, ui
internal/terminal -> container, assertion, lessons, ui, xpty, x/vt
internal/assertion -> small consumer-owned Executor interface
internal/storage -> database/sql + modernc SQLite adapter
lessons/embed.go -> version-controlled YAML and Markdown
```

Keep lower-level packages independent of `internal/app`. Presentation packages
return strings or dimensions and perform no Docker, database, or filesystem
work. External systems remain behind the packages that own them.

## Runtime Flows

### Startup And Stored-Data Loading

1. `cmd/shellforge` configures file logging, attempts to open `storage.Store`,
   and passes it once to `app.NewWithStorage`.
2. `app.NewWithStorage` reads the saved learner settings (cursor style and theme)
   synchronously under a one-second bound and applies them, so the first frame is
   already themed. `app.State.Init` then schedules lesson and completion loading
   as `tea.Cmd` work.
3. Messages merge successful results into state; storage failure is logged and
   does not prevent use of the application.

### Terminal Session

1. Navigation enters a sandbox or lesson screen and increments the terminal
   generation.
2. A timed `tea.Cmd` asks `terminal.Start` to create the container, run trusted
   lesson setup as root when present, and attach `xpty` to Bash as `student`.
3. `xpty` supplies a Unix PTY or Windows ConPTY. `x/vt` handles ANSI state,
   cursor state, terminal input encoding, and rendering. Shellforge owns output
   notification, resize, process wait, and shutdown ordering, including killing
   and reaping the owned host-side `docker exec` process.
4. Process completion closes a receive-only channel. A waiting `tea.Cmd`
   converts that close into a generation-tagged Bubble Tea message; terminal
   output messages also carry terminal identity and ignore stale sessions.
5. Current-generation exits close the terminal engine and force-remove exactly
   that session's container. Stale start, output, exit, and assertion messages
   are ignored.
6. A failed start carries its cause end to end. `container` returns
   `*ContainerUnavailableError{Reason, Cause}` (matching `ErrContainerUnavailable`
   with `errors.Is`) and keeps Docker's output in a `*DockerCommandError`.
   `terminal.Start` logs the full chain once and wraps it in `TerminalStartError`.
   `terminal.Describe` turns a recognised reason into a learner-facing summary,
   remedy, and one line of Docker's output; `app` passes those plain strings, the
   return-key label, and `logging.Path()` to `screens.TerminalProblem.View`. Stage
   names and the sentinel text stay in the log; unrecognised errors still show
   their own text.

### Progress Check

1. F12 schedules a five-second assertion command for the active generation.
2. `assertion.Evaluate` asks the active container to run fixed Bash scripts as
   `student` in `/home/student/workspace` through separate `docker exec` calls.
3. One result is returned per declared assertion. A missing outcome is a failed
   check; an execution error is reported as an inability to check.
4. Passing all results marks the lesson complete in memory and asynchronously
   persists its ID when storage is available.

### Lesson Content

`lessons/embed.go` embeds `*.yaml` and referenced `*/*.md` files. Runtime loading
parses and hydrates content; `cmd/lessonlint` owns semantic authoring validation
and runs in `make check` so invalid shipped content does not reach a release.

## State And Storage

- `app.State` groups navigation, lesson, settings, terminal, viewport, and help
  state. `State.Update` is a thin message dispatcher.
- Completed lesson IDs and learner settings (currently the cursor style and theme IDs) are
  stored at `~/.local/state/shellforge/completions.db`; the directory is mode
  `0700` and database is mode `0600` on supported systems. Settings are read once,
  before the program starts, so there is no late load to guard against.
- The active theme is process-wide presentation state in `internal/ui`:
  `ui.ApplyTheme` rebuilds the shared colors and styles and is called only through
  `State.setTheme`: by `NewWithStorage` before the program starts and by
  `State.Update` handlers afterwards, never from background commands.
  Anything that copies a style at construction (the help model) must be
  restyled on a theme change, and `lessonrender` keys its cache by theme ID.
- Diagnostics are appended to `~/.local/state/shellforge/shellforge.log` with
  mode `0600`. `logging.Path()` reports the file only after it was opened, so the
  UI never points the learner at a log that does not exist.
- Command history, latest working directories, and environment snapshots live
  only inside each disposable container and are educational evidence, not
  tamper-proof records.

## Boundary And Lifecycle Rules

1. Blocking container, assertion, and persistence work runs in `tea.Cmd`
   functions with explicit timeouts; never block `State.Update`.
2. Message handlers that return a new model use value receivers. Mutation-only
   helpers use pointer receivers.
3. Preserve terminal generations on all asynchronous session-scoped messages.
4. `internal/container` is the only package that constructs or executes Docker
   CLI arguments. Never add privileged mode, networking, host mounts, or Docker
   socket access to lesson containers.
5. Trusted lesson setup runs as root before the interactive shell; learner work
   and assertions run as `student` unless a new contract is explicitly designed.
6. A `TermSession` owns its PTY, virtual terminal, process wait, and container. Every failure after
   container creation must remove that exact container.
7. Keep key bindings in `internal/config`; use the same bindings for matching and
   help rendering.
8. Keep lesson renderers presentation-only and preserve fixed header/footer
   chrome around the scrollable guide.
9. Keep SQLite behind `storage.Storage`; application state holds only the
   narrower `storage.CompletionStore` and `storage.SettingsStore` views, and
   tests use a small fake rather than the real user database.
10. Do not introduce a new package or interface unless it represents a cohesive
    responsibility or a real substitution/boundary need.

## Security Boundary

The Docker container reduces accidental host damage but shares the host kernel
and is not a virtual machine. The current safety contract is no network, no host
mounts, no privileged mode, default Docker security policy, 256 MiB memory,
0.5 CPU, and 128 PIDs. Container-local passwordless sudo is intentional for
lessons and must never imply host privilege.
