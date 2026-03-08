package clipboard

import (
	"encoding/base64"
	"fmt"
	"os"
)

// CopyToClipboard copies the given text to the clipboard using OSC 52 escape sequence
func CopyToClipboard(text string) error {
	// Encode the text as base64
	encoded := base64.StdEncoding.EncodeToString([]byte(text))

	// OSC 52 sequence to copy to clipboard
	// Format: ESC ] 52 ; c ; BASE64_ENCODED_TEXT BEL
	// We use OSC 52 ; c ; ... to copy to the clipboard (c = clipboard)
	sequence := fmt.Sprintf("\033]52;c;%s\a", encoded)

	// Write the escape sequence to stdout
	// This works in terminal emulators that support OSC 52 (iTerm2, tmux, etc.)
	_, err := os.Stdout.WriteString(sequence)
	return err
}

// CopyToSelection copies the given text to the clipboard selection using OSC 52
// The 's' parameter specifies the selection (clipboard, primary, secondary)
func CopyToSelection(text string, selection string) error {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))

	// Format: ESC ] 52 ; s ; BASE64_ENCODED_TEXT BEL
	// s = selection: c=clipboard, p=primary, s=secondary
	sequence := fmt.Sprintf("\033]52;%s;%s\a", selection, encoded)

	_, err := os.Stdout.WriteString(sequence)
	return err
}
