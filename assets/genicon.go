//go:build ignore

// アイコン生成: go run assets/genicon.go
package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

const size = 1024

// 5x5 のモザイク配色（左上が明るく右下が濃い青）
var palette = [5][5]color.RGBA{}

func main() {
	light := [3]float64{150, 220, 255}
	dark := [3]float64{20, 70, 170}
	pattern := [5][5]float64{
		{0.10, 0.25, 0.15, 0.40, 0.55},
		{0.30, 0.05, 0.45, 0.35, 0.60},
		{0.20, 0.50, 0.30, 0.65, 0.75},
		{0.45, 0.35, 0.70, 0.55, 0.90},
		{0.60, 0.80, 0.65, 0.95, 0.70},
	}
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			t := pattern[y][x]
			palette[y][x] = color.RGBA{
				uint8(light[0] + (dark[0]-light[0])*t),
				uint8(light[1] + (dark[1]-light[1])*t),
				uint8(light[2] + (dark[2]-light[2])*t), 255}
		}
	}

	const (
		margin = 64    // 外側の余白
		gap    = 12    // タイル間の隙間
		radius = 180.0 // 全体の角丸半径
	)
	body := size - margin*2
	cell := float64(body) / 5

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := margin; y < size-margin; y++ {
		for x := margin; x < size-margin; x++ {
			if !insideRoundRect(float64(x-margin)+0.5, float64(y-margin)+0.5, float64(body), radius) {
				continue
			}
			cx := int(float64(x-margin) / cell)
			cy := int(float64(y-margin) / cell)
			// タイル間の隙間は白
			fx := float64(x-margin) - float64(cx)*cell
			fy := float64(y-margin) - float64(cy)*cell
			if fx < gap/2 || fx > cell-gap/2 || fy < gap/2 || fy > cell-gap/2 {
				img.SetRGBA(x, y, color.RGBA{255, 255, 255, 255})
				continue
			}
			img.SetRGBA(x, y, palette[cy][cx])
		}
	}

	f, err := os.Create("assets/icon.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

func insideRoundRect(x, y, s, r float64) bool {
	dx, dy := 0.0, 0.0
	if x < r {
		dx = r - x
	} else if x > s-r {
		dx = x - (s - r)
	}
	if y < r {
		dy = r - y
	} else if y > s-r {
		dy = y - (s - r)
	}
	return dx*dx+dy*dy <= r*r
}
