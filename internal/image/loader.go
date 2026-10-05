package image

import (
	"fmt"
	"image"
	_ "image/jpeg" // JPEG形式のサポートを登録
	_ "image/png"  // PNG形式のサポートを登録
	"os"
	"path/filepath"
	"strings"
)

// Loader は画像ファイルを読み込む構造体です
type Loader struct {
	// 必要に応じて設定を追加
}

// NewLoader は新しいLoaderインスタンスを作成します
func NewLoader() *Loader {
	return &Loader{}
}

// LoadImage は指定されたパスから画像を読み込みます
func (l *Loader) LoadImage(path string) (image.Image, error) {
	// ファイルの存在確認
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("ファイルが存在しません: %s", path)
	}

	// ファイル拡張子をチェック
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		return nil, fmt.Errorf("サポートされていない形式です: %s (PNG, JPEGのみ対応)", ext)
	}

	// ファイルを開く
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ファイルを開けません: %w", err)
	}
	defer file.Close()

	// 画像をデコード
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("画像のデコードに失敗しました: %w", err)
	}

	return img, nil
}
