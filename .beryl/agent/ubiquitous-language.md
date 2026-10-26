# Ubiquitous Language

Use these terms consistently in code, tests, lesson content, and design records.

| Domain Term | Technical Symbol | Definition | Constraints | Avoid |
| --- | --- | --- | --- | --- |
| Learner | `student` | The person using Shellforge and the unprivileged account that runs interactive lesson commands. | May use container-local `sudo`; never equate it with a host user or host privilege. | User when specifically discussing lesson identity |
| Lesson | `lessons.Lesson` | One embedded exercise with stable ID, number, title, pages, hints, optional setup, assertions, and a success message. | YAML is version-controlled and validated by `lessonlint`. | Module when referring to one exercise |
| Page | `lessons.Page` | One Markdown instruction unit within a lesson. | File paths are valid embedded `.md` paths and content is hydrated after YAML parsing. | Screen |
| Guide | `lessonState.guide` | The scrollable rendered Markdown area in the lesson's left pane. | Header, disclosures, and help remain outside its scrollable content. | Viewport when domain meaning matters |
| Sandbox | `container.Container` | One disposable Docker container that holds learner-visible filesystem and process state. | No network, host mounts, Docker socket, or privileged mode; exact-session cleanup only. | VM, host shell |
| Terminal session | `terminal.TermSession` | Host PTY/ConPTY, virtual-terminal emulator, sandbox, and asynchronous exit signal as one owned lifecycle. | Closing it stops the process and terminal workers and removes its sandbox. | Container when PTY/emulator ownership is included |
| Terminal generation | `terminalState.gen` | Monotonic identifier for one terminal-start attempt. | Every session-scoped asynchronous message carries it; stale generations are ignored. | Session ID, which can be confused with the Docker name |
| Lesson setup | `Lesson.Setup` | Trusted repository-authored Bash that prepares a fresh sandbox before learner attachment. | Runs as root with strict non-interactive Bash; must establish intended learner ownership. | Learner command |
| Progress check | `checkAssertions`, F12 | One explicit evaluation of every assertion declared by the active lesson. | The bottom status is white before and during evaluation, returns to the F12 prompt while work remains, becomes green when all checks pass, and becomes red when any check cannot run. | Autograde after every command |
| Assertion | `assertion.Assertion` | Declarative observable outcome checked inside the active sandbox. | Validate before shipping; do not require one exact learner solution unless the assertion kind explicitly checks history. | Test, which refers to Go verification code |
| Pending progress check | `Result{Passed: false, Err: nil}` | A learner task that has not yet produced its expected observable outcome. | Render neutrally with `○`; it prevents completion but is not an application error. | Failed assertion in learner-facing text |
| Assertion execution failure | `Result.Err` | Shellforge could not perform a check. | Distinct from a pending check whose expected state is absent; render as a red error. | Failed lesson outcome |
| Completion | `completion.CompletionStore` | Durable record that a lesson ID has passed all declared assertions. | Idempotent, local, and independent of transient current-screen state. | Score, unlock |
| Disclosure | `ui.Disclosure` | Compact F-key-labelled panel that can show or hide hints, progress checks, or help. | `▶` is collapsed and `▼` is expanded; expanded bodies keep their vertical guide. Footer order is hints, progress checks, help. | Modal, popup |
| Standalone sandbox | `config.SandboxTerminalScreen` | Interactive shell launched without a lesson guide or assertions. | F3 and lesson-only shortcuts are forwarded to Bash rather than consumed by app help. | Lesson terminal |
| Completion store | `completion.Store` | SQLite adapter for completed lesson IDs and learner settings under the user's local state directory. | Application behavior degrades to in-memory completion and settings if unavailable. | Progress JSON |
| Cursor style | `config.CursorStyle` | Learner-selected terminal cursor shape and blink combination, persisted by stable ID. | Applies to sandbox and lesson terminals; unknown or missing IDs fall back to a blinking block. | Caret |
| Command history evidence | `.shellforge-history` | Container-local append-only Bash history used by one assertion kind. | Learner-editable educational evidence, not a security boundary. | Audit log |
| Design Tree | `design-tree.md` | Living record of current, open, and settled design choices. | Update before a choice becomes a durable boundary change. | Backlog |
| ADR | `.beryl/agent/adr/*` | Record of a durable architectural decision and its consequences. | Use for boundary, schema, persistence, security, terminology, or test-strategy changes. | Session note |
