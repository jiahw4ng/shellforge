# Design Tree

## Current Design Concept

Shellforge combines a guided, state-driven terminal UI with a real Bash session
inside one disposable Docker container. Content is immutable and embedded in the
binary; learner actions change only the container. The application verifies
observable outcomes on demand, then records only durable lesson completion in a
small local SQLite store.

## Open Decisions

| Decision | Options | Current Lean | Why |
| --- | --- | --- | --- |
| How should users obtain the sandbox image? | Manual local build, automatic pull, bundle/build on install | Undecided; current manual build | Released binaries still require the exact local `shellforge-sandbox:0.1.0` image |
| What host platforms are supported? | WSL 2 amd64 only, broader Linux, macOS, multi-architecture | Stabilize the documented WSL 2 amd64 path first | Docker creation currently forces `linux/amd64`; broader claims need runtime evidence |
| How should prerequisites be diagnosed? | Startup errors only, `shellforge doctor`, install-time checks | Add an explicit diagnostic path before distribution | Current errors are useful but there is no dedicated doctor command |
| How should abandoned containers be recovered? | Manual cleanup, label-based startup cleanup, external janitor | Conservative label-based cleanup needs design | Active-session cleanup is exact and safe, but SIGKILL can leave a labelled container |
| How should the curriculum expand? | More linear lessons, modules/unlocks, independent missions | Keep embedded linear lessons until progression semantics are specified | Legacy specs describe more content and unlocking than the current application implements |

## Settled Decisions

| Decision | Choice | Recorded | ADR |
| --- | --- | --- | --- |
| Outer application model | Bubble Tea v2 `Model`/`Update`/`View` with asynchronous `tea.Cmd` work | 2026-10-08 | n/a; existing implementation |
| Interactive terminal | Bubbleterm owns PTY and terminal emulation for `docker exec -it` Bash | 2026-10-08 | n/a; existing implementation |
| Lesson packaging | Embed YAML and Markdown in the Go binary and validate them with `lessonlint` | 2026-10-08 | n/a; existing implementation |
| Sandbox lifecycle | One uniquely named, resource-limited, no-network container per session; no host mounts | 2026-10-08 | n/a; existing implementation |
| Verification | Run declarative assertions through separate non-interactive commands against the active container | 2026-10-08 | n/a; existing implementation |
| Check timing | Learner explicitly runs F12 checks; do not inspect after every command | 2026-10-08 | n/a; existing implementation |
| Completion storage | Store completed lesson IDs in private per-user SQLite state | 2026-10-08 | n/a; existing implementation |
| Key vocabulary | Centralize input and help bindings in `internal/config/keymap.go` | 2026-10-08 | n/a; existing implementation |

## Pressure Points

- Terminal startup, reset, assertion, and exit messages race naturally; generation
  tags and idempotent cleanup are required correctness mechanisms.
- Bubbleterm is a runtime PTY/emulator dependency and indirectly uses
  `github.com/creack/pty`; do not replace it with a plain text viewport and still
  claim interactive terminal correctness.
- The sandbox image name is duplicated between the Makefile default and
  `internal/container`; release work must make that contract configurable or
  guarantee image availability.
- Legacy files under `docs/` mix implemented behavior with proposals. Verify
  every claim against current code and tests before promoting it.
- Lesson setup is trusted repository-authored Bash. Learner commands and
  container-local tracking files are untrusted.
- The integration test needs Docker and the image, and intentionally skips only
  when the container runtime is unavailable. Unit tests must remain useful
  without Docker.

## Recording Rule (Design Tree Vs ADR)

Update this file while a choice is still being compared or may change after one
or two iterations. Create an ADR when a decision changes package boundaries,
the lesson schema, persistence shape, sandbox security, adapter contracts,
cross-package terminology, or test strategy.
