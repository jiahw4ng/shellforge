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
| How should users obtain the sandbox image? | Manual local build, automatic pull, bundle/build on install | Automatic pull from GHCR | CI (`publish-image.yml`) publishes `ghcr.io/jiahw4ng/shellforge-sandbox:<version>`; the app pulls it when missing. The workflow reads the image from `internal/container/consts.go`; `:<version>` is only pushed from a matching `v*` tag (`v<version>`), `main` pushes go to `:edge` |
| How should users get the Shellforge binary? | Build from source, GoReleaser archives on GitHub Releases, hand-written build matrix, package managers | GoReleaser archives on GitHub Releases for `linux/amd64`, `linux/arm64`, `darwin/arm64` and `darwin/amd64` (stripped `tar.gz`) and `windows/amd64` (unstripped `zip`, because Windows Defender flags `-s -w` binaries), plus `checksums.txt` | `release.yml` runs on `v*` tags and is independent of `publish-image.yml`; it first waits for the image pinned in `internal/container/consts.go` to exist, so a download never points at a missing image, and a tag must equal `v<image version>`. Release notes are GitHub-generated. Binaries are unsigned (macOS binaries get Go's ad hoc signature but are not notarized, so users clear quarantine with `xattr`) and do not report a version (`--version` is a follow-up). Homebrew, Scoop, winget, Windows arm64, and Docker packaging are out of scope |
| What additional host platforms are supported? | macOS, Windows arm64, Linux arm64 | Keep native Windows and Linux/WSL amd64 as the tested contract for the Shellforge binaries | The sandbox image is multi-arch (`linux/amd64` + `linux/arm64`) and Docker creation no longer forces a platform, so arm64 hosts run it natively. macOS (arm64, amd64) and Linux arm64 archives are published as experimental: they compile and pass unit tests in CI (the macOS job has no Docker), but need runtime evidence on real hardware before they count as supported |
| How should prerequisites be diagnosed? | Startup errors only, `shellforge doctor`, install-time checks | Add an explicit diagnostic path before distribution | Current errors are useful but there is no dedicated doctor command |
| How should abandoned containers be recovered? | Manual cleanup, label-based startup cleanup, external janitor | Conservative label-based cleanup needs design | Active-session cleanup is exact and safe, but SIGKILL can leave a labelled container |
| How should the curriculum expand? | More linear lessons, modules/unlocks, independent missions | Keep embedded linear lessons until progression semantics are specified | Legacy specs describe more content and unlocking than the current application implements |

## Settled Decisions

| Decision | Choice | Recorded | ADR |
| --- | --- | --- | --- |
| Outer application model | Bubble Tea v2 `Model`/`Update`/`View` with asynchronous `tea.Cmd` work | 2026-10-08 | n/a; existing implementation |
| Interactive terminal | Shellforge owns a single cross-platform engine using `xpty` for PTY/ConPTY and `x/vt` for emulation | 2026-10-17 | [ADR 0002](adr/0002-own-cross-platform-terminal-engine.md) |
| Lesson packaging | Embed YAML and Markdown in the Go binary and validate them with `lessonlint` | 2026-10-08 | n/a; existing implementation |
| Sandbox lifecycle | One uniquely named, resource-limited, no-network container per session; no host mounts | 2026-10-08 | n/a; existing implementation |
| Verification | Run declarative assertions through separate non-interactive commands against the active container | 2026-10-08 | n/a; existing implementation |
| Check timing | Learner explicitly runs F12 checks; do not inspect after every command | 2026-10-08 | n/a; existing implementation |
| Completion storage | Store completed lesson IDs in private per-user SQLite state | 2026-10-08 | n/a; existing implementation |
| Learner settings storage | Store learner settings in a `settings` key/value table in the same SQLite file; `internal/storage` exposes one `storage.Storage` interface composed of `CompletionStore` and `SettingsStore`, and the database file keeps its `completions.db` name so existing progress survives; settings are read synchronously in `app.NewWithStorage` before the program starts, since the local read is negligible and it avoids a repaint | 2026-10-27 | n/a; persistence extension |
| Terminal cursor style | Learners choose one of six `ui.CursorStyles` (block, underline, or bar; blinking or steady) in Settings; the app applies it in `positionedTerminalCursor` and ignores cursor shapes requested by programs inside the sandbox | 2026-10-27 | n/a; presentation contract |
| Learner color theme | Learners choose Dark (default), Light, or High contrast in Settings; `ui.ApplyTheme` swaps the shared palette (the saved theme is read synchronously in `app.NewWithStorage`, so the first frame is already themed; later changes happen in the update loop), `lessonrender` keys its cache by theme and picks Glamour's matching base, and the choice persists under the `theme` key in the existing `settings` table. Contrast targets are enforced by `ui/theme_test.go`. The theme does not reach the Bash pane: the sandbox prompt in `docker/shellforge.bashrc` is fixed bright white, so Light is not fully readable inside it on a light terminal | 2026-11-04 | n/a; presentation contract, no new package boundary, schema, or sandbox change |
| Startup failure presentation | Docker prerequisite failures are typed (`container.ContainerUnavailableError` with a `Reason` and the original cause); the learner sees a summary, a remedy, the return key, one line of Docker's own output, and the log path from `terminal.Describe`. Stage names, the sentinel text, and the full error chain go to the log only, once. The remedy wording is OS-neutral ("Start Docker") because the supported platforms differ and per-OS commands are untested. Unrecognised errors keep showing their own text | 2026-11-06 | n/a; no new package boundary or schema |
| Key vocabulary | Centralize input and help bindings in `internal/config/keymap.go` | 2026-10-08 | n/a; existing implementation |
| Semantic string ownership | Name repeated or behavior-bearing strings in the package that owns the contract; keep one-off prose and declarative syntax local | 2026-10-16 | n/a; package-boundary convention |
| Lesson presentation hierarchy | Use one blue accent for headings and controls, readable neutral text for explanations, and strong colors only for success, warnings, and errors | 2026-10-17 | n/a; presentation contract |
| Learner progress states | Present unmet assertions as neutral pending tasks; keep the bottom status white before/during checks, restore the F12 prompt when work remains, and reserve green/red summaries for complete checks or execution failures | 2026-10-17 | n/a; presentation contract |

## Pressure Points

- Terminal startup, reset, assertion, and exit messages race naturally; generation
  tags and idempotent cleanup are required correctness mechanisms.
- `xpty` and `x/vt` are pre-v1 terminal dependencies behind `internal/terminal`;
  preserve PTY process ownership, ANSI/cursor behavior, stale-output filtering,
  and deterministic worker shutdown when upgrading them.
- The sandbox image name is duplicated between the Makefile default and
  `internal/container/consts.go`, and the README names its version. Both CI
  workflows read `consts.go`, and a release tag must equal `v<image version>`, so
  bump the Makefile and README with it. The release workflow waits for the
  pinned image before it creates a release.
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
