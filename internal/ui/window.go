package ui

import (
	"fmt"
	"image"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"gomosaic/internal/config"
	fyneImage "gomosaic/internal/image"
	"gomosaic/internal/nativefiledialog"
)

// Window はメインウィンドウを管理します
type Window struct {
	window         fyne.Window
	imageCanvas    *ImageCanvas
	imageLoader    *fyneImage.Loader
	imageSaver     *fyneImage.Saver
	statusBar      *widget.Label
	config         *config.Config
	historyManager *HistoryManager
	zoomLevel      float64 // 拡大・縮小レベル（1.0 = 100%）
}

// NewWindow は新しいウィンドウを作成します
func NewWindow(fyneApp fyne.App, cfg *config.Config) *Window {
	mosaicStrength := max(5, min(cfg.MosaicBlockSize, 30))
	cfg.MosaicBlockSize = mosaicStrength
	w := &Window{
		imageLoader:    fyneImage.NewLoader(),
		imageSaver:     fyneImage.NewSaver(),
		statusBar:      widget.NewLabel("画像を開いてください"),
		config:         cfg,
		historyManager: NewHistoryManager(20), // 最大20ステップの履歴
		zoomLevel:      1.0,                   // デフォルトのズームレベル
	}

	w.window = fyneApp.NewWindow("ゴモザイク - 画像モザイクツール")
	w.imageCanvas = NewImageCanvas()

	// ドラッグ＆ドロップの設定
	w.window.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		if len(uris) == 0 {
			return
		}

		// 最初の画像ファイルを探す
		for _, uri := range uris {
			// ファイルスキームかチェック
			if uri.Scheme() != "file" {
				continue
			}

			// 画像ファイルの場合のみ処理
			path := uri.Path()
			img, err := w.imageLoader.LoadImage(path)
			if err != nil {
				dialog.ShowError(fmt.Errorf("画像の読み込みに失敗しました: %w", err), w.window)
				continue
			}

			w.imageCanvas.SetImage(img)
			w.historyManager.AddImage(img)
			w.updateStatus()

			// 画面サイズに合わせて表示
			w.imageCanvas.FitToScreen()
			return
		}

		// 有効な画像ファイルが見つからなかった場合
		if len(uris) > 0 {
			dialog.ShowError(fmt.Errorf("サポートされている画像ファイル（PNG、JPEG）をドロップしてください"), w.window)
		}
	})

	// 設定からモザイク強度を復元
	w.imageCanvas.SetMosaicStrength(mosaicStrength)

	// ファイル開くボタン
	openBtn := widget.NewButton("📁 画像を開く", func() {
		w.OpenFileDialog()
	})

	// 保存ボタン
	saveBtn := widget.NewButton("💾 保存", func() {
		w.SaveFileDialog()
	})

	// Undo button for the current image-edit history.
	undoBtn := widget.NewButton("↶ Undo", func() {
		w.undoHandler()
	})

	// ズームアウトボタン
	zoomOutBtn := widget.NewButton("🔍-", func() {
		w.zoomOut()
	})

	// ズームインボタン
	zoomInBtn := widget.NewButton("🔍+", func() {
		w.zoomIn()
	})

	// モザイク幅スライダー（画像短辺に対する‰、設定から初期値を復元）
	mosaicSlider := widget.NewSlider(5, 30)
	mosaicSlider.SetValue(float64(mosaicStrength))
	mosaicValue := widget.NewLabel(fmt.Sprintf("%d‰", mosaicStrength))
	mosaicSlider.OnChanged = func(value float64) {
		strength := int(value)
		w.imageCanvas.SetMosaicStrength(strength)
		// 設定に保存
		w.config.MosaicBlockSize = strength
		mosaicValue.SetText(fmt.Sprintf("%d‰", strength))
	}
	sliderLabel := widget.NewLabel("モザイク幅:")
	slider := container.NewGridWrap(fyne.NewSize(200, mosaicSlider.MinSize().Height), mosaicSlider)

	// モザイク適用ボタン
	applyBtn := widget.NewButton("✨ モザイク適用", func() {
		w.applyMosaic()
		w.updateStatus()
	})

	// 赤枠描画ボタン
	redFrameBtn := widget.NewButton("🟥 赤枠を描画", func() {
		w.applyRedFrame()
	})

	// ツールバー風のレイアウト
	topBar := container.NewHBox(openBtn, saveBtn, undoBtn, zoomOutBtn, zoomInBtn, sliderLabel, slider, mosaicValue, applyBtn, redFrameBtn)
	content := container.NewBorder(topBar, w.statusBar, nil, nil, w.imageCanvas.GetContainer())

	w.window.SetContent(content)
	w.window.Resize(fyne.NewSize(1200, 800))

	// メニューバーの設定
	w.setupMenu()

	return w
}

