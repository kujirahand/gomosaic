package ui

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	fyneImage "gomosaic/internal/image"
	"gomosaic/internal/selection"
)

// SelectableImage はドラッグで領域選択可能な画像ウィジェット
type SelectableImage struct {
	widget.BaseWidget
	img                 *canvas.Image
	selectRect          *canvas.Rectangle
	image               image.Image
	originalImage       image.Image // Original/current full-resolution source image
	selectionMgr        *selection.Manager
	dragStart           *fyne.Position
	selectedIndex       int  // 選択中のインデックス（-1で未選択）
	isDraggingSelection bool // 選択領域を移動中
	isResizing          bool // リサイズ中
	resizeHandle        int  // リサイズハンドル番号（1-8）
	dragOriginRect      image.Rectangle
	handleSize          int     // ハンドルのサイズ
	zoomLevel           float64 // ズームレベル（1.0 = 100%）
}

// NewSelectableImage は新しいSelectableImageを作成します
func NewSelectableImage() *SelectableImage {
	si := &SelectableImage{
		selectionMgr:  selection.NewManager(),
		selectedIndex: -1,
		handleSize:    8,
		zoomLevel:     1.0,
	}

	si.selectRect = canvas.NewRectangle(color.RGBA{R: 0, G: 120, B: 215, A: 80})
	si.selectRect.StrokeColor = color.RGBA{R: 0, G: 120, B: 215, A: 255}
	si.selectRect.StrokeWidth = 2
	si.selectRect.Hide()

	si.ExtendBaseWidget(si)
	return si
}

// SetImage は画像を設定します
func (si *SelectableImage) SetImage(img image.Image) {
	si.image = img
	si.originalImage = img
	si.zoomLevel = 1
	si.selectionMgr.ClearSelections()
	si.selectedIndex = -1
	si.dragStart = nil
	si.isDraggingSelection = false
	si.isResizing = false
	si.selectRect.Hide()

	// デフォルトでは元のサイズで表示
	si.img = canvas.NewImageFromImage(img)
	si.img.FillMode = canvas.ImageFillContain
	si.Refresh()
}

// FitToScreen は画像を指定された画面サイズに合わせます
func (si *SelectableImage) FitToScreen(screenWidth, screenHeight int) {
	if si.originalImage == nil {
		return
	}

	if screenWidth <= 0 || screenHeight <= 0 {
		return
	}

	bounds := si.originalImage.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	if imgWidth <= 0 || imgHeight <= 0 {
		return
	}

	// 画面に収まるようにズームレベルを計算
	zoomX := float64(screenWidth) / float64(imgWidth)
	zoomY := float64(screenHeight) / float64(imgHeight)

	// 小さい方に合わせる（画像全体が表示されるように）
	zoomLevel := zoomX
	if zoomY < zoomX {
		zoomLevel = zoomY
	}

	// 最小値を設定（見えなくならないように）
	if zoomLevel < 0.1 {
		zoomLevel = 0.1
	}

	si.SetZoomLevel(zoomLevel)
}

// SetZoomLevel はズームレベルを設定し、画像の表示サイズを変更します
func (si *SelectableImage) SetZoomLevel(level float64) {
	if level < 0.1 {
		level = 0.1
	}
	if level > 10.0 {
		level = 10.0
	}

	si.zoomLevel = level

	// 画像が設定されている場合は表示サイズを変更
	if si.originalImage != nil {
		bounds := si.originalImage.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()

		newWidth := max(1, int(float64(width)*level))
		newHeight := max(1, int(float64(height)*level))

		// canvas.Imageの表示サイズを変更（画像データは変更しない）
		si.img = canvas.NewImageFromImage(si.originalImage)
		si.img.FillMode = canvas.ImageFillContain
		si.img.SetMinSize(fyne.NewSize(float32(newWidth), float32(newHeight)))
		si.image = si.originalImage // 座標計算用に元の画像を維持
		si.Refresh()
	}
}

