package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config はアプリケーションの設定を保持します
type Config struct {
	MosaicBlockSize int `json:"mosaic_block_size"`
}

// DefaultConfig はデフォルトの設定を返します
func DefaultConfig() *Config {
	return &Config{
		MosaicBlockSize: 10,
	}
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
