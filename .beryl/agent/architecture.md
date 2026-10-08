# Architecture

## System Shape

Shellforge is a single Go binary with an Elm-style Bubble Tea state machine. The
binary embeds lesson content, starts one disposable Docker container for each
terminal session, connects Bubbleterm to an interactive Bash process, and runs
progress assertions through separate non-interactive Docker commands. Completion
and diagnostic data stay in private per-user files outside the repository.

## Bounded Contexts

| Context | Owns | Does Not Own | Public Entry Point |
| --- | --- | --- | --- |
| Application | Bubble Tea state, navigation, command scheduling, message handling, and cross-context orchestration | Docker commands, SQLite details, lesson parsing, terminal rendering primitives | `internal/app.New`, `internal/app.NewWithCompletionStore`, `app.State` |
| Learning content | Lesson schema, embedded YAML/Markdown loading, and authoring validation | TUI state, sandbox execution, progress persistence | `internal/lessons`, `lessondata.Files`, `cmd/lessonlint` |
| Progress assertions | Assertion vocabulary, validation, execution, and results | Container lifecycle and UI rendering | `assertion.Assertion.Validate`, `assertion.Evaluate` |
| Terminal session | Bubbleterm lifecycle, PTY-facing updates, exit notification, and assertion access to the active sandbox | Docker policy and application navigation | `internal/terminal.Start`, `terminal.TermSession` |
| Sandbox runtime | Docker CLI arguments, availability checks, container limits, trusted setup, exact-session removal, and non-interactive execution | TUI state and lesson semantics | `internal/container.CreateAndStart`, `container.Container` |
| Completion persistence | SQLite schema and completed-lesson operations | Navigation and presentation | `internal/completion.Open`, `completion.Store`, consumer-owned `completion.CompletionStore` |
| Presentation | Screen rendering, shared styles, disclosure panels, key vocabulary, and Markdown-to-ANSI rendering | Blocking I/O and lifecycle orchestration | `internal/screens`, `internal/ui`, `internal/lessonrender`, `internal/config` |
| Diagnostics | Private file-backed structured logging | User-facing status and recovery decisions | `internal/logging.Configure` |

## Dependency Direction

```text
cmd/shellforge
  -> internal/app
       -> config, screens, lessons, terminal, assertion, completion, ui
internal/screens -> config, lessons, assertion, lessonrender, ui
internal/terminal -> container, assertion, lessons, ui, Bubbleterm
internal/assertion -> small consumer-owned Executor interface
internal/completion -> database/sql + modernc SQLite adapter
lessons/embed.go -> version-controlled YAML and Markdown
```

Keep lower-level packages independent of `internal/app`. Presentation packages
return strings or dimensions and perform no Docker, database, or filesystem
work. External systems remain behind the packages that own them.

## Runtime Flows

### Startup And Completion Loading

1. `cmd/shellforge` configures file logging and attempts to open the completion
   store.
2. `app.State.Init` schedules lesson and completion loading as `tea.Cmd` work.
3. Messages merge successful results into state; storage failure is logged and
   does not prevent use of the application.

### Terminal Session

1. Navigation enters a sandbox or lesson screen and increments the terminal
   generation.
2. A timed `tea.Cmd` asks `terminal.Start` to create the container, run trusted
   lesson setup as root when present, and attach Bubbleterm to Bash as `student`.
3. Bubbleterm handles PTY input, output, ANSI state, and resize. Shellforge owns
   reserved lesson shortcuts and forwards other keys.
4. The terminal exit callback closes a receive-only channel. A waiting `tea.Cmd`
   converts that close into a generation-tagged Bubble Tea message.
5. Current-generation exits close Bubbleterm and force-remove exactly that
   session's container. Stale start, exit, and assertion messages are ignored.

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
- Completed lesson IDs are stored at
  `~/.local/state/shellforge/completions.db`; the directory is mode `0700` and
  database is mode `0600` on supported systems.
- Diagnostics are appended to `~/.local/state/shellforge/shellforge.log` with
  mode `0600`.
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
6. A `TermSession` owns both Bubbleterm and its container. Every failure after
   container creation must remove that exact container.
7. Keep key bindings in `internal/config`; use the same bindings for matching and
   help rendering.
8. Keep lesson renderers presentation-only and preserve fixed header/footer
   chrome around the scrollable guide.
9. Keep SQLite behind the consumer-owned `completion.CompletionStore` interface;
   tests use small fakes rather than the real user database.
10. Do not introduce a new package or interface unless it represents a cohesive
    responsibility or a real substitution/boundary need.

## Security Boundary

The Docker container reduces accidental host damage but shares the host kernel
and is not a virtual machine. The current safety contract is no network, no host
mounts, no privileged mode, default Docker security policy, 256 MiB memory,
0.5 CPU, and 128 PIDs. Container-local passwordless sudo is intentional for
lessons and must never imply host privilege.
