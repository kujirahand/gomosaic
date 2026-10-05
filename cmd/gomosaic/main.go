package main

import (
	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"gomosaic/assets"
	"gomosaic/internal/app"
)

func main() {
	fyneApp := fyneapp.NewWithID("com.github.gomosaic")
	fyneApp.SetIcon(fyne.NewStaticResource("icon.png", assets.IconPNG))
	application := app.NewApplication(fyneApp)
	application.Run()
}
