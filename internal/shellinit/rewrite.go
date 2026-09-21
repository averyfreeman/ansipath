package shellinit

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	beginMarker = "# >>> ansipath shell helpers >>>"
	endMarker   = "# <<< ansipath shell helpers <<<"
)

// Options describes the installed paths and environment used while rewriting.
type Options struct {
	DataDir     string
	BinaryPath  string
	Environment map[string]string
}

// Change records one rewritten PATH assignment.
type Change struct {
	Line  int
	Kind  string
	Paths []string
}

type CleanOptions struct {
	Options
	Exists func(string) bool
}

type CleanChange struct {
	Line   int
	Reason string
	Path   string
}

type sourceLine struct {
	text   string
	ending string
}

type pathSegment struct {
	raw    string
	value  string
	anchor bool
}

var assignmentPattern = regexp.MustCompile(`^([[:space:]]*)(export[[:space:]]+)?PATH[[:space:]]*=[[:space:]]*(.*)$`)

// Rewrite replaces PATH assignments and ensures the helper loader is first.
func Rewrite(content string, options Options) (string, []Change, error) {
	lines := splitLines(content)
	lines, err := removeManagedBlock(lines)
	if err != nil {
		return "", nil, err
	}

	environment := options.Environment
	if environment == nil {
		environment = environmentMap(os.Environ())
	}
	virtualPath := splitPath(environment["PATH"])
	changes := make([]Change, 0)
	rewritten := make([]sourceLine, 0, len(lines))

	for index, line := range lines {
		match := assignmentPattern.FindStringSubmatch(line.text)
		if match == nil || strings.HasPrefix(strings.TrimSpace(line.text), "#") {
			rewritten = append(rewritten, line)
			continue
		}

		segments, err := parseAssignmentValue(match[3])
		if err != nil {
			return "", nil, fmt.Errorf("parse PATH assignment on line %d: %w", index+1, err)
		}
		replacement, resultPath, paths, err := rewriteAssignment(match[1], segments, virtualPath, environment)
		if err != nil {
			return "", nil, fmt.Errorf("rewrite PATH assignment on line %d: %w", index+1, err)
		}
		virtualPath = resultPath
		for _, replacementLine := range replacement {
			rewritten = append(rewritten, sourceLine{text: replacementLine, ending: ""})
		}
		changes = append(changes, Change{Line: index + 1, Kind: "path assignment", Paths: paths})
	}

	eol := detectLineEnding(lines)
	if eol == "" {
		eol = "\n"
	}
	for index := range rewritten {
		if rewritten[index].ending == "" {
			rewritten[index].ending = eol
		}
	}

	blockLines := splitLines(ManagedBlock(options))
	for index := range blockLines {
		blockLines[index].ending = eol
	}
	insertAt := 0
	if len(rewritten) > 0 && strings.HasPrefix(rewritten[0].text, "#!") {
		insertAt = 1
	}
	inserted := make([]sourceLine, 0, len(rewritten)+len(blockLines)+1)
	inserted = append(inserted, rewritten[:insertAt]...)
	inserted = append(inserted, blockLines...)
	if len(rewritten) > insertAt && strings.TrimSpace(rewritten[insertAt].text) != "" {
		inserted = append(inserted, sourceLine{text: "", ending: eol})
	}
	inserted = append(inserted, rewritten[insertAt:]...)

	return renderLines(inserted), changes, nil
}