// displayToImage converts widget-local display coordinates into source pixels.
func (si *SelectableImage) displayToImage(pos fyne.Position) fyne.Position {
	if si.zoomLevel <= 0 || si.originalImage == nil {
		return pos
	}
	bounds := si.originalImage.Bounds()
	return fyne.NewPos(
		float32(bounds.Min.X)+pos.X/float32(si.zoomLevel),
		float32(bounds.Min.Y)+pos.Y/float32(si.zoomLevel),
	)
}

// CreateRenderer はウィジェットのレンダラーを作成します
func (si *SelectableImage) CreateRenderer() fyne.WidgetRenderer {
	if si.img == nil {
		label := canvas.NewText("画像を開いてください", color.Gray{Y: 128})
		label.Alignment = fyne.TextAlignCenter
		return &selectableImageRenderer{
			widget:  si,
			objects: []fyne.CanvasObject{label},
		}
	}
	return &selectableImageRenderer{
		widget:  si,
		objects: []fyne.CanvasObject{si.img, si.selectRect},
	}
}

// Dragged はドラッグ中に呼ばれます
func (si *SelectableImage) Dragged(e *fyne.DragEvent) {
	if si.image == nil {
		return
	}

	pos := e.Position

	// Fyne starts delivering Dragged events after the pointer has moved. Recover
	// the mouse-down position from the first event's delta so the initial motion
	// is included in the selection.
	if si.dragStart == nil {
		start := fyne.NewPos(pos.X-e.Dragged.DX, pos.Y-e.Dragged.DY)
		imageStart := si.displayToImage(start)
		x, y := int(imageStart.X), int(imageStart.Y)
		handleSize := max(1, int(float64(si.handleSize)/si.zoomLevel))

		// Start resizing only from a corner handle.
		idx, handleType := si.selectionMgr.IsInsideHandle(x, y, handleSize)
		if handleType != 1 && handleType != 3 && handleType != 6 && handleType != 8 {
			idx, handleType = -1, 0
		}
		if idx >= 0 {
			si.selectedIndex = idx
			si.isResizing = true
			si.isDraggingSelection = false
			si.resizeHandle = handleType
			si.dragStart = &start
			si.dragOriginRect = si.selectionMgr.GetSelections()[idx].Rect
			si.Refresh()
		} else if idx = si.selectionMgr.FindSelectionAt(x, y); idx >= 0 {
			// Drag inside an existing selection to move it.
			si.selectedIndex = idx
			si.isDraggingSelection = true
			si.isResizing = false
			si.dragStart = &start
			si.dragOriginRect = si.selectionMgr.GetSelections()[idx].Rect
			si.Refresh()
		} else {
			// Start a new selection in empty image space.
			si.selectedIndex = -1
			si.isDraggingSelection = false
			si.isResizing = false
			si.dragStart = &start
			si.selectRect.Move(start)
			si.selectRect.Resize(fyne.NewSize(1, 1))
			si.selectRect.Show()
		}
	}

	// 選択領域の移動中
	if si.isDraggingSelection && si.selectedIndex >= 0 {
		if si.dragStart != nil {
			start := si.displayToImage(*si.dragStart)
			current := si.displayToImage(pos)
			dx := int(current.X - start.X)
			dy := int(current.Y - start.Y)
			newRect := si.clampRect(si.dragOriginRect.Add(image.Pt(dx, dy)))
			si.selectionMgr.UpdateSelection(si.selectedIndex, newRect)
			si.Refresh()
		}
		return
	}

	// 選択領域のリサイズ中
	if si.isResizing && si.selectedIndex >= 0 {
		if si.selectionMgr.GetSelection(si.selectedIndex) != nil {
			newRect := si.resizeSelection(si.dragOriginRect, si.displayToImage(pos))
			si.selectionMgr.UpdateSelection(si.selectedIndex, newRect)
			si.Refresh()
		}
		return
	}

	// 新しい選択領域の作成
	if si.dragStart != nil {
		// 選択領域を更新
		x1, y1 := si.dragStart.X, si.dragStart.Y
		x2, y2 := pos.X, pos.Y

		minX, maxX := x1, x2
		if x2 < x1 {
			minX, maxX = x2, x1
		}
		minY, maxY := y1, y2
		if y2 < y1 {
			minY, maxY = y2, y1
		}

		si.selectRect.Move(fyne.NewPos(minX, minY))
		si.selectRect.Resize(fyne.NewSize(maxX-minX, maxY-minY))

		si.Refresh()
	}
}

