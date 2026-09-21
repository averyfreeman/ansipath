package shellinit_test

import (
	"strings"
	"testing"

	"github.com/averyfreeman/ansipath/internal/shellinit"
)

func TestRewriteMovesHelperLoadBeforeGeneratedCalls(t *testing.T) {
	input := "export PATH=\"/opt/first:/opt/second:$PATH\"\n"
	got, changes, err := shellinit.Rewrite(input, shellinit.Options{
		DataDir:    "/home/test/.local/share/ansipath",
		BinaryPath: "/home/test/.local/bin/ansipath",
		Environment: map[string]string{
			"PATH": "/usr/bin:/bin",
		},
	})
	if err != nil {
		t.Fatalf("Rewrite returned error: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("Rewrite returned %d changes, want 1", len(changes))
	}
	if strings.Contains(got, "export PATH=\"") {
		t.Fatalf("original export remains:\n%s", got)
	}
	load := strings.Index(got, "whichfunc.sh")
	firstCall := strings.Index(got, "pathmunge prepend")
	if load < 0 || firstCall < 0 || load > firstCall {
		t.Fatalf("helper load does not precede calls:\n%s", got)
	}
	if !strings.Contains(got, "pathmunge prepend '/opt/second'") ||
		!strings.Contains(got, "pathmunge prepend '/opt/first'") {
		t.Fatalf("prepend calls missing:\n%s", got)
	}
}

func TestRewriteResetsUnanchoredAssignmentsAndAppendsInOrder(t *testing.T) {
	got, _, err := shellinit.Rewrite(`PATH="/one:/two"
`, shellinit.Options{
		DataDir:     "/data/ansipath",
		BinaryPath:  "/bin/ansipath",
		Environment: map[string]string{"PATH": "/existing"},
	})
	if err != nil {
		t.Fatalf("Rewrite returned error: %v", err)
	}
	if !strings.Contains(got, "PATH=; export PATH") {
		t.Fatalf("missing reset:\n%s", got)
	}
	first := strings.Index(got, "pathmunge append '/one'")
	second := strings.Index(got, "pathmunge append '/two'")
	if first < 0 || second < 0 || first > second {
		t.Fatalf("append order is wrong:\n%s", got)
	}
}

func TestRewriteRecognizesBarePathAnchor(t *testing.T) {
	got, _, err := shellinit.Rewrite(`export PATH="/local:PATH:/extra"
`, shellinit.Options{
		DataDir:     "/data/ansipath",
		BinaryPath:  "/bin/ansipath",
		Environment: map[string]string{"PATH": "/usr/bin"},
	})
	if err != nil {
		t.Fatalf("Rewrite returned error: %v", err)
	}
	if !strings.Contains(got, "pathmunge prepend '/local'") ||
		!strings.Contains(got, "pathmunge append '/extra'") {
		t.Fatalf("bare PATH anchor was not converted:\n%s", got)
	}
}

func TestRewriteIsIdempotent(t *testing.T) {
	input := "export PATH=\"/one:$PATH\"\n"
	options := shellinit.Options{
		DataDir:     "/data/ansipath",
		BinaryPath:  "/bin/ansipath",
		Environment: map[string]string{"PATH": "/usr/bin"},
	}
	first, _, err := shellinit.Rewrite(input, options)
	if err != nil {
		t.Fatalf("first Rewrite returned error: %v", err)
	}
	second, _, err := shellinit.Rewrite(first, options)
	if err != nil {
		t.Fatalf("second Rewrite returned error: %v", err)
	}
	if first != second {
		t.Fatalf("Rewrite is not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestCleanRemovesDuplicateAndDeadLiteralEntries(t *testing.T) {
	input := "export PATH=\"/one:/one:/missing:$PATH\"\n"
	got, changes, err := shellinit.Clean(input, shellinit.CleanOptions{
		Options: shellinit.Options{
			DataDir:     "/data/ansipath",
			BinaryPath:  "/bin/ansipath",
			Environment: map[string]string{"PATH": "/usr/bin"},
		},
		Exists: func(path string) bool { return path == "/one" || path == "/usr/bin" },
	})
	if err != nil {
		t.Fatalf("Clean returned error: %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("Clean returned %d changes, want 2: %#v", len(changes), changes)
	}
	if !strings.Contains(got, "pathmunge prepend '/one'") || strings.Contains(got, "/missing") {
		t.Fatalf("cleaned profile is wrong:\n%s", got)
	}
	if changes[0].Reason != "duplicate" || changes[1].Reason != "dead" {
		t.Fatalf("unexpected change reasons: %#v", changes)
	}
}

func TestCleanPathValueRemovesDuplicatesAndDeadEntries(t *testing.T) {
	got, changes := shellinit.CleanPathValue(
		"/one:/one:/missing:/two",
		func(path string) bool { return path == "/one" || path == "/two" },
	)
	if got != "/one:/two" {
		t.Fatalf("CleanPathValue = %q", got)
	}
	if len(changes) != 2 || changes[0].Reason != "duplicate" || changes[1].Reason != "dead" {
		t.Fatalf("unexpected changes: %#v", changes)
	}
}
