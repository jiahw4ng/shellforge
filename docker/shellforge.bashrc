# Shellforge controls this file so every lesson begins with a predictable,
# readable terminal theme. It is intentionally separate from the student's
# personal Bash configuration.

# Make the prompt bold but leave its color alone: Shellforge draws text that sets
# no color of its own in the learner's theme, so the prompt and the text they
# type stay readable on both light and dark terminals. \[ and \] mark
# non-printing escape sequences so Bash calculates cursor positions correctly.
# In the default workspace, the prompt looks like: student:~/workspace$
PS1='\[\e[1m\]\u:\w\$\[\e[0m\] '

# Keep lesson-local records of interactive commands, physical working
# directories, and exported environment variables for progress assertions. The
# history and directory records append after each completed command; the
# environment snapshot is replaced with the latest exported state. This is
# learning-progress data, not a security boundary: the student can still
# inspect or change files in their disposable sandbox.
export HISTFILE="$HOME/.shellforge-history"
export HISTSIZE=10000
export HISTFILESIZE=10000
PROMPT_COMMAND='history -a; pwd -P >> "$HOME/.shellforge-working-directories"; env -0 > "$HOME/.shellforge-environment"'
