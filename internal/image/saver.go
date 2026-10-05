package image

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// Saver は画像を保存する構造体
type Saver struct {
	jpegQuality int // JPEG品質 (1-100)
}

// NewSaver は新しいSaverインスタンスを作成します
func NewSaver() *Saver {
	return &Saver{
		jpegQuality: 90,
	}
}

// SaveImage は画像を指定されたパスに保存します
func (s *Saver) SaveImage(img image.Image, path string) error {
	// ファイルを作成
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("ファイルを作成できません: %w", err)
	}
	defer file.Close()

	// 拡張子に基づいてエンコード
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return png.Encode(file, img)
	case ".jpg", ".jpeg":
		return jpeg.Encode(file, img, &jpeg.Options{Quality: s.jpegQuality})
	default:
		return fmt.Errorf("サポートされていない形式です: %s", ext)
	}
}

// SetJPEGQuality はJPEG品質を設定します
func (s *Saver) SetJPEGQuality(quality int) {
	if quality < 1 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}
	s.jpegQuality = quality
}

// GetJPEGQuality は現在のJPEG品質を取得します
func (s *Saver) GetJPEGQuality() int {
	return s.jpegQuality
}
