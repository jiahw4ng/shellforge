package container

import "errors"

const Image = "shellforge-sandbox:0.1.0"

var ErrUnavailable = errors.New("docker lesson containers are unavailable")
