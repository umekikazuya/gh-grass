package infrastructure

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GetGHToken executes `gh auth token` to retrieve the active GitHub token.
func GetGHToken() (string, error) {
	cmd := exec.Command("gh", "auth", "token")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get gh token: %w. Make sure gh cli is installed and authenticated", err)
	}
	return strings.TrimSpace(out.String()), nil
}
