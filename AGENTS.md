# Agent Instructions

Shellforge is a Go terminal UI for learning Bash in disposable Docker containers.
Repository files are the source of truth; do not rely on chat history when a file
answers the question.

## Project Documentation

Read only what the task needs, from `docs/agent/`:

- `coding-policy.md`: mandatory Go style and engineering rules. Read first.
- `project-brief.md`: product goal, platform contract, definition of done.
- `architecture.md`: package boundaries and dependency rules.
- `design-tree.md`: settled and open design decisions. Record new ones here.
- `ubiquitous-language.md`: shared vocabulary. Use these terms.
- `testing-policy.md`: which checks to run for which change.
- `adr/`: architecture decision records.

## Commands

- `make fmt`: format with golangci-lint.
- `make check`: format check, lint, lesson validation, and all tests (the main gate).
- `make build`: build `bin/shellforge`.
- `make sandbox-image`: build the lesson sandbox image for the host architecture.
- `goreleaser check` and `goreleaser release --snapshot --clean`: validate release packaging without publishing.

## Rules

- Do not implement a feature until the user has ratified a plan. Before coding,
  state the success checks and the commit boundaries (one purpose, files, and the
  validating command each). When there are several viable approaches, present
  them and wait for approval.
- Do not use sub-agents unless the user explicitly asks for them.
- Never weaken a test to make an implementation pass. Explain any intentional
  test change and the behavior change behind it.
- Run `make fmt`, the focused checks from `docs/agent/testing-policy.md`, then
  `make check` after edits. Repair from the actual tool output.
- Keep to the requested change; do not edit unrelated files.
- Before deleting or consolidating root planning or documentation files, list
  every file, what is preserved, what is lost, and where replacement content
  lives, then wait for explicit approval.
- Update the design tree, architecture, glossary, or an ADR when durable
  knowledge changes.
- Do not store secrets in repository files, prompts, tests, or logs.

## Releases

The sandbox image version is `imageVersion` in `internal/container/consts.go`.
Bump it for every release, even when the Dockerfile did not change, and keep
`SANDBOX_IMAGE` in the `Makefile` and the image name in the README in step. A
release tag must be exactly `v` plus that version; both release workflows fail
otherwise. Never reuse or re-point a published tag.

## Completion Report

State what changed, which commit boundary each changed file belongs to (flag any
that belongs to none), the checks run and skipped, and the documentation updated.
