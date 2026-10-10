# Ubiquitous Language

Use these terms consistently in code, tests, lesson content, and design records.

| Domain Term | Technical Symbol | Definition | Constraints | Avoid |
| --- | --- | --- | --- | --- |
| Learner | `student` | The person using Shellforge and the unprivileged account that runs interactive lesson commands. | May use container-local `sudo`; never equate it with a host user or host privilege. | User when specifically discussing lesson identity |
| Lesson | `lessons.Lesson` | One embedded exercise with stable ID, number, title, pages, hints, optional setup, assertions, and a success message. | YAML is version-controlled and validated by `lessonlint`. | Module when referring to one exercise |
| Page | `lessons.Page` | One Markdown instruction unit within a lesson. | File paths are valid embedded `.md` paths and content is hydrated after YAML parsing. | Screen |
| Guide | `lessonState.guide` | The scrollable rendered Markdown area in the lesson's left pane. | Header, disclosures, and help remain outside its scrollable content. | Viewport when domain meaning matters |
| Sandbox | `container.Container` | One disposable Docker container that holds learner-visible filesystem and process state. | No network, host mounts, Docker socket, or privileged mode; exact-session cleanup only. | VM, host shell |
| Terminal session | `terminal.TermSession` | Bubbleterm emulator, its sandbox, and its asynchronous exit signal as one owned lifecycle. | Closing it stops the emulator and removes its sandbox. | Container when PTY/emulator ownership is included |
| Terminal generation | `terminalState.gen` | Monotonic identifier for one terminal-start attempt. | Every session-scoped asynchronous message carries it; stale generations are ignored. | Session ID, which can be confused with the Docker name |
| Lesson setup | `Lesson.Setup` | Trusted repository-authored Bash that prepares a fresh sandbox before learner attachment. | Runs as root with strict non-interactive Bash; must establish intended learner ownership. | Learner command |
| Progress check | `checkAssertions`, F12 | One explicit evaluation of every assertion declared by the active lesson. | Runs asynchronously with a timeout and returns one result per assertion. | Autograde after every command |
| Assertion | `assertion.Assertion` | Declarative observable outcome checked inside the active sandbox. | Validate before shipping; do not require one exact learner solution unless the assertion kind explicitly checks history. | Test, which refers to Go verification code |
| Assertion execution failure | `Result{Passed: false, Message: "Could not check..."}` | Shellforge could not perform a check. | Distinct from a valid check whose expected state is absent. | Failed lesson outcome |
| Completion | `app.CompletionStore` | Durable record that a lesson ID has passed all declared assertions. | Idempotent, local, and independent of transient current-screen state. | Score, unlock |
| Disclosure | `ui.Disclosure` | Compact F-key-labelled panel that can show or hide hints, assertions, or help. | Empty expanded panels show an intentional instruction; footer order is hints, assertions, help. | Modal, popup |
| Standalone sandbox | `config.SandboxTerminalScreen` | Interactive shell launched without a lesson guide or assertions. | F3 and lesson-only shortcuts are forwarded to Bash rather than consumed by app help. | Lesson terminal |
| Completion store | `completion.Store` | SQLite adapter for completed lesson IDs under the user's local state directory. | Application behavior degrades to in-memory completion if unavailable. | Progress JSON |
| Command history evidence | `.shellforge-history` | Container-local append-only Bash history used by one assertion kind. | Learner-editable educational evidence, not a security boundary. | Audit log |
| Design Tree | `design-tree.md` | Living record of current, open, and settled design choices. | Update before a choice becomes a durable boundary change. | Backlog |
| ADR | `.beryl/agent/adr/*` | Record of a durable architectural decision and its consequences. | Use for boundary, schema, persistence, security, terminology, or test-strategy changes. | Session note |