// Clean removes duplicate and dead literal PATH entries from a profile.
func Clean(content string, options CleanOptions) (string, []CleanChange, error) {
	lines := splitLines(content)
	withoutBlock, err := removeManagedBlock(lines)
	if err != nil {
		return "", nil, err
	}
	environment := options.Environment
	if environment == nil {
		environment = environmentMap(os.Environ())
	}
	exists := options.Exists
	if exists == nil {
		exists = func(path string) bool {
			_, err := os.Stat(path)
			return err == nil
		}
	}
	virtualPath := splitPath(environment["PATH"])
	seen := make(map[string]struct{})
	changes := make([]CleanChange, 0)
	rewritten := make([]sourceLine, 0, len(withoutBlock))

	for index, line := range withoutBlock {
		match := assignmentPattern.FindStringSubmatch(line.text)
		if match == nil || strings.HasPrefix(strings.TrimSpace(line.text), "#") {
			rewritten = append(rewritten, line)
			continue
		}
		rawSegments, err := parseAssignmentValue(match[3])
		if err != nil {
			return "", nil, fmt.Errorf("parse PATH assignment on line %d: %w", index+1, err)
		}
		anchorIndex := -1
		for segmentIndex, raw := range rawSegments {
			if isPATHAnchor(strings.TrimSpace(raw)) {
				if anchorIndex >= 0 {
					return "", nil, fmt.Errorf("multiple PATH anchors are not supported on line %d", index+1)
				}
				anchorIndex = segmentIndex
			}
		}
		assignmentSeen := seen
		if anchorIndex < 0 {
			assignmentSeen = make(map[string]struct{})
		}
		before := make([]string, 0)
		after := make([]string, 0)
		literalIndex := 0
		removed := false
		for _, raw := range rawSegments {
			trimmed := strings.TrimSpace(raw)
			if isPATHAnchor(trimmed) {
				literalIndex++
				continue
			}
			value, err := expandSegment(trimmed, environment)
			if err != nil {
				return "", nil, fmt.Errorf("expand PATH assignment on line %d: %w", index+1, err)
			}
			if value == "" {
				literalIndex++
				continue
			}
			reason := ""
			if _, duplicate := assignmentSeen[value]; duplicate {
				reason = "duplicate"
			} else if !exists(value) {
				reason = "dead"
			}
			if reason != "" {
				changes = append(changes, CleanChange{Line: index + 1, Reason: reason, Path: value})
				removed = true
				literalIndex++
				continue
			}
			assignmentSeen[value] = struct{}{}
			if anchorIndex >= 0 && literalIndex < anchorIndex {
				before = append(before, value)
			} else {
				after = append(after, value)
			}
			literalIndex++
		}
		if anchorIndex < 0 {
			virtualPath = append([]string(nil), after...)
			seen = assignmentSeen
		} else {
			virtualPath = append(append(append([]string(nil), before...), virtualPath...), after...)
			seen = assignmentSeen
		}
		if !removed {
			rewritten = append(rewritten, line)
			continue
		}

		if anchorIndex < 0 {
			rewritten = append(rewritten, sourceLine{text: match[1] + "PATH=; export PATH"})
			for _, value := range after {
				rewritten = append(rewritten, sourceLine{text: match[1] + "pathmunge append " + shellQuote(value)})
			}
		} else {
			for valueIndex := len(before) - 1; valueIndex >= 0; valueIndex-- {
				rewritten = append(rewritten, sourceLine{text: match[1] + "pathmunge prepend " + shellQuote(before[valueIndex])})
			}
			for _, value := range after {
				rewritten = append(rewritten, sourceLine{text: match[1] + "pathmunge append " + shellQuote(value)})
			}
		}
	}

	if len(changes) == 0 {
		return content, nil, nil
	}
	eol := detectLineEnding(lines)
	if eol == "" {
		eol = "\n"
	}
	for index := range rewritten {
		if rewritten[index].ending == "" {
			rewritten[index].ending = eol
		}
	}
	blockLines := splitLines(ManagedBlock(options.Options))
	for index := range blockLines {
		blockLines[index].ending = eol
	}
	insertAt := 0
	if len(rewritten) > 0 && strings.HasPrefix(rewritten[0].text, "#!") {
		insertAt = 1
	}
	inserted := make([]sourceLine, 0, len(rewritten)+len(blockLines)+1)
	inserted = append(inserted, rewritten[:insertAt]...)
	inserted = append(inserted, blockLines...)
	if len(rewritten) > insertAt && strings.TrimSpace(rewritten[insertAt].text) != "" {
		inserted = append(inserted, sourceLine{text: "", ending: eol})
	}
	inserted = append(inserted, rewritten[insertAt:]...)
	return renderLines(inserted), changes, nil
}

// CleanPathValue returns a cleaned PATH value and its audit changes.
func CleanPathValue(value string, exists func(string) bool) (string, []CleanChange) {
	if exists == nil {
		exists = func(path string) bool {
			_, err := os.Stat(path)
			return err == nil
		}
	}
	seen := make(map[string]struct{})
	cleaned := make([]string, 0)
	changes := make([]CleanChange, 0)
	for _, path := range splitPath(value) {
		if _, duplicate := seen[path]; duplicate {
			changes = append(changes, CleanChange{Reason: "duplicate", Path: path})
			continue
		}
		if !exists(path) {
			changes = append(changes, CleanChange{Reason: "dead", Path: path})
			continue
		}
		seen[path] = struct{}{}
		cleaned = append(cleaned, path)
	}
	return strings.Join(cleaned, ":"), changes
}

// ManagedBlock returns the sourceable helper loader for an installed binary.
func ManagedBlock(options Options) string {
	return strings.Join([]string{
		beginMarker,
		"ANSIPATH_BIN=" + shellQuote(options.BinaryPath),
		"export ANSIPATH_BIN",
		"if [ -r " + shellQuote(options.DataDir+"/whichfunc.sh") + " ]; then",
		"    . " + shellQuote(options.DataDir+"/whichfunc.sh"),
		"fi",
		"if [ -r " + shellQuote(options.DataDir+"/pathmunge.sh") + " ]; then",
		"    . " + shellQuote(options.DataDir+"/pathmunge.sh"),
		"fi",
		endMarker,
	}, "\n") + "\n"
}

