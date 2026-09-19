# This Dockerfile creates the sandboxed filesystem used by Shellforge learners.
# It is not the runtime image for the Shellforge application itself.
FROM debian:bookworm-slim@sha256:88200866dfff7ea7f5cbcb6ec7c8a701889efe6fe859fe64d6990e4b07ea4171

# Install the packages needed in student lesson containers.
# TODO: this may be too much for the most basic lesson on file system navigation.
# Maybe we shld create a smaller base image for that lesson.
RUN apt-get update \
    && apt-get install --yes --no-install-recommends \
        bash \
        coreutils \
        findutils \
        gawk \
        grep \
        less \
        nano \
        procps \
        sed \
        sudo \
        util-linux \
    && rm -rf /var/lib/apt/lists/*

# Use a dedicated Bash configuration for the learner terminal. The application
# selects this file explicitly, so it never depends on a student's dotfiles.
COPY docker/shellforge.bashrc /etc/shellforge/bashrc

# Create the learner account, its workspace, and container-only sudo access.

# Create a group called "student" with GID 1000
RUN groupadd --gid 1000 student \
    # Create a user called "student" with UID 1000, add it to the "student" group, create its home directory, and set its shell to /bin/bash
    && useradd \
        --uid 1000 \
        --gid student \
        --create-home \
        --shell /bin/bash \
        student \
    # Create the workspace directory and set its ownership to the "student" user and group
    && mkdir -p /home/student/workspace \
    && chown -R student:student /home/student \
    # Add a sudoers file for the "student" user that allows it to run any command without a password prompt
    && printf 'student ALL=(ALL:ALL) NOPASSWD: ALL\n' \
        > /etc/sudoers.d/student \
    && chmod 0440 /etc/sudoers.d/student \
    && visudo --check --file=/etc/sudoers.d/student

# Set the default user and working directory for the lesson containers
USER student
WORKDIR /home/student/workspace

# The default command is to sleep indefinitely, so that the container can be used interactively
CMD ["sleep", "infinity"]
