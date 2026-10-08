# Project Brief

## Product Goal

Build Shellforge for learners who want practical Unix and Bash experience, so
they can work through guided terminal missions in a disposable environment
without exposing their host filesystem to lesson commands.

## Current Product Surface

Shellforge is a Go terminal UI. It embeds version-controlled lesson YAML and
Markdown, opens either a standalone sandbox or a lesson beside an interactive
Bash terminal, evaluates lesson outcomes on demand, and stores completed lesson
IDs locally.

The checked-in implementation is the source of truth. Files under `docs/` include
older specifications and user stories; treat unimplemented items there as
aspirational until code, tests, and current design records agree.

## Primary Workflows

1. **Explore a sandbox**: Start an interactive container-local Bash session,
   experiment, then return to the main menu with the exact disposable container
   removed.
2. **Complete a lesson**: Select an embedded lesson, read and page through its
   guide, reveal hints, work in the adjacent sandbox, and run F12 progress checks.
3. **Resume progress**: Load completed lesson IDs from the per-user SQLite store,
   show completion checkmarks, and allow an explicitly confirmed full reset.
4. **Author a lesson**: Keep one YAML definition and its referenced Markdown
   pages aligned, then validate metadata, page references, setup, and assertions
   with `make lessonlint`.

## Current Platform Contract

- Shellforge requires the Docker CLI, a reachable Docker engine, and the locally
  built `shellforge-sandbox:0.1.0` image.
- Lesson containers run as `linux/amd64`; WSL 2 with Docker Desktop is the
  documented reference environment.
- The application is a native terminal binary; there is no web frontend or
  JavaScript toolchain.

## Non-Goals Of The Current Implementation

- Host filesystem mounts, host Docker-socket mounts, networking inside lessons,
  or privileged containers.
- Automatic image download, abandoned-container recovery, prerequisite doctor,
  updater, release installer, or multi-platform distribution.
- Automatic checks after every command, score/points, lesson unlocking, or
  automatic navigation to the next lesson.
- Treating command-history assertions or learner-editable tracking files as a
  security boundary.

## External Systems

| System | Why it exists | Interface owner | Failure behavior |
| --- | --- | --- | --- |
| Docker CLI and engine | Creates, starts, executes in, and removes disposable lesson containers | `internal/container` | Returns `ErrContainerUnavailable` for missing prerequisites; the TUI shows a recoverable terminal error |
| Local sandbox image | Supplies Debian, Bash, command-line tools, the `student` account, and controlled Bash configuration | `docker/`, `internal/container` | Startup tells the user to run `make sandbox-image` when the image is absent |
| Bubbleterm and its PTY stack | Runs and emulates the interactive `docker exec -it` session | `internal/terminal` | Startup errors are wrapped as `TerminalStartError`; unexpected exits preserve the last frame |
| Per-user SQLite database | Persists completed lesson IDs | `internal/completion` | The application remains usable in memory and logs persistence failures |
| Per-user log file | Stores diagnostics away from the full-screen TUI | `internal/logging` | Logging falls back to a discard logger if setup fails |

## Definition Of Done

A change is complete only when it satisfies all applicable items:

1. The change follows `.beryl/agent/coding-policy.md` and existing package
   boundaries.
2. Focused behavior tests cover the normal path and at least one meaningful edge
   or lifecycle path.
3. Lesson changes keep YAML and Markdown references aligned and pass
   `make lessonlint`.
4. Docker or terminal changes preserve exact-container cleanup, resource limits,
   no host mounts, timeouts, and stale asynchronous-message guards.
5. `make fmt`, focused checks, `make check`, and
   `./.beryl/scripts/check.sh` run successfully, or unavailable prerequisites are
   reported exactly.
6. Durable changes update the design tree, architecture, vocabulary, or an ADR;
   test changes update `tests/.manifest.sha256` through the repository script.
