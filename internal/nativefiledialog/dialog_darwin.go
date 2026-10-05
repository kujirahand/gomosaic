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

const saveScript = `on run argv
	try
		set selectedFile to choose file name with prompt "画像を保存" default name (item 1 of argv)
		return POSIX path of selectedFile
	on error number -128
		return ""
	end try
end run`

func OpenImagePath() (string, error) {
	return runAppleScript(openScript)
}

func SaveImagePath(defaultName string) (string, error) {
	return runAppleScript(saveScript, defaultName)
}

func runAppleScript(script string, args ...string) (string, error) {
	output, err := exec.Command("osascript", append([]string{"-e", script}, args...)...).Output()
	if err != nil {
		return "", fmt.Errorf("macOSファイルダイアログを開けません: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
