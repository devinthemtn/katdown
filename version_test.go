package main

import (
	"os"
	"regexp"
	"testing"
)

func TestVersionMatchesBuildConfig(t *testing.T) {
	b, err := os.ReadFile("build/config.yml")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`(?m)^\s+version:\s*"([^"]+)"`).FindSubmatch(b)
	if m == nil {
		t.Fatal("no info.version in build/config.yml")
	}
	if got := string(m[1]); got != version {
		t.Errorf("build/config.yml version is %q, main.version is %q", got, version)
	}
}
