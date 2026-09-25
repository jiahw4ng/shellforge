# Shellforge controls this file so every lesson begins with a predictable,
# readable terminal theme. It is intentionally separate from the student's
# personal Bash configuration.

# Keep the prompt and the text the learner types bright white. \[ and \] mark
# non-printing escape sequences so Bash calculates cursor positions correctly.
# eg. if container id is 123456, this prompt will look like:
# [student@123456 /home/student/workspace]$ command here...
PS1='\[\e[97m\][\u@\h \w]\$ '

# Keep a lesson-local record of interactive commands for progress assertions.
# history -a appends after each completed command, rather than waiting until
# Bash exits. This is learning-progress data, not a security boundary: the
# student can still inspect or change files in their disposable sandbox.
export HISTFILE="$HOME/.shellforge-history"
export HISTSIZE=10000
export HISTFILESIZE=10000
PROMPT_COMMAND='history -a'

# Before Bash runs each command, change the terminal's default foreground to a
# muted gray. Programs that deliberately emit their own colours keep them.
trap 'printf "\e[38;5;241m"' DEBUG
