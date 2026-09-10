//go:build dev

package main

// Reports whether this binary was built by 'wails dev', which compiles with the 'dev' build tag for live reload.
func isDevBuild() bool {
	return true
}