// OpenFileDialog はファイル選択ダイアログを表示します
func (w *Window) OpenFileDialog() {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		path, err := nativefiledialog.OpenImagePath()
		if err != nil {
			dialog.ShowError(err, w.window)
			return
		}
		if path != "" { // Empty path means the native dialog was cancelled.
			w.openImage(path)
		}
		return
	}

	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		w.openImage(reader.URI().Path())
	}, w.window)

	// サポートされているファイル形式をフィルタ
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg"}))

	fd.Show()
}

// ShowAndRun はウィンドウを表示してアプリケーションを実行します
func (w *Window) ShowAndRun() {
	w.window.ShowAndRun()
}

// setupMenu はメニューバーを設定します
func (w *Window) setupMenu() {
	// ファイルメニュー
	fileMenu := fyne.NewMenu("ファイル",
		fyne.NewMenuItem("画像を開く", func() {
			w.OpenFileDialog()
		}),
		fyne.NewMenuItem("保存", func() {
			w.SaveFileDialog()
		}),
	)
	fileMenu.Items[0].Shortcut = &fyne.ShortcutCut{} // Ctrl+O の代わり

	// 編集メニュー
	editMenu := fyne.NewMenu("編集",
		fyne.NewMenuItem("取り消し", func() {
			w.imageCanvas.UndoSelection()
			w.updateStatus()
		}),
		fyne.NewMenuItem("全クリア", func() {
			w.imageCanvas.ClearSelections()
			w.updateStatus()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("元に戻す", func() {
			w.undoHandler()
		}),
		fyne.NewMenuItem("やり直す", func() {
			w.redoHandler()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("モザイク適用", func() {
			w.applyMosaic()
			w.updateStatus()
		}),
	)

	// ヘルプメニュー
	helpMenu := fyne.NewMenu("ヘルプ",
		fyne.NewMenuItem("このアプリについて", func() {
			dialog.ShowInformation("ゴモザイクについて",
				"ゴモザイク v1.0\n\n画像の選択部分にモザイクを掛けるツールです。",
				w.window)
		}),
	)

	// メインメニューを設定
	mainMenu := fyne.NewMainMenu(fileMenu, editMenu, helpMenu)
	w.window.SetMainMenu(mainMenu)
}

// SaveFileDialog は保存ダイアログを表示します
func (w *Window) SaveFileDialog() {
	if !w.imageCanvas.HasImage() {
		dialog.ShowInformation("警告", "画像が開かれていません", w.window)
		return
	}

	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		path, err := nativefiledialog.SaveImagePath()
		if err != nil {
			dialog.ShowError(err, w.window)
			return
		}
		if path != "" { // Empty path means the native dialog was cancelled.
			w.saveImage(path)
		}
		return
	}

	fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w.window)
			return
		}
		if writer == nil {
			return
		}
		path := writer.URI().Path()
		if closeErr := writer.Close(); closeErr != nil {
			dialog.ShowError(closeErr, w.window)
			return
		}
		w.saveImage(path)
	}, w.window)

	// サポートされているファイル形式をフィルタ
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg"}))

	fd.Show()
}

func (w *Window) openImage(path string) {
	img, err := w.imageLoader.LoadImage(path)
	if err != nil {
		dialog.ShowError(err, w.window)
		return
	}
	w.imageCanvas.SetImage(img)
	w.historyManager.AddImage(img)
	w.updateStatus()
}

