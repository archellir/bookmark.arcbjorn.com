package web

import "embed"

//go:embed all:dist
var EmbededFilesystem embed.FS
