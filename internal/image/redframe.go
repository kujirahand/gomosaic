package image

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// DrawRedFrames は指定された矩形の枠線を赤色で画像に描き込みます。
// 各ピクセルと線の重なり面積から被覆率を計算するため、アンチエイリアスが効きます。
// widthPermille は線幅を画像の短辺に対する‰で指定します（最低2px）。
// frameColor は枠線の色です。
func DrawRedFrames(img image.Image, rects []image.Rectangle, widthPermille int, frameColor color.RGBA) image.Image {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	draw.Draw(result, bounds, img, bounds.Min, draw.Src)

		short := bounds.Dx()
	if bounds.Dy() < short {
		short = bounds.Dy()
	}
	width := math.Max(2, float64(short)*float64(widthPermille)/1000)
	half := width / 2

	for _, r := range rects {
		r = r.Intersect(bounds)
		if r.Empty() {
			continue
		}
		ox0, oy0 := float64(r.Min.X)-half, float64(r.Min.Y)-half
		ox1, oy1 := float64(r.Max.X)+half, float64(r.Max.Y)+half
		ix0, iy0 := float64(r.Min.X)+half, float64(r.Min.Y)+half
		ix1, iy1 := float64(r.Max.X)-half, float64(r.Max.Y)-half

		area := image.Rect(int(math.Floor(ox0)), int(math.Floor(oy0)),
			int(math.Ceil(ox1)), int(math.Ceil(oy1))).Intersect(bounds)
		for y := area.Min.Y; y < area.Max.Y; y++ {
			for x := area.Min.X; x < area.Max.X; x++ {
				outer := overlap(x, ox0, ox1) * overlap(y, oy0, oy1)
				inner := 0.0
				if ix1 > ix0 && iy1 > iy0 {
					inner = overlap(x, ix0, ix1) * overlap(y, iy0, iy1)
				}
				cov := outer - inner
				if cov <= 0 {
					continue
				}
				blend(result, x, y, frameColor, cov)
			}
		}
	}
	return result
}

// overlap はピクセル[p, p+1)と区間[a, b)の重なり長さ(0〜1)を返します
func overlap(p int, a, b float64) float64 {
	lo := math.Max(float64(p), a)
	hi := math.Min(float64(p+1), b)
	if hi <= lo {
		return 0
	}
	return hi - lo
}

func blend(dst *image.RGBA, x, y int, c color.RGBA, a float64) {
	i := dst.PixOffset(x, y)
	p := dst.Pix[i : i+4]
	inv := 1 - a
	p[0] = uint8(float64(p[0])*inv + float64(c.R)*a + 0.5)
	p[1] = uint8(float64(p[1])*inv + float64(c.G)*a + 0.5)
	p[2] = uint8(float64(p[2])*inv + float64(c.B)*a + 0.5)
	p[3] = 255
}
