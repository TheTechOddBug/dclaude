package extensions

import "embed"

// FS embeds all extension files from subdirectories
// Each extension is a directory containing config.yaml and optional scripts
//
//go:embed */*
var FS embed.FS
