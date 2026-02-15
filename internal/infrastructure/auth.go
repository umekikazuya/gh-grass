package infrastructure

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GetGHToken は gh CLI の `gh auth token` を実行してアクティブな GitHub トークンを返します。
// コマンド実行に失敗した場合は空文字とエラーを返します（エラーには gh がインストールされていないか認証されていない可能性を示す旨が含まれます）。
func GetGHToken() (string, error) {
	cmd := exec.Command("gh", "auth", "token")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get gh token: %w. stderr: %s. Make sure gh cli is installed and authenticated", err, stderr.String())
	}
	return strings.TrimSpace(out.String()), nil
}

