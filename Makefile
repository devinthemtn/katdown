# Ubuntu 24.04+ ships webkit2gtk-4.1 only; drop TAGS on distros with 4.0.
TAGS ?= webkit2_41

# Windows install dir, resolved through WSL interop (empty outside WSL).
WIN_INSTALL_DIR ?= $(shell command -v wslpath >/dev/null && wslpath "$$(cmd.exe /c echo %LOCALAPPDATA% 2>/dev/null | tr -d '\r')" 2>/dev/null)/Programs/katdown

.PHONY: build build-windows installer-windows dev test generate install install-windows

build:
	wails build -tags $(TAGS)

# Cross-compiles build/bin/katdown.exe (WebView2 needs no cgo, so no mingw toolchain required).
build-windows:
	wails build -platform windows/amd64

# NSIS installer in build/bin; needs makensis.
installer-windows:
	wails build -platform windows/amd64 -nsis

dev:
	wails dev -tags $(TAGS) $(if $(FILE),-appargs "$(FILE)")

test:
	go test ./internal/...

# Regenerates internal/config/chroma.css (syntax-highlighting colors).
generate:
	go generate ./internal/config

install: build
	install -Dm755 build/bin/katdown $(HOME)/.local/bin/katdown

# From WSL: copies katdown.exe to %LOCALAPPDATA%\Programs\katdown, then adds a
# Start Menu shortcut and registers it for .md files (build/windows/install.ps1).
install-windows: build-windows
	@test "$(WIN_INSTALL_DIR)" != "/Programs/katdown" || { echo "not in WSL; set WIN_INSTALL_DIR"; exit 1; }
	install -Dm755 build/bin/katdown.exe "$(WIN_INSTALL_DIR)/katdown.exe"
	powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$$(wslpath -w build/windows/install.ps1)" -Exe "$$(wslpath -w "$(WIN_INSTALL_DIR)/katdown.exe")"
