# Agent Operating Instructions

This is a generated tool shim. Do not edit it directly; update `.beryl/agent/tool-instruction-template.md` and run `.beryl/agent/scripts/sync-agent-env.sh`.

Repository files are the source of truth. Before changing code, tests, documentation, configuration, or Git history, read `.beryl/agent/coding-policy.md` first. Then load only the smallest relevant project-context files from `.beryl/agent/`.

## Workflow Selection

Choose one matching workflow from `.beryl/agent/skills/`:

* `planning` for plans and unratified feature requests.
* `adding-features` for approved implementation.
* `debugging` for failures and regressions.
* `explaining-codebase` for read-only explanations.
* `grill-me` for risky, ambiguous, cross-context, or security-sensitive design work.

The coding policy is the primary contract. Follow the selected workflow only where it does not conflict with that policy.

## Always-On Rules

* Do not implement a feature without a user-ratified plan.
* Do not use sub-agents unless the user explicitly requests them.
* Never weaken tests to make implementation pass. If tests change intentionally, run `./.beryl/scripts/update-test-manifest.sh`.
* For method Javadocs, start the first summary sentence with a third-person verb such as `Returns`, `Sends`, or `Adds`; follow the full formatting rules in `.beryl/agent/coding-policy.md`.
* Use `.beryl/agent/session-state.md` only for temporary state; clear it when the task ends.
* Run the formatter if configured, focused checks, and `./.beryl/scripts/check.sh` after edits.
* Update the design tree, architecture, vocabulary, or ADRs when durable knowledge changes.
* Do not store secrets in repository files, prompts, tests, or logs.

## Completion

Report what changed, each changed file's commit boundary, checks run or skipped, design updates, test-manifest changes, and whether temporary state was cleared.
