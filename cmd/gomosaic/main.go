package main

import (
	fyneapp "fyne.io/fyne/v2/app"
	"gomosaic/internal/app"
)

func main() {
	fyneApp := fyneapp.NewWithID("com.github.gomosaic")
	application := app.NewApplication(fyneApp)
	application.Run()
}
