//go:build production

package embed

import (
	"ekken/ui"
	"embed"
)

func hasEmbeddedUI() bool     { return true }
func getEmbeddedUI() embed.FS { return ui.FS }
