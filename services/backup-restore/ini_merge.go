package backuprestore

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// ErrManualINIMerge means an INI input cannot be merged without user review.
var ErrManualINIMerge = errors.New("INI file requires manual merge")

// INIConflict identifies one key whose snapshot and current values differ.
// An empty Section identifies a key before the first section header.
type INIConflict struct {
	Section       string
	Key           string
	SnapshotValue string
	CurrentValue  string
}

// INIMergeResult contains a candidate with current values retained at conflicts.
// Callers must resolve every Conflict before applying Content to a live file.
type INIMergeResult struct {
	Content   string
	Conflicts []INIConflict
}

// iniEntry retains a parsed value and its source line for snapshot-only insertion.
type iniEntry struct {
	value string
	line  string
}

// iniSection indexes unique keys and where additions belong in the current file.
type iniSection struct {
	name     string
	keys     map[string]iniEntry
	order    []string
	insertAt int
}

// iniDocument preserves source lines alongside the parsed section index.
type iniDocument struct {
	lines    []string
	sections []*iniSection
	byName   map[string]*iniSection
}

// MergeINI combines snapshot and current INI content by section and key.
// It preserves current lines, appends snapshot-only keys, and reports differing
// values as conflicts. Unsupported or ambiguous syntax requires manual merge.
func MergeINI(snapshot string, current string) (INIMergeResult, error) {
	stored, err := parseINI(snapshot)
	if err != nil {
		return INIMergeResult{}, fmt.Errorf("snapshot: %w", err)
	}
	live, err := parseINI(current)
	if err != nil {
		return INIMergeResult{}, fmt.Errorf("current: %w", err)
	}

	insertions := make(map[int][]string)
	result := INIMergeResult{}
	lineEnding := preferredLineEnding(current)
	for _, storedSection := range stored.sections {
		liveSection := live.byName[storedSection.name]
		if liveSection == nil {
			appendMissingSection(&result, storedSection, lineEnding)
			continue
		}
		for _, key := range storedSection.order {
			storedEntry := storedSection.keys[key]
			liveEntry, exists := liveSection.keys[key]
			if !exists {
				insertions[liveSection.insertAt] = append(insertions[liveSection.insertAt], storedEntry.line)
				continue
			}
			if storedEntry.value != liveEntry.value {
				result.Conflicts = append(result.Conflicts, INIConflict{
					Section: storedSection.name, Key: key,
					SnapshotValue: storedEntry.value, CurrentValue: liveEntry.value,
				})
			}
		}
	}
	result.Content = renderINIMerge(live.lines, insertions, result.Content, lineEnding)
	return result, nil
}

// parseINI accepts the simple section and key syntax that can be round-tripped.
// Duplicate sections or keys and unsupported syntax go to manual resolution.
func parseINI(content string) (iniDocument, error) {
	if !utf8.ValidString(content) || strings.ContainsRune(content, 0) {
		return iniDocument{}, fmt.Errorf("%w: invalid text", ErrManualINIMerge)
	}
	doc := iniDocument{lines: splitLines(content), byName: make(map[string]*iniSection)}
	section := &iniSection{name: "", keys: make(map[string]iniEntry)}
	doc.sections = append(doc.sections, section)
	doc.byName[""] = section
	for index, line := range doc.lines {
		body := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		trimmed := strings.TrimSpace(body)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}
		if strings.HasPrefix(trimmed, "[") {
			name, err := parseSectionName(trimmed)
			if err != nil {
				return iniDocument{}, err
			}
			if _, exists := doc.byName[name]; exists {
				return iniDocument{}, fmt.Errorf("%w: duplicate section %q", ErrManualINIMerge, name)
			}
			section.insertAt = index
			section = &iniSection{name: name, keys: make(map[string]iniEntry)}
			doc.sections = append(doc.sections, section)
			doc.byName[name] = section
			continue
		}
		key, value, err := parseINIKey(trimmed)
		if err != nil {
			return iniDocument{}, err
		}
		if _, exists := section.keys[key]; exists {
			return iniDocument{}, fmt.Errorf("%w: duplicate key %q in section %q", ErrManualINIMerge, key, section.name)
		}
		section.keys[key] = iniEntry{value: value, line: body}
		section.order = append(section.order, key)
	}
	section.insertAt = len(doc.lines)
	return doc, nil
}

// parseSectionName accepts one unambiguous section header.
func parseSectionName(line string) (string, error) {
	if !strings.HasSuffix(line, "]") || strings.Count(line, "[") != 1 || strings.Count(line, "]") != 1 {
		return "", fmt.Errorf("%w: unsupported section header", ErrManualINIMerge)
	}
	name := strings.TrimSpace(line[1 : len(line)-1])
	if name == "" {
		return "", fmt.Errorf("%w: empty section name", ErrManualINIMerge)
	}
	return name, nil
}

// parseINIKey accepts an ordinary key=value entry and compares trimmed values.
func parseINIKey(line string) (string, string, error) {
	separator := strings.IndexByte(line, '=')
	if separator < 1 {
		return "", "", fmt.Errorf("%w: unsupported key entry", ErrManualINIMerge)
	}
	key := strings.TrimSpace(line[:separator])
	if key == "" || strings.ContainsAny(key, "[]") {
		return "", "", fmt.Errorf("%w: invalid key", ErrManualINIMerge)
	}
	return key, strings.TrimSpace(line[separator+1:]), nil
}

// splitLines keeps each existing line ending so current content survives intact.
func splitLines(content string) []string {
	if content == "" {
		return nil
	}
	lines := strings.SplitAfter(content, "\n")
	if lines[len(lines)-1] == "" {
		return lines[:len(lines)-1]
	}
	return lines
}

// preferredLineEnding uses the first current newline for inserted entries.
func preferredLineEnding(current string) string {
	if index := strings.IndexByte(current, '\n'); index > 0 && current[index-1] == '\r' {
		return "\r\n"
	}
	return "\n"
}

// appendMissingSection stages one snapshot-only section after current content.
func appendMissingSection(result *INIMergeResult, section *iniSection, lineEnding string) {
	if section.name != "" {
		result.Content += "[" + section.name + "]" + lineEnding
	}
	for _, key := range section.order {
		result.Content += section.keys[key].line + lineEnding
	}
}

// renderINIMerge adds snapshot-only entries at their current section boundary.
func renderINIMerge(lines []string, insertions map[int][]string, missingSections string, lineEnding string) string {
	var output strings.Builder
	for index := 0; index <= len(lines); index++ {
		if entries := insertions[index]; len(entries) > 0 {
			if output.Len() > 0 && !strings.HasSuffix(output.String(), "\n") {
				output.WriteString(lineEnding)
			}
			for _, entry := range entries {
				output.WriteString(entry)
				output.WriteString(lineEnding)
			}
		}
		if index < len(lines) {
			output.WriteString(lines[index])
		}
	}
	if missingSections != "" {
		if output.Len() > 0 && !strings.HasSuffix(output.String(), "\n") {
			output.WriteString(lineEnding)
		}
		output.WriteString(missingSections)
	}
	return output.String()
}
