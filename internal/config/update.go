package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

var (
	tableHeader     = regexp.MustCompile(`^\s*\[\s*([A-Za-z0-9_.-]+)\s*\]\s*(#.*)?$`)
	keyLine         = regexp.MustCompile(`^\s*([A-Za-z0-9_-]+)\s*=`)
	commentedHeader = regexp.MustCompile(`^\s*#\s*\[\s*([A-Za-z0-9_.-]+)\s*\]`)
)

// defaultEntry is one setting in DefaultSettingsFile: the comment lines
// directly above it followed by its key = value line.
type defaultEntry struct {
	table, key string
	lines      []string
}

// defaultTable is a whole [table] section of DefaultSettingsFile.
type defaultTable struct {
	name  string
	lines []string
}

func parseDefaultFile() ([]defaultEntry, map[string]*defaultTable) {
	var entries []defaultEntry
	tables := map[string]*defaultTable{}
	table := ""
	var comments []string
	for _, line := range strings.Split(DefaultSettingsFile, "\n") {
		trimmed := strings.TrimSpace(line)
		if m := tableHeader.FindStringSubmatch(line); m != nil {
			table = m[1]
			tables[table] = &defaultTable{name: table, lines: []string{line}}
			comments = nil
			continue
		}
		if table != "" {
			tables[table].lines = append(tables[table].lines, line)
		}
		switch {
		case trimmed == "":
			comments = nil
		case strings.HasPrefix(trimmed, "#"):
			comments = append(comments, line)
		default:
			if m := keyLine.FindStringSubmatch(line); m != nil {
				entries = append(entries, defaultEntry{table, m[1], append(append([]string(nil), comments...), line)})
			}
			comments = nil
		}
	}
	for _, t := range tables {
		for len(t.lines) > 0 && strings.TrimSpace(t.lines[len(t.lines)-1]) == "" {
			t.lines = t.lines[:len(t.lines)-1]
		}
	}
	return entries, tables
}

// commentOut turns setting lines into comments, leaving comments and blank
// lines as they are.
func commentOut(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		t := strings.TrimSpace(l)
		if t == "" || strings.HasPrefix(t, "#") {
			out[i] = l
		} else {
			out[i] = "# " + l
		}
	}
	return out
}

// hasCommentedKey reports whether lines contain "# key = ...".
func hasCommentedKey(lines []string, key string) bool {
	re := regexp.MustCompile(`^\s*#\s*` + regexp.QuoteMeta(key) + `\s*=`)
	for _, l := range lines {
		if re.MatchString(l) {
			return true
		}
	}
	return false
}

