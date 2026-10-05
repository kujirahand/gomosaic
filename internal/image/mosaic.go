package image

import (
	"image"
	"image/color"
)

// MosaicProcessor はモザイク処理を行う構造体
type MosaicProcessor struct {
	blockSize int // モザイクのブロックサイズ（ピクセル）
}

// NewMosaicProcessor は新しいMosaicProcessorを作成します
func NewMosaicProcessor(blockSize int) *MosaicProcessor {
	if blockSize < 1 {
		blockSize = 10
	}
	return &MosaicProcessor{blockSize: blockSize}
}

// ApplyMosaic は指定された領域にモザイクを適用します
func (mp *MosaicProcessor) ApplyMosaic(img image.Image, rect image.Rectangle) image.Image {
	// 元画像のBoundsを取得
	bounds := img.Bounds()
	
	// 新しいRGBA画像を作成
	result := image.NewRGBA(bounds)
	
	// 元画像をコピー
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			result.Set(x, y, img.At(x, y))
		}
	}
	
	// モザイクを適用する領域を制限
	mosaicRect := rect.Intersect(bounds)
	if mosaicRect.Empty() {
		return result
	}
	
	// モザイク処理
	for y := mosaicRect.Min.Y; y < mosaicRect.Max.Y; y += mp.blockSize {
		for x := mosaicRect.Min.X; x < mosaicRect.Max.X; x += mp.blockSize {
			// ブロックの範囲を計算
			blockEndX := x + mp.blockSize
			if blockEndX > mosaicRect.Max.X {
				blockEndX = mosaicRect.Max.X
			}
			blockEndY := y + mp.blockSize
			if blockEndY > mosaicRect.Max.Y {
				blockEndY = mosaicRect.Max.Y
			}
			
			// ブロック内の平均色を計算
			avgColor := mp.calculateAverageColor(img, x, y, blockEndX, blockEndY)
			
			// ブロック全体を平均色で塗りつぶす
			for by := y; by < blockEndY; by++ {
				for bx := x; bx < blockEndX; bx++ {
					result.Set(bx, by, avgColor)
				}
			}
		}
	}
	
	return result
}

// ApplyMosaicMultiple は複数の領域にモザイクを適用します
func (mp *MosaicProcessor) ApplyMosaicMultiple(img image.Image, rects []image.Rectangle) image.Image {
	result := img
	for _, rect := range rects {
		result = mp.ApplyMosaic(result, rect)
	}
	return result
}

// calculateAverageColor は指定された領域の平均色を計算します
func (mp *MosaicProcessor) calculateAverageColor(img image.Image, x1, y1, x2, y2 int) color.Color {
	var rSum, gSum, bSum, aSum uint64
	count := uint64(0)
	
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			rSum += uint64(r)
			gSum += uint64(g)
			bSum += uint64(b)
			aSum += uint64(a)
			count++
		}
	}
	
	if count == 0 {
		return color.Black
	}
	
	// 平均を計算（RGBAは16ビットなので8ビットに変換）
	rAvg := uint8((rSum / count) >> 8)
	gAvg := uint8((gSum / count) >> 8)
	bAvg := uint8((bSum / count) >> 8)
	aAvg := uint8((aSum / count) >> 8)
	
	return color.RGBA{R: rAvg, G: gAvg, B: bAvg, A: aAvg}
}

// SetBlockSize はブロックサイズを設定します
func (mp *MosaicProcessor) SetBlockSize(blockSize int) {
	if blockSize < 1 {
		blockSize = 10
	}
	mp.blockSize = blockSize
}

// GetBlockSize は現在のブロックサイズを取得します
func (mp *MosaicProcessor) GetBlockSize() int {
	return mp.blockSize
}
