package pathview

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Style contains the ANSI sequences used by the renderer and diagnostics.
type Style struct {
	Palette    []string
	Reset      string
	YellowBold string
	DimRed     string
}

// Render returns the colorized PATH report for pathSource.
func Render(pathSource string, style Style, exists func(string) bool) (string, error) {
	pathSource = Sanitize(pathSource)
	if pathSource == "" {
		return "", fmt.Errorf("no valid path tracking string found")
	}
	if len(style.Palette) == 0 {
		return "", fmt.Errorf("palette is empty")
	}
	if exists == nil {
		exists = func(path string) bool {
			_, err := os.Stat(path)
			return err == nil
		}
	}

	entries := splitEntries(pathSource)
	if len(entries) == 0 {
		return "", fmt.Errorf("no valid path tracking string found")
	}

	counts := make(map[string]int, len(entries))
	orderedPaths := make([]string, 0, len(entries))
	deadCounts := make(map[string]int)
	rendered := make([]string, 0, len(entries))
	for _, entry := range entries {
		if counts[entry] == 0 {
			orderedPaths = append(orderedPaths, entry)
		}
		counts[entry]++

		line := colorizeLine(entry, style.Palette) + style.Reset
		if !exists(entry) {
			line += " " + style.DimRed + "[DEAD LINK]" + style.Reset
			deadCounts[entry]++
		}
		rendered = append(rendered, line)
	}

	var output strings.Builder
	output.WriteString(strings.Join(rendered, "\n"))
	output.WriteByte('\n')

	duplicates := make([]string, 0)
	for _, path := range orderedPaths {
		if counts[path] > 1 {
			duplicates = append(duplicates, path)
		}
	}
	if len(duplicates) > 0 {
		output.WriteString("\n")
		output.WriteString(style.YellowBold)
		output.WriteString("⚠️  Warning: Duplicate entries detected in your $PATH:")
		output.WriteString(style.Reset)
		output.WriteByte('\n')
		for _, path := range duplicates {
			fmt.Fprintf(&output, "  • %s (%d times)\n", path, counts[path])
		}
	}

	if len(deadCounts) > 0 {
		dead := make([]string, 0, len(deadCounts))
		for path := range deadCounts {
			dead = append(dead, path)
		}
		sort.Strings(dead)

		output.WriteString("\n")
		output.WriteString(style.DimRed)
		output.WriteString("✖  Alert: Broken paths found (directories do not exist):")
		output.WriteString(style.Reset)
		output.WriteByte('\n')
		for _, path := range dead {
			fmt.Fprintf(&output, "  • %s", path)
			if deadCounts[path] > 1 {
				fmt.Fprintf(&output, " (%d times)", deadCounts[path])
			}
			output.WriteByte('\n')
		}
	}

	return output.String(), nil
}

// Sanitize removes the wrapper emitted by common which implementations.
func Sanitize(raw string) string {
	text := strings.TrimSpace(raw)
	if strings.Contains(text, "no ") && strings.Contains(text, " in (") {
		start := strings.IndexByte(text, '(')
		end := strings.LastIndexByte(text, ')')
		if start >= 0 && end > start {
			text = text[start+1 : end]
		}
	}
	return strings.Trim(text, "()")
}

func splitEntries(pathSource string) []string {
	parts := strings.Split(pathSource, ":")
	entries := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			entries = append(entries, part)
		}
	}
	return entries
}

func colorizeLine(line string, colors []string) string {
	components := strings.Split(line, "/")
	var output strings.Builder
	for i, component := range components {
		if i == 0 && component == "" {
			continue
		}
		output.WriteString(colors[i%len(colors)])
		output.WriteByte('/')
		output.WriteString(component)
	}
	return output.String()
}
