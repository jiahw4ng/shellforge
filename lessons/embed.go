// Package lessondata embeds Shellforge's version-controlled lesson files.
package lessondata

import "embed"

// Files contains every lesson YAML file distributed with Shellforge.
//
//go:embed *.yaml
var Files embed.FS

// go: embed *.yaml means
// go will take every YAML file in this lessons/ directory and package it inside the compiled Shellforge binary
