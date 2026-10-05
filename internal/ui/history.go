package ui

import "image"

// HistoryManager は画像処理の履歴を管理します
type HistoryManager struct {
	history []image.Image
	current int
	maxSize int
}

// NewHistoryManager は新しいHistoryManagerを作成します
func NewHistoryManager(maxSize int) *HistoryManager {
	return &HistoryManager{
		history: make([]image.Image, 0),
		current: -1,
		maxSize: maxSize,
	}
}

// AddImage は画像を履歴に追加します
func (h *HistoryManager) AddImage(img image.Image) {
	// 現在の位置以降の履歴を削除
	if h.current < len(h.history)-1 {
		h.history = h.history[:h.current+1]
	}

	// 新しい画像を追加
	h.history = append(h.history, img)
	h.current++

	// 最大サイズを超えた場合、古いものを削除
	if len(h.history) > h.maxSize {
		h.history = h.history[1:]
		h.current--
	}
}

// Undo は一つ前の画像を返します
func (h *HistoryManager) Undo() (image.Image, bool) {
	if h.current <= 0 {
		return nil, false
	}
	h.current--
	return h.history[h.current], true
}

// Redo は一つ次の画像を返します
func (h *HistoryManager) Redo() (image.Image, bool) {
	if h.current >= len(h.history)-1 {
		return nil, false
	}
	h.current++
	return h.history[h.current], true
}

// CanUndo はUndo可能かどうかを返します
func (h *HistoryManager) CanUndo() bool {
	return h.current > 0
}

// CanRedo はRedo可能かどうかを返します
func (h *HistoryManager) CanRedo() bool {
	return h.current < len(h.history)-1
}

// GetCurrent は現在の画像を返します
func (h *HistoryManager) GetCurrent() image.Image {
	if h.current < 0 || h.current >= len(h.history) {
		return nil
	}
	return h.history[h.current]
}