// resizeSelection はハンドルに基づいて選択領域をリサイズします
func (si *SelectableImage) resizeSelection(rect image.Rectangle, pos fyne.Position) image.Rectangle {
	x, y := int(pos.X), int(pos.Y)
	if si.originalImage != nil {
		bounds := si.originalImage.Bounds()
		x = max(bounds.Min.X, min(x, bounds.Max.X))
		y = max(bounds.Min.Y, min(y, bounds.Max.Y))
	}

	switch si.resizeHandle {
	case 1: // 左上
		return si.clampRect(image.Rect(min(x, rect.Max.X-1), min(y, rect.Max.Y-1), rect.Max.X, rect.Max.Y))
	case 3: // 右上
		return si.clampRect(image.Rect(rect.Min.X, min(y, rect.Max.Y-1), max(x, rect.Min.X+1), rect.Max.Y))
	case 6: // 左下
		return si.clampRect(image.Rect(min(x, rect.Max.X-1), rect.Min.Y, rect.Max.X, max(y, rect.Min.Y+1)))
	case 8: // 右下
		return si.clampRect(image.Rect(rect.Min.X, rect.Min.Y, max(x, rect.Min.X+1), max(y, rect.Min.Y+1)))
	}
	return rect
}

func (si *SelectableImage) clampRect(rect image.Rectangle) image.Rectangle {
	if si.originalImage == nil {
		return rect
	}
	bounds := si.originalImage.Bounds()
	width, height := rect.Dx(), rect.Dy()
	if width > bounds.Dx() || height > bounds.Dy() {
		return rect.Intersect(bounds)
	}
	if rect.Min.X < bounds.Min.X {
		rect = rect.Add(image.Pt(bounds.Min.X-rect.Min.X, 0))
	}
	if rect.Min.Y < bounds.Min.Y {
		rect = rect.Add(image.Pt(0, bounds.Min.Y-rect.Min.Y))
	}
	if rect.Max.X > bounds.Max.X {
		rect = rect.Add(image.Pt(bounds.Max.X-rect.Max.X, 0))
	}
	if rect.Max.Y > bounds.Max.Y {
		rect = rect.Add(image.Pt(0, bounds.Max.Y-rect.Max.Y))
	}
	return rect
}

// DragEnd はドラッグ終了時に呼ばれます
func (si *SelectableImage) DragEnd() {
	// 移動またはリサイズ操作の終了
	if si.isDraggingSelection || si.isResizing {
		si.isDraggingSelection = false
		si.isResizing = false
		si.dragStart = nil
		si.Refresh()
		return
	}

	if si.dragStart == nil || si.image == nil {
		si.dragStart = nil
		si.selectRect.Hide()
		return
	}

	// 選択領域を確定
	pos := si.selectRect.Position()
	size := si.selectRect.Size()

	if size.Width > 5 && size.Height > 5 {
		start := si.displayToImage(pos)
		end := si.displayToImage(fyne.NewPos(pos.X+size.Width, pos.Y+size.Height))
		rect := si.clampRect(image.Rect(int(start.X), int(start.Y), int(end.X), int(end.Y)))
		si.selectionMgr.AddSelection(rect)
		si.selectedIndex = si.selectionMgr.Count() - 1

	}

	si.dragStart = nil
	si.selectRect.Hide()
	si.Refresh()
}

