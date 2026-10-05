package app

import (
	"fyne.io/fyne/v2"
	"gomosaic/internal/config"
	"gomosaic/internal/ui"
)

// Application はメインアプリケーションを管理します
type Application struct {
	fyneApp fyne.App
	config  *config.Config
	window  *ui.Window
}

// NewApplication は新しいApplicationインスタンスを作成し、起動します
func NewApplication(fyneApp fyne.App) *Application {
	// 設定を読み込む
	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	a := &Application{
		fyneApp: fyneApp,
		config:  cfg,
		window:  ui.NewWindow(fyneApp, cfg),
	}
	return a
}

// Run はアプリケーションを実行します
func (a *Application) Run() {
	// 終了時に設定を保存
	defer func() {
		if err := config.SaveConfig(a.config); err != nil {
			// エラーは無視（設定保存の失敗は致命的ではない）
		}
	}()
	a.window.ShowAndRun()
}
