package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

func readUndoResponse(in io.Reader) (bool, error) {
	file, ok := in.(*os.File)
	if !ok || !term.IsTerminal(int(file.Fd())) {
		return readUndoLine(in)
	}

	state, err := term.MakeRaw(int(file.Fd()))
	if err != nil {
		return readUndoLine(in)
	}
	answer, readErr := readUndoKeys(file)
	restoreErr := term.Restore(int(file.Fd()), state)
	if readErr != nil {
		return false, readErr
	}
	if restoreErr != nil {
		return false, fmt.Errorf("restore terminal: %w", restoreErr)
	}
	return answer, nil
}

func readUndoLine(in io.Reader) (bool, error) {
	answer, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(answer), "y"), nil
}

func readUndoKeys(in io.Reader) (bool, error) {
	var key [1]byte
	for {
		count, err := in.Read(key[:])
		if count > 0 {
			switch key[0] {
			case 'y', 'Y':
				return true, nil
			case 'n', 'N', '\r', '\n':
				return false, nil
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return false, nil
			}
			return false, err
		}
	}
}