// UpdateSettingsText adds every setting in DefaultSettingsFile that the
// settings file data neither sets nor already mentions in a comment, as a
// commented-out line with its explanation, so it is easy to turn on. A key
// goes inside its table when the file has that table, and a missing table is
// appended whole, with its [table] header left on. added lists the new settings as "table.key"; out is data
// unchanged when there is nothing to add.
func UpdateSettingsText(data string) (out string, added []string, err error) {
	var probe map[string]any
	meta, err := toml.Decode(data, &probe)
	if err != nil {
		return data, nil, fmt.Errorf("%s: %w", SettingsFile, err)
	}

	lines := strings.Split(strings.TrimRight(data, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	// Where each [table] starts in the user's file, and where its end is.
	firstHeader := len(lines)
	starts := map[string]int{}
	var order []int
	for i, l := range lines {
		if m := tableHeader.FindStringSubmatch(l); m != nil {
			if _, dup := starts[m[1]]; !dup {
				starts[m[1]] = i
			}
			order = append(order, i)
			if i < firstHeader {
				firstHeader = i
			}
		}
	}
	regionEnd := func(start int) int {
		next := len(lines)
		for _, h := range order {
			if h > start {
				next = h
				break
			}
		}
		k := next - 1
		// Comments directly above the next header belong to it.
		if next < len(lines) {
			for k > start && strings.HasPrefix(strings.TrimSpace(lines[k]), "#") {
				k--
			}
		}
		for k > start && strings.TrimSpace(lines[k]) == "" {
			k--
		}
		return k + 1
	}
	commentedTables := map[string]bool{}
	for _, l := range lines {
		if m := commentedHeader.FindStringSubmatch(l); m != nil {
			commentedTables[m[1]] = true
		}
	}

	inserts := map[int][]string{}
	var tail []string
	entries, tables := parseDefaultFile()
	doneTables := map[string]bool{}
	for _, e := range entries {
		name := e.key
		if e.table != "" {
			name = e.table + "." + e.key
		}
		if (e.table == "" && meta.IsDefined(e.key)) || (e.table != "" && meta.IsDefined(e.table, e.key)) {
			continue
		}

		switch start, ok := starts[e.table]; {
		case e.table == "":
			if hasCommentedKey(lines[:firstHeader], e.key) {
				continue
			}
			block := append(commentOut(e.lines), "")
			inserts[firstHeader] = append(inserts[firstHeader], block...)
		case ok:
			end := regionEnd(start)
			if hasCommentedKey(lines[start:end], e.key) {
				continue
			}
			block := commentOut(e.lines)
			// A key without its own comment stays with the one added before it.
			if len(e.lines) > 1 || len(inserts[end]) == 0 {
				block = append([]string{""}, block...)
			}
			inserts[end] = append(inserts[end], block...)
		case meta.IsDefined(e.table):
			// The table is set with dotted keys or inline, not a [table]
			// header, so add a dotted key at the top level.
			if hasCommentedKey(lines[:firstHeader], name) {
				continue
			}
			l := e.lines[len(e.lines)-1]
			l = strings.Replace(l, e.key, name, 1)
			block := append(commentOut(append(e.lines[:len(e.lines)-1:len(e.lines)-1], l)), "")
			inserts[firstHeader] = append(inserts[firstHeader], block...)
		default:
			if commentedTables[e.table] {
				continue // added by an earlier update
			}
			if !doneTables[e.table] {
				doneTables[e.table] = true
				// The [table] header stays on, so uncommenting one setting
				// under it is enough; only the settings are commented out.
				// An empty table changes nothing.
				t := tables[e.table].lines
				tail = append(tail, "", t[0])
				tail = append(tail, commentOut(t[1:])...)
			}
		}
		added = append(added, name)
	}
	if len(added) == 0 {
		return data, nil, nil
	}

	var b strings.Builder
	for i := 0; i <= len(lines); i++ {
		for _, l := range inserts[i] {
			b.WriteString(l + "\n")
		}
		if i < len(lines) {
			b.WriteString(lines[i] + "\n")
		}
	}
	for _, l := range tail {
		b.WriteString(l + "\n")
	}
	return strings.TrimLeft(b.String(), "\n"), added, nil
}

// UpdateSettings adds the settings config.toml is missing to it as comments;
// see UpdateSettingsText. A missing config.toml is created as by MakeSettings,
// with created true.
func UpdateSettings() (path string, added []string, created bool, err error) {
	path, err = SettingsPath()
	if err != nil {
		return "", nil, false, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		_, created, err = MakeSettings()
		return path, nil, created, err
	}
	if err != nil {
		return path, nil, false, err
	}
	out, added, err := UpdateSettingsText(string(data))
	if err != nil || len(added) == 0 {
		return path, nil, false, err
	}
	return path, added, false, replaceFile(path, out)
}

// replaceFile writes data to path through a temporary file, so a failed
// write leaves the old file intact. The file keeps its permissions.
func replaceFile(path, data string) error {
	mode := fs.FileMode(0o644)
	if fi, err := os.Stat(path); err == nil {
		mode = fi.Mode().Perm()
	}
	f, err := os.CreateTemp(filepath.Dir(path), SettingsFile+".*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Chmod(mode); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
