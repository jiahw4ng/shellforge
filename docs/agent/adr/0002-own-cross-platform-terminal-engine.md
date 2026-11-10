# ADR 0002: Own The Cross-Platform Terminal Engine

## Status

Accepted

## Context

Shellforge previously delegated PTY process ownership and terminal emulation to
Bubbleterm. Bubbleterm configures Unix-only process fields during compilation,
which prevents producing a native Windows executable. A Windows-only adapter
would leave Linux and Windows with different input, rendering, and lifecycle
behavior.

## Decision

Own one terminal engine inside `internal/terminal` on every supported host.
Use `github.com/charmbracelet/x/xpty` for Unix PTY and Windows ConPTY process
attachment, and `github.com/charmbracelet/x/vt` for input encoding, ANSI state,
cursor state, and rendering. Keep `terminal.Start` and `terminal.TermSession` as
the application-facing boundary.

Terminal output notifications carry terminal identity so a replacement session
ignores stale messages. Shutdown kills the owned host-side process, wakes the
virtual-terminal response reader, closes the PTY, waits for I/O workers, closes
the emulator, waits for process reaping, and removes exactly the owned sandbox.

## Consequences

- **Benefit:** Linux, WSL, and native Windows share one terminal behavior and
  lifecycle implementation.
- **Benefit:** Native Windows builds use ConPTY without changing the Linux Bash
  lesson environment.
- **Tradeoff:** Shellforge now owns key forwarding, worker synchronization, and
  terminal rendering integration previously supplied by Bubbleterm.
- **Tradeoff:** `xpty` and `x/vt` are pre-v1 dependencies and must remain pinned
  behind the terminal package boundary.
- **Follow-up:** Treat macOS as build-compatible until real Docker Desktop
  runtime evidence supports an advertised platform claim.
