# Ubuntu 24.04+ ships webkit2gtk-4.1 only; drop TAGS on distros with 4.0.
TAGS ?= webkit2_41

.PHONY: build dev test generate install

build:
	wails build -tags $(TAGS)

dev:
	wails dev -tags $(TAGS) $(if $(FILE),-appargs "$(FILE)")

test:
	go test ./internal/...

# Regenerates internal/config/chroma.css (syntax-highlighting colors).
generate:
	go generate ./internal/config

install: build
	install -Dm755 build/bin/katdown $(HOME)/.local/bin/katdown