// Tapped はクリック時に呼ばれます
func (si *SelectableImage) Tapped(e *fyne.PointEvent) {
	if si.image == nil {
		return
	}

	imagePos := si.displayToImage(e.Position)
	x, y := int(imagePos.X), int(imagePos.Y)

	// ハンドル上のクリックをチェック
	handleSize := max(1, int(float64(si.handleSize)/si.zoomLevel))
	idx, handleType := si.selectionMgr.IsInsideHandle(x, y, handleSize)
	if idx >= 0 {
		si.selectedIndex = idx
		si.resizeHandle = handleType
		si.Refresh()
		return
	}

	// 既存の選択領域内のクリックをチェック
	idx = si.selectionMgr.FindSelectionAt(x, y)
	if idx >= 0 {
		si.selectedIndex = idx
		si.Refresh()
		return
	}

	// A click outside all selection rectangles clears the current selections.
	si.ClearSelections()
	si.dragStart = nil
	si.isDraggingSelection = false
	si.isResizing = false
	si.selectRect.Hide()
}

// UndoSelection は最後の選択を取り消します
func (si *SelectableImage) UndoSelection() {
	si.selectionMgr.UndoSelection()
	if si.selectedIndex >= si.selectionMgr.Count() {
		si.selectedIndex = si.selectionMgr.Count() - 1
	}
	si.Refresh()
}

// ClearSelections は全ての選択をクリアします
func (si *SelectableImage) ClearSelections() {
	si.selectionMgr.ClearSelections()
	si.selectedIndex = -1
	si.Refresh()
}

// GetSelectionManager は選択マネージャーを返します
func (si *SelectableImage) GetSelectionManager() *selection.Manager {
	return si.selectionMgr
}

// HasImage は画像が読み込まれているかを返します
func (si *SelectableImage) HasImage() bool {
	return si.image != nil
}

// GetImage は現在の画像を返します
func (si *SelectableImage) GetImage() image.Image {
	return si.image
}

// selectableImageRenderer はSelectableImageのレンダラー
type selectableImageRenderer struct {
	widget  *SelectableImage
	objects []fyne.CanvasObject
}

func (r *selectableImageRenderer) Layout(size fyne.Size) {
	// Scroll expands its content to at least the viewport size. Do not stretch
	// the image to that larger area: pointer-to-image coordinates and selection
	// overlays are based on the zoomed image size, anchored at the content origin.
	if r.widget.img == nil || r.widget.originalImage == nil {
		// テキストラベルがある場合は、サイズと位置を設定
		for _, obj := range r.objects {
			obj.Move(fyne.NewPos(0, 0))
			obj.Resize(size)
		}
		return
	}
	bounds := r.widget.originalImage.Bounds()
	imageSize := fyne.NewSize(
		max(1, float32(float64(bounds.Dx())*r.widget.zoomLevel)),
		max(1, float32(float64(bounds.Dy())*r.widget.zoomLevel)),
	)
	r.widget.img.Move(fyne.NewPos(0, 0))
	r.widget.img.Resize(imageSize)
}

func (r *selectableImageRenderer) MinSize() fyne.Size {
	if r.widget.originalImage != nil {
		bounds := r.widget.originalImage.Bounds()
		width := float32(float64(bounds.Dx()) * r.widget.zoomLevel)
		height := float32(float64(bounds.Dy()) * r.widget.zoomLevel)
		return fyne.NewSize(width, height)
	}
	return fyne.NewSize(400, 300)
}

