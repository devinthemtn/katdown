package main

// version is katdown's version. Keep it in step with info.version in
// build/config.yml (TestVersionMatchesBuildConfig checks this).
var version = "1.2.0"

// Version returns katdown's version, for the help overlay.
func (a *App) Version() string {
	return version
}
