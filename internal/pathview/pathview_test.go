package pathview_test

import (
	"strings"
	"testing"

	"github.com/averyfreeman/ansipath/internal/pathview"
)

func TestRenderReportsColorizedEntriesAndDiagnostics(t *testing.T) {
	style := pathview.Style{
		Palette:    []string{"R", "G"},
		Reset:      "RESET",
		YellowBold: "YELLOW",
		DimRed:     "DIMRED",
	}

	output, err := pathview.Render(
		"/usr/bin:/usr/bin:/missing:/missing",
		style,
		func(path string) bool { return path != "/missing" },
	)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	for _, want := range []string{
		"G/usrR/binRESET",
		"G/missingRESET DIMRED[DEAD LINK]RESET",
		"Duplicate entries detected",
		"/usr/bin (2 times)",
		"Broken paths found",
		"  • /missing (2 times)",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q:\n%s", want, output)
		}
	}
}

func TestSanitizeWhichFailureOutput(t *testing.T) {
	got := pathview.Sanitize("which: no missingapp in (/usr/bin:/bin)")
	if got != "/usr/bin:/bin" {
		t.Fatalf("Sanitize returned %q", got)
	}
}

func TestRenderRejectsEmptyInput(t *testing.T) {
	if _, err := pathview.Render("", pathview.Style{Palette: []string{"R"}}, nil); err == nil {
		t.Fatal("Render accepted empty input")
	}
}
