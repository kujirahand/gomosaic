package image

import (
	"image"
	"image/color"
	"testing"
)

func TestDrawRedFrames(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for i := range src.Pix {
		src.Pix[i] = 255
	}
	out := DrawRedFrames(src, []image.Rectangle{image.Rect(20, 20, 80, 80)}, 3, color.RGBA{255, 0, 0, 255}).(*image.RGBA)

	if got := out.RGBAAt(20, 50); got != (color.RGBA{255, 0, 0, 255}) {
		t.Errorf("edge pixel = %v, want solid red", got)
	}
	if got := out.RGBAAt(50, 50); got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("center pixel changed: %v", got)
	}
	if src.RGBAAt(20, 50) != (color.RGBA{255, 255, 255, 255}) {
		t.Error("source image was modified")
	}
}
