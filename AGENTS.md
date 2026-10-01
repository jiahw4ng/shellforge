# AGENTS.md

## 💻 Tech Stack & Architecture
- Go 1.26.5 TUI using Bubble Tea v2's Model/Update/View pattern, Bubbles, Lip Gloss v2, and Glamour.
- `cmd/shellforge` is the entrypoint; `internal/app` owns state and orchestration, while `screens`, `terminal`, `container`, `assertion`, and `completion` isolate UI and infrastructure concerns.
- Lessons are YAML plus Markdown embedded from `lessons/`; progress is stored locally with `modernc.org/sqlite`, and each shell runs in a disposable Docker container.

## 🛠️ Operational Commands
- Development: `go run ./cmd/shellforge` (run Docker and build its image once with `make sandbox-image`).
- Build & Compile: `make build` (writes `bin/shellforge`); use `make sandbox-image` for the Linux lesson image.
- Run Tests: `make test`
- Lint & Format: `make fmt` applies formatting; `make check` is the full CI gate.

## 🎨 Code Style & Preferences
- Let `golangci-lint` v2 apply `gofumpt` and `goimports`; follow existing package boundaries and wrapped-error style.
- Keep `State.Update` a thin event dispatcher. Put concrete message handlers in `internal/app/handlemessages.go`, screen key routing in `keypress.go`, and blocking work in `tea.Cmd` functions.
- Use value receivers for handlers returning the next model and pointer receivers for mutation-only helpers; preserve terminal generation checks and explicit cleanup.
- When editing a lesson, keep its YAML and referenced Markdown aligned and run `make lessonlint`.

## 🛑 Strict Guardrails & Restrictions
- Always use Go modules and the Makefile. Never introduce npm, pnpm, yarn, or their lockfiles.
- Do not edit or commit `bin/`; it is generated. Tests must use temporary paths, never the real `~/.local/state/shellforge` data.
- Do not remove stale-message guards, Docker cleanup, timeouts, or unprivileged-container boundaries without proving the lifecycle remains safe.
- Do not install new packages or add Go dependencies without human consent.
- Always run linting and testing commands to verify your work before declaring a task complete.
