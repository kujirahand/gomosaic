package config

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
)

// Config はアプリケーションの設定を保持します
type Config struct {
	MosaicBlockSize int    `json:"mosaic_block_size"`
	RedFrameWidth   int    `json:"red_frame_width"` // 画像短辺に対する‰
	RedFrameColor   string `json:"red_frame_color"` // "#RRGGBB"
}

// DefaultConfig はデフォルトの設定を返します
func DefaultConfig() *Config {
	return &Config{
		MosaicBlockSize: 10,
		RedFrameWidth:   3,
		RedFrameColor:   "#FF0000",
	}
}

// ParseHexColor は "#RRGGBB" 形式の文字列を色に変換します
func ParseHexColor(s string) (color.RGBA, bool) {
	var r, g, b uint8
	if len(s) != 7 || s[0] != '#' {
		return color.RGBA{}, false
	}
	if _, err := fmt.Sscanf(s[1:], "%02x%02x%02x", &r, &g, &b); err != nil {
		return color.RGBA{}, false
	}
	return color.RGBA{r, g, b, 255}, true
}

// FormatHexColor は色を "#RRGGBB" 形式の文字列にします
func FormatHexColor(c color.Color) string {
	n := color.RGBAModel.Convert(c).(color.RGBA)
	return fmt.Sprintf("#%02X%02X%02X", n.R, n.G, n.B)
}

// LoadConfig は設定ファイルから設定を読み込みます
func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		// ファイルが存在しない場合はデフォルトを返す
		return DefaultConfig(), nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}

	if cfg.RedFrameWidth == 0 {
		cfg.RedFrameWidth = DefaultConfig().RedFrameWidth
	}

	if _, ok := ParseHexColor(cfg.RedFrameColor); !ok {
		cfg.RedFrameColor = DefaultConfig().RedFrameColor
	}

	return &cfg, nil
}

// SaveConfig は設定をファイルに保存します
func SaveConfig(cfg *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// ディレクトリが存在しない場合は作成
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// getConfigPath は設定ファイルのパスを返します
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".gomosaic", "config.json"), nil
}
