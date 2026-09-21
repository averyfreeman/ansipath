package app

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/averyfreeman/ansipath/internal/shellinit"
)

type clipboardCommand struct {
	name string
	args []string
}

var clipboardCommands = []clipboardCommand{
	{name: "pbcopy"},
	{name: "wl-copy"},
	{name: "xclip", args: []string{"-selection", "clipboard"}},
	{name: "xsel", args: []string{"--clipboard", "--input"}},
}

var clipboardCopy = copyToClipboard

func runEnvironmentCleanup(out io.Writer) int {
	cleaned, changes := shellinit.CleanPathValue(os.Getenv("PATH"), nil)
	exportLine := "export PATH=" + shellQuote(cleaned)

	if len(changes) == 0 {
		_, _ = io.WriteString(out, "# ansipath: no duplicate or dead PATH entries found.\n")
	} else {
		for _, change := range changes {
			_, _ = fmt.Fprintf(out, "# changed: env PATH removed %s %s\n", change.Reason, shellQuote(change.Path))
		}
		_, _ = io.WriteString(out, "if command -v pathmunge >/dev/null 2>&1; then\n")
		for _, change := range changes {
			_, _ = fmt.Fprintf(out, "    pathmunge remove %s\n", shellQuote(change.Path))
		}
		_, _ = io.WriteString(out, "fi\n")
	}

	copied, copyErr := clipboardCopy(exportLine)
	switch {
	case copyErr != nil:
		_, _ = io.WriteString(out, "# ansipath: clipboard copy failed; copy the final export manually.\n")
	case copied:
		_, _ = io.WriteString(out, "# ansipath: copied the cleaned PATH export to the clipboard.\n")
	default:
		_, _ = io.WriteString(out, "# ansipath: clipboard utility unavailable; copy the final export manually.\n")
	}
	_, _ = fmt.Fprintln(out, exportLine)
	return 0
}

func copyToClipboard(value string) (bool, error) {
	found := false
	var lastErr error
	for _, candidate := range clipboardCommands {
		binary, err := exec.LookPath(candidate.name)
		if err != nil {
			continue
		}
		found = true
		command := exec.Command(binary, candidate.args...)
		command.Stdin = strings.NewReader(value + "\n")
		if err := command.Run(); err == nil {
			return true, nil
		} else {
			lastErr = fmt.Errorf("%s: %w", candidate.name, err)
		}
	}
	if found {
		return true, lastErr
	}
	return false, nil
}