func (r *selectableImageRenderer) Refresh() {
	r.objects = []fyne.CanvasObject{}
	if r.widget.img != nil {
		objects := []fyne.CanvasObject{r.widget.img}

		// 既存の選択領域を描画
		selections := r.widget.selectionMgr.GetSelections()
		for i, sel := range selections {
			bounds := r.widget.originalImage.Bounds()
			rect := image.Rect(
				int(float64(sel.Rect.Min.X-bounds.Min.X)*r.widget.zoomLevel),
				int(float64(sel.Rect.Min.Y-bounds.Min.Y)*r.widget.zoomLevel),
				int(float64(sel.Rect.Max.X-bounds.Min.X)*r.widget.zoomLevel),
				int(float64(sel.Rect.Max.Y-bounds.Min.Y)*r.widget.zoomLevel),
			)
			selRect := canvas.NewRectangle(color.RGBA{R: 0, G: 120, B: 215, A: 40})
			selRect.StrokeColor = color.RGBA{R: 0, G: 120, B: 215, A: 200}
			selRect.StrokeWidth = 2

			// 選択中の場合は別の色
			if i == r.widget.selectedIndex {
				selRect.FillColor = color.RGBA{R: 255, G: 165, B: 0, A: 60}
				selRect.StrokeColor = color.RGBA{R: 255, G: 165, B: 0, A: 255}
				selRect.StrokeWidth = 3

				// リサイズハンドルを描画
				handleSize := r.widget.handleSize
				handles := []image.Rectangle{
					image.Rect(rect.Min.X-handleSize/2, rect.Min.Y-handleSize/2, rect.Min.X+handleSize/2, rect.Min.Y+handleSize/2),                         // 1:左上
					image.Rect(rect.Min.X+(rect.Dx()-handleSize)/2, rect.Min.Y-handleSize/2, rect.Min.X+(rect.Dx()+handleSize)/2, rect.Min.Y+handleSize/2), // 2:上
					image.Rect(rect.Max.X-handleSize/2, rect.Min.Y-handleSize/2, rect.Max.X+handleSize/2, rect.Min.Y+handleSize/2),                         // 3:右上
					image.Rect(rect.Min.X-handleSize/2, rect.Min.Y+(rect.Dy()-handleSize)/2, rect.Min.X+handleSize/2, rect.Min.Y+(rect.Dy()+handleSize)/2), // 4:左
					image.Rect(rect.Max.X-handleSize/2, rect.Min.Y+(rect.Dy()-handleSize)/2, rect.Max.X+handleSize/2, rect.Min.Y+(rect.Dy()+handleSize)/2), // 5:右
					image.Rect(rect.Min.X-handleSize/2, rect.Max.Y-handleSize/2, rect.Min.X+handleSize/2, rect.Max.Y+handleSize/2),                         // 6:左下
					image.Rect(rect.Min.X+(rect.Dx()-handleSize)/2, rect.Max.Y-handleSize/2, rect.Min.X+(rect.Dx()+handleSize)/2, rect.Max.Y+handleSize/2), // 7:下
					image.Rect(rect.Max.X-handleSize/2, rect.Max.Y-handleSize/2, rect.Max.X+handleSize/2, rect.Max.Y+handleSize/2),                         // 8:右下
				}

				for _, handleIndex := range []int{0, 2, 5, 7} {
					handle := handles[handleIndex]
					handleRect := canvas.NewRectangle(color.RGBA{R: 255, G: 255, B: 255, A: 255})
					handleRect.StrokeColor = color.RGBA{R: 255, G: 165, B: 0, A: 255}
					handleRect.StrokeWidth = 2
					handleRect.Move(fyne.NewPos(float32(handle.Min.X), float32(handle.Min.Y)))
					handleRect.Resize(fyne.NewSize(float32(handle.Dx()), float32(handle.Dy())))
					objects = append(objects, handleRect)
				}
			}

			selRect.Move(fyne.NewPos(float32(rect.Min.X), float32(rect.Min.Y)))
			selRect.Resize(fyne.NewSize(float32(rect.Dx()), float32(rect.Dy())))
			objects = append(objects, selRect)
		}

		// 新しい選択用の矩形
		objects = append(objects, r.widget.selectRect)
		r.objects = objects
	}
	// A custom renderer must invalidate the widget's canvas after rebuilding
	// its child objects, otherwise pointer-driven overlay changes can remain
	// invisible until some unrelated redraw occurs.
	canvas.Refresh(r.widget)
}

func (r *selectableImageRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *selectableImageRenderer) Destroy() {}

// ImageCanvas はUI全体を管理するキャンバス（後方互換性のため）
type ImageCanvas struct {
	selectableImage *SelectableImage
	scroll          *container.Scroll
	container       *fyne.Container
	mosaicProcessor *fyneImage.MosaicProcessor
	processedImage  image.Image
	mosaicStrength  int // モザイクブロック幅（画像短辺に対する‰）
}

