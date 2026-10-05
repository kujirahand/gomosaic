//go:build windows

package nativefiledialog

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"unicode/utf16"
)

func OpenImagePath() (string, error) {
	return runPowerShellDialog(`
Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.OpenFileDialog
$dialog.Title = '画像を開く'
$dialog.Filter = '画像ファイル (*.png;*.jpg;*.jpeg)|*.png;*.jpg;*.jpeg|すべてのファイル (*.*)|*.*'
$dialog.Multiselect = $false
if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
    [Console]::Out.WriteLine($dialog.FileName)
}
$dialog.Dispose()
`)
}

func SaveImagePath(defaultName string) (string, error) {
	escaped := strings.ReplaceAll(defaultName, "'", "''")
	return runPowerShellDialog(`
Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.SaveFileDialog
$dialog.Title = '画像を保存'
$dialog.Filter = 'PNG画像 (*.png)|*.png|JPEG画像 (*.jpg;*.jpeg)|*.jpg;*.jpeg'
$dialog.FilterIndex = 1
$dialog.DefaultExt = 'png'
$dialog.AddExtension = $true
$dialog.OverwritePrompt = $true
$dialog.FileName = '` + escaped + `'
if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
    [Console]::Out.WriteLine($dialog.FileName)
}
$dialog.Dispose()
`)
}

func runPowerShellDialog(script string) (string, error) {
	script = "$OutputEncoding = [System.Text.Encoding]::UTF8\n[Console]::OutputEncoding = [System.Text.Encoding]::UTF8\n" + script
	encodedScript := base64.StdEncoding.EncodeToString(utf16LE(script))
	cmd := exec.Command("powershell.exe", "-NoLogo", "-NoProfile", "-STA", "-EncodedCommand", encodedScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Windowsファイルダイアログを開けません: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func utf16LE(value string) []byte {
	codes := utf16.Encode([]rune(value))
	data := make([]byte, len(codes)*2)
	for i, code := range codes {
		binary.LittleEndian.PutUint16(data[i*2:], code)
	}
	return data
}
