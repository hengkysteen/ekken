//go:build !production

package embed

import "embed"

func hasEmbeddedUI() bool     { return false }
func getEmbeddedUI() embed.FS { return embed.FS{} }