func rewriteAssignment(indent string, rawSegments []string, current []string, environment map[string]string) ([]string, []string, []string, error) {
	segments := make([]pathSegment, 0, len(rawSegments))
	anchorIndex := -1
	for _, raw := range rawSegments {
		trimmed := strings.TrimSpace(raw)
		if isPATHAnchor(trimmed) {
			if anchorIndex >= 0 {
				return nil, nil, nil, fmt.Errorf("multiple PATH anchors are not supported")
			}
			anchorIndex = len(segments)
			segments = append(segments, pathSegment{raw: trimmed, anchor: true})
			continue
		}
		value, err := expandSegment(trimmed, environment)
		if err != nil {
			return nil, nil, nil, err
		}
		if value == "" {
			continue
		}
		segments = append(segments, pathSegment{raw: trimmed, value: value})
	}

	result := make([]string, 0)
	replacement := make([]string, 0)
	paths := make([]string, 0)
	if anchorIndex < 0 {
		replacement = append(replacement, indent+"PATH=; export PATH")
		for _, segment := range segments {
			result = append(result, segment.value)
			paths = append(paths, segment.value)
			replacement = append(replacement, indent+"pathmunge append "+shellQuote(segment.value))
		}
		return replacement, result, paths, nil
	}

	before := make([]string, 0, anchorIndex)
	for _, segment := range segments[:anchorIndex] {
		before = append(before, segment.value)
	}
	after := make([]string, 0, len(segments)-anchorIndex-1)
	for _, segment := range segments[anchorIndex+1:] {
		after = append(after, segment.value)
	}
	result = append(result, before...)
	result = append(result, current...)
	result = append(result, after...)

	for index := len(before) - 1; index >= 0; index-- {
		paths = append(paths, before[index])
		replacement = append(replacement, indent+"pathmunge prepend "+shellQuote(before[index]))
	}
	for _, value := range after {
		paths = append(paths, value)
		replacement = append(replacement, indent+"pathmunge append "+shellQuote(value))
	}
	return replacement, result, paths, nil
}

func parseAssignmentValue(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	if value[0] == '"' || value[0] == '\'' {
		quote := value[0]
		end := -1
		for index := 1; index < len(value); index++ {
			if value[index] == quote && value[index-1] != '\\' {
				end = index
				break
			}
		}
		if end < 0 {
			return nil, fmt.Errorf("unterminated quote")
		}
		trailing := strings.TrimSpace(value[end+1:])
		if trailing != "" && !strings.HasPrefix(trailing, "#") {
			return nil, fmt.Errorf("unexpected text after quoted value")
		}
		value = value[1:end]
	} else if fields := strings.Fields(value); len(fields) > 0 {
		value = fields[0]
	}
	return strings.Split(value, ":"), nil
}

func expandSegment(segment string, environment map[string]string) (string, error) {
	if strings.Contains(segment, "$(") || strings.Contains(segment, "`") {
		return "", fmt.Errorf("command substitution in PATH segment is unsupported")
	}
	if home := environment["HOME"]; home != "" && strings.HasPrefix(segment, "~/") {
		segment = home + segment[1:]
	}
	return os.Expand(segment, func(key string) string {
		return environment[key]
	}), nil
}

func isPATHAnchor(segment string) bool {
	return segment == "$PATH" || segment == "${PATH}" || segment == "PATH"
}

func splitPath(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ":")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func environmentMap(values []string) map[string]string {
	result := make(map[string]string, len(values))
	for _, value := range values {
		key, raw, ok := strings.Cut(value, "=")
		if ok {
			result[key] = raw
		}
	}
	return result
}

func splitLines(content string) []sourceLine {
	if content == "" {
		return nil
	}
	parts := strings.SplitAfter(content, "\n")
	lines := make([]sourceLine, 0, len(parts))
	for index, part := range parts {
		if index == len(parts)-1 && part == "" {
			continue
		}
		ending := ""
		text := part
		if strings.HasSuffix(text, "\n") {
			ending = "\n"
			text = strings.TrimSuffix(text, "\n")
			if strings.HasSuffix(text, "\r") {
				ending = "\r\n"
				text = strings.TrimSuffix(text, "\r")
			}
		}
		lines = append(lines, sourceLine{text: text, ending: ending})
	}
	return lines
}

func renderLines(lines []sourceLine) string {
	var result strings.Builder
	for _, line := range lines {
		result.WriteString(line.text)
		result.WriteString(line.ending)
	}
	return result.String()
}

func detectLineEnding(lines []sourceLine) string {
	for _, line := range lines {
		if line.ending != "" {
			return line.ending
		}
	}
	return ""
}

func removeManagedBlock(lines []sourceLine) ([]sourceLine, error) {
	result := make([]sourceLine, 0, len(lines))
	inBlock := false
	found := false
	for _, line := range lines {
		switch strings.TrimSpace(line.text) {
		case beginMarker:
			if inBlock {
				return nil, fmt.Errorf("nested helper marker block")
			}
			inBlock = true
			found = true
		case endMarker:
			if !inBlock {
				return nil, fmt.Errorf("helper end marker without start")
			}
			inBlock = false
		default:
			if !inBlock {
				result = append(result, line)
			}
		}
	}
	if inBlock {
		return nil, fmt.Errorf("unterminated helper marker block")
	}
	_ = found
	return result, nil
}