func (w *Window) saveImage(path string) {
	if filepath.Ext(path) == "" {
		path += ".png"
	}
	if err := w.imageSaver.SaveImage(w.imageCanvas.GetImage(), path); err != nil {
		dialog.ShowError(err, w.window)
		return
	}
	dialog.ShowInformation("成功", fmt.Sprintf("画像を保存しました: %s", path), w.window)
}

// applyMosaic は選択領域にモザイクを適用します
func (w *Window) applyMosaic() {
	if !w.imageCanvas.HasImage() {
		dialog.ShowInformation("警告", "画像が開かれていません", w.window)
		return
	}

	selections := w.imageCanvas.GetSelectionManager().GetSelections()
	if len(selections) == 0 {
		dialog.ShowInformation("警告", "選択領域がありません", w.window)
		return
	}

	// 選択領域の矩形を取得
	rects := make([]image.Rectangle, len(selections))
	for i, sel := range selections {
		rects[i] = sel.Rect
	}

	// モザイクを適用
	img := w.imageCanvas.GetImage()
	processed := w.imageCanvas.ApplyMosaic(img, rects)
	w.imageCanvas.SetImage(processed)
	w.imageCanvas.ClearSelections()

	// 履歴に追加
	w.historyManager.AddImage(processed)
	w.updateStatus()
}

// applyRedFrame は選択領域の枠線を赤色で画像に描き込みます
func (w *Window) applyRedFrame() {
	if !w.imageCanvas.HasImage() {
		dialog.ShowInformation("警告", "画像が開かれていません", w.window)
		return
	}

	selections := w.imageCanvas.GetSelectionManager().GetSelections()
	if len(selections) == 0 {
		dialog.ShowInformation("警告", "選択領域がありません", w.window)
		return
	}

	rects := make([]image.Rectangle, len(selections))
	for i, sel := range selections {
		rects[i] = sel.Rect
	}

	processed := fyneImage.DrawRedFrames(w.imageCanvas.GetImage(), rects)
	w.imageCanvas.SetImage(processed)
	w.imageCanvas.ClearSelections()

	// 履歴に追加（Undo対応）
	w.historyManager.AddImage(processed)
	w.updateStatus()
}

// undoHandler はモザイク適用を元に戻します
func (w *Window) undoHandler() {
	if img, ok := w.historyManager.Undo(); ok {
		w.imageCanvas.SetImage(img)
		w.updateStatus()
	} else {
		dialog.ShowInformation("情報", "これ以上元に戻せません", w.window)
	}
}

// redoHandler はモザイク適用をやり直します
func (w *Window) redoHandler() {
	if img, ok := w.historyManager.Redo(); ok {
		w.imageCanvas.SetImage(img)
		w.updateStatus()
	} else {
		dialog.ShowInformation("情報", "これ以上やり直せません", w.window)
	}
}

// zoomIn は画像を10%拡大します
func (w *Window) zoomIn() {
	currentZoom := w.imageCanvas.GetZoomLevel()
	newZoom := currentZoom * 1.1
	w.imageCanvas.SetZoomLevel(newZoom)
	w.zoomLevel = newZoom
	w.updateStatus()
}

// zoomOut は画像を10%縮小します
func (w *Window) zoomOut() {
	currentZoom := w.imageCanvas.GetZoomLevel()
	newZoom := currentZoom * 0.9
	w.imageCanvas.SetZoomLevel(newZoom)
	w.zoomLevel = newZoom
	w.updateStatus()
}

// updateStatus はステータスバーを更新します
func (w *Window) updateStatus() {
	if !w.imageCanvas.HasImage() {
		w.statusBar.SetText("画像を開いてください")
		return
	}

	img := w.imageCanvas.GetImage()
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	selectionCount := w.imageCanvas.GetSelectionManager().Count()
	w.statusBar.SetText(fmt.Sprintf("画像サイズ: %dx%d | 表示倍率: %.0f%% | 選択領域: %d", width, height, w.imageCanvas.GetZoomLevel()*100, selectionCount))
}
