# Shellforge controls this file so every lesson begins with a predictable,
# readable terminal theme. It is intentionally separate from the student's
# personal Bash configuration.

# Keep the prompt and the text the learner types bright white. \[ and \] mark
# non-printing escape sequences so Bash calculates cursor positions correctly.
# eg. if container id is 123456, this prompt will look like:
# [student@123456 /home/student/workspace]$ command here...
PS1='\[\e[97m\][\u@\h \w]\$ '

# Before Bash runs each command, change the terminal's default foreground to a
# muted gray. Programs that deliberately emit their own colours keep them.
trap 'printf "\e[38;5;241m"' DEBUG