// NewImageCanvas は新しいImageCanvasを作成します
func NewImageCanvas() *ImageCanvas {
	ic := &ImageCanvas{
		selectableImage: NewSelectableImage(),
		mosaicProcessor: fyneImage.NewMosaicProcessor(10), // デフォルトのブロックサイズ
		mosaicStrength:  10,
	}
	// The scroll viewport fills the available area; centering it would make its
	// viewport follow the content size and prevent useful scrolling.
	ic.scroll = container.NewScroll(ic.selectableImage)
	ic.container = container.NewMax(ic.scroll)
	return ic
}

// SetImage は画像を設定して表示します
func (ic *ImageCanvas) SetImage(img image.Image) {
	ic.selectableImage.SetImage(img)
	// 画面サイズに合わせて画像をリサイズ
	ic.FitToScreen()
	ic.container.Refresh()
}

// FitToScreen は画像を画面サイズに合わせます
func (ic *ImageCanvas) FitToScreen() {
	if ic.selectableImage.HasImage() {
		// Scroll already occupies the center area between toolbar and status bar.
		viewport := ic.scroll.Size()
		ic.selectableImage.FitToScreen(int(viewport.Width), int(viewport.Height))
		ic.scroll.Refresh()
	}
}

// GetContainer はこのキャンバスのコンテナを返します
func (ic *ImageCanvas) GetContainer() *fyne.Container {
	return ic.container
}

// HasImage は画像が読み込まれているかを返します
func (ic *ImageCanvas) HasImage() bool {
	return ic.selectableImage.HasImage()
}

// GetImage は現在の画像を返します
func (ic *ImageCanvas) GetImage() image.Image {
	return ic.selectableImage.GetImage()
}

// UndoSelection は最後の選択を取り消します
func (ic *ImageCanvas) UndoSelection() {
	ic.selectableImage.UndoSelection()
}

// ClearSelections は全ての選択をクリアします
func (ic *ImageCanvas) ClearSelections() {
	ic.selectableImage.ClearSelections()
}

// GetSelectionManager は選択マネージャーを返します
func (ic *ImageCanvas) GetSelectionManager() *selection.Manager {
	return ic.selectableImage.GetSelectionManager()
}

// ApplyMosaic は画像の指定された領域にモザイクを適用します
func (ic *ImageCanvas) ApplyMosaic(img image.Image, rects []image.Rectangle) image.Image {
	ic.mosaicProcessor.SetBlockSize(mosaicBlockSize(img.Bounds(), ic.mosaicStrength))
	return ic.mosaicProcessor.ApplyMosaicMultiple(img, rects)
}

// mosaicBlockSize calculates a pixel block width from the image's shorter edge.
// Strength is expressed in per-mille: 10 means 1% of the shorter edge.
func mosaicBlockSize(bounds image.Rectangle, strength int) int {
	strength = max(5, min(strength, 30))
	shortEdge := min(bounds.Dx(), bounds.Dy())
	if shortEdge <= 0 {
		return 1
	}
	return max(1, (shortEdge*strength+500)/1000)
}

// SetMosaicStrength sets the block width as a per-mille ratio of image size.
func (ic *ImageCanvas) SetMosaicStrength(strength int) {
	ic.mosaicStrength = max(5, min(strength, 30))
}

// MosaicStrength returns the configured per-mille strength.
func (ic *ImageCanvas) MosaicStrength() int {
	return ic.mosaicStrength
}

// SetZoomLevel はズームレベルを設定します
func (ic *ImageCanvas) SetZoomLevel(level float64) {
	ic.selectableImage.SetZoomLevel(level)
	// Recalculate the scroll bars after the content's MinSize changes.
	ic.scroll.Refresh()
	ic.container.Refresh()
}

// GetZoomLevel は現在のズームレベルを取得します
func (ic *ImageCanvas) GetZoomLevel() float64 {
	return ic.selectableImage.zoomLevel
}
