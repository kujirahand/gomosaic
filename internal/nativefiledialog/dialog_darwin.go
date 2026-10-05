//go:build darwin

package nativefiledialog

import (
	"fmt"
	"os/exec"
	"strings"
)

const openScript = `on run
	try
		set selectedFile to choose file with prompt "画像を開く"
		return POSIX path of selectedFile
	on error number -128
		return ""
	end try
end run`

const saveScript = `on run
	try
		set selectedFile to choose file name with prompt "画像を保存" default name "gomosaic.png"
		return POSIX path of selectedFile
	on error number -128
		return ""
	end try
end run`

func OpenImagePath() (string, error) {
	return runAppleScript(openScript)
}

func SaveImagePath() (string, error) {
	return runAppleScript(saveScript)
}

func runAppleScript(script string) (string, error) {
	output, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("macOSファイルダイアログを開けません: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
