// Package lessondata embeds Shellforge's version-controlled lesson files.
package lessondata

import "embed"

// Files contains every lesson YAML definition and Markdown page distributed
// with Shellforge.
//
//go:embed *.yaml */*.md
var Files embed.FS
