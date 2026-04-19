// Package web embeds the dashboard static files so the server binary is
// self-contained and does not depend on the working directory at runtime.
package web

import "embed"

//go:embed static
var Static embed.FS
