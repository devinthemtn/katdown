package main

// version is katdown's version. Keep it in step with info.version in
// build/config.yml (TestVersionMatchesBuildConfig checks this). Builds made
// through task overwrite it with `git describe` (see VERSION in Taskfile.yml),
// so a build between tags shows e.g. 1.2.0-15-gea54352-dirty.
var version = "1.2.0"

// Version returns katdown's version, for the help overlay.
func (a *App) Version() string {
	return version
}
