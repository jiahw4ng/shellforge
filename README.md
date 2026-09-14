# shellforge
An interactive, mission-based terminal learning tool written in Go that runs natively inside the user's terminal

## Docker lesson image

Shellforge runs learner commands inside a disposable Docker container. Build the
local lesson image from `docker/lesson.Dockerfile` before launching a learning
session:

```bash
make sandbox-image
```

The MVP requires Docker Desktop with WSL integration enabled for the WSL
distribution running Shellforge.
