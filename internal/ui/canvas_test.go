package ui

import (
	"image"
	"testing"

	"fyne.io/fyne/v2"
	fyneTest "fyne.io/fyne/v2/test"
)

func startTestApp(t *testing.T) {
	t.Helper()
	app := fyneTest.NewApp()
	t.Cleanup(app.Quit)
}

func TestSelectableImageMinSizeTracksZoom(t *testing.T) {
	startTestApp(t)
	selectable := NewSelectableImage()
	selectable.SetImage(image.NewRGBA(image.Rect(0, 0, 200, 100)))
	renderer := selectable.CreateRenderer()

	selectable.SetZoomLevel(1.5)
	got := renderer.MinSize()
	if got.Width != 300 || got.Height != 150 {
		t.Fatalf("MinSize at 150%% zoom = %v, want 300x150", got)
	}
	renderer.Layout(fyne.NewSize(500, 400))
	if got := selectable.img.Size(); got.Width != 300 || got.Height != 150 {
		t.Fatalf("image display size in larger viewport = %v, want 300x150", got)
	}
	if got := selectable.img.Position(); got != fyne.NewPos(0, 0) {
		t.Fatalf("image display position = %v, want top-left origin", got)
	}

	selectable.SetZoomLevel(0.5)
	got = renderer.MinSize()
	if got.Width != 100 || got.Height != 50 {
		t.Fatalf("MinSize at 50%% zoom = %v, want 100x50", got)
	}
}

func TestSelectableImageCreateMoveAndResizeSelection(t *testing.T) {
	startTestApp(t)
	selectable := NewSelectableImage()
	source := image.NewRGBA(image.Rect(0, 0, 100, 100))
	selectable.SetImage(source)

	// Drag in empty space from (10,10) to (40,40).
	selectable.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{Position: fyne.NewPos(40, 40)},
		Dragged:    fyne.NewDelta(30, 30),
	})
	if !selectable.selectRect.Visible() {
		t.Fatal("selection rectangle should be shown while dragging")
	}
	if got := selectable.selectRect.Position(); got != fyne.NewPos(10, 10) {
		t.Fatalf("live selection position = %v, want (10,10)", got)
	}
	if got := selectable.selectRect.Size(); got.Width != 30 || got.Height != 30 {
		t.Fatalf("live selection size = %v, want 30x30", got)
	}
	if selectable.GetImage() != source {
		t.Fatal("dragging a selection must not modify the image before Apply Mosaic")
	}
	selectable.DragEnd()
	if got := selectable.selectionMgr.GetSelections()[0].Rect; got != image.Rect(10, 10, 40, 40) {
		t.Fatalf("created selection = %v, want (10,10)-(40,40)", got)
	}

	// Drag inside the rectangle to move it by (10,10).
	selectable.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{Position: fyne.NewPos(30, 30)},
		Dragged:    fyne.NewDelta(10, 10),
	})
	selectable.DragEnd()
	if got := selectable.selectionMgr.GetSelections()[0].Rect; got != image.Rect(20, 20, 50, 50) {
		t.Fatalf("moved selection = %v, want (20,20)-(50,50)", got)
	}

	// Drag the bottom-right corner out by 10 pixels in both directions.
	selectable.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{Position: fyne.NewPos(60, 60)},
		Dragged:    fyne.NewDelta(10, 10),
	})
	selectable.DragEnd()
	if got := selectable.selectionMgr.GetSelections()[0].Rect; got != image.Rect(20, 20, 60, 60) {
		t.Fatalf("resized selection = %v, want (20,20)-(60,60)", got)
	}
}

func TestMosaicBlockSizeScalesWithImageDimensions(t *testing.T) {
	smallImageBlock := mosaicBlockSize(image.Rect(0, 0, 1000, 500), 10)
	largeImageBlock := mosaicBlockSize(image.Rect(0, 0, 4000, 2000), 10)
	if smallImageBlock != 5 || largeImageBlock != 20 {
		t.Fatalf("block sizes for scaled images = %d and %d, want 5 and 20", smallImageBlock, largeImageBlock)
	}
	if got := mosaicBlockSize(image.Rect(0, 0, 4000, 2000), 30); got != 60 {
		t.Fatalf("block size at 30‰ = %d, want 60", got)
	}
}

func TestClickOutsideSelectionsClearsThem(t *testing.T) {
	startTestApp(t)
	selectable := NewSelectableImage()
	selectable.SetImage(image.NewRGBA(image.Rect(0, 0, 100, 100)))
	selectable.selectionMgr.AddSelection(image.Rect(10, 10, 40, 40))
	selectable.selectedIndex = 0

	selectable.Tapped(&fyne.PointEvent{Position: fyne.NewPos(80, 80)})
	if selectable.selectionMgr.Count() != 0 {
		t.Fatalf("selection count after outside click = %d, want 0", selectable.selectionMgr.Count())
	}
}
