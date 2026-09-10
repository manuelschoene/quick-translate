//go:build !dev

package main

// Reports whether this binary was built by 'wails dev'. Everything besides a 'wails dev' run, in particular 'wails build' and its own 'production'/'debug' tags, counts as a production build here.
func isDevBuild() bool {
	return false
}
