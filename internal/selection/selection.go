package selection

import "image"

// Selection は選択領域を表します
type Selection struct {
	Rect image.Rectangle
}

// Manager は複数の選択領域を管理します
type Manager struct {
	selections []Selection
}

// NewManager は新しいManagerインスタンスを作成します
func NewManager() *Manager {
	return &Manager{
		selections: make([]Selection, 0),
	}
}

// AddSelection は新しい選択領域を追加します
func (m *Manager) AddSelection(rect image.Rectangle) {
	m.selections = append(m.selections, Selection{Rect: rect})
}

// UndoSelection は最後の選択を取り消します
func (m *Manager) UndoSelection() {
	if len(m.selections) > 0 {
		m.selections = m.selections[:len(m.selections)-1]
	}
}

// ClearSelections は全ての選択をクリアします
func (m *Manager) ClearSelections() {
	m.selections = make([]Selection, 0)
}

// GetSelections は全ての選択領域を返します
func (m *Manager) GetSelections() []Selection {
	return m.selections
}

// HasSelection は選択が存在するかを返します
func (m *Manager) HasSelection() bool {
	return len(m.selections) > 0
}

// Count は選択の数を返します
func (m *Manager) Count() int {
	return len(m.selections)
}

// GetSelection は指定されたインデックスの選択を返します
func (m *Manager) GetSelection(index int) *Selection {
	if index < 0 || index >= len(m.selections) {
		return nil
	}
	return &m.selections[index]
}

// UpdateSelection は指定されたインデックスの選択を更新します
func (m *Manager) UpdateSelection(index int, rect image.Rectangle) {
	if index >= 0 && index < len(m.selections) {
		m.selections[index].Rect = rect
	}
}

// RemoveSelection は指定されたインデックスの選択を削除します
func (m *Manager) RemoveSelection(index int) {
	if index >= 0 && index < len(m.selections) {
		m.selections = append(m.selections[:index], m.selections[index+1:]...)
	}
}

// FindSelectionAt は指定された座標にある選択のインデックスを返します
func (m *Manager) FindSelectionAt(x, y int) int {
	point := image.Pt(x, y)
	for i := len(m.selections) - 1; i >= 0; i-- {
		if point.In(m.selections[i].Rect) {
			return i
		}
	}
	return -1
}

// IsInsideHandle は指定された座標が選択領域のハンドル上にあるかをチェックします
// handleSize: ハンドルのサイズ
// 戻り値: インデックス, ハンドルタイプ (0:なし, 1-8: 隅と辺)
func (m *Manager) IsInsideHandle(x, y int, handleSize int) (int, int) {
	for i := len(m.selections) - 1; i >= 0; i-- {
		rect := m.selections[i].Rect

		// 8つのハンドル位置をチェック
		// 1:左上, 2:上, 3:右上, 4:左, 5:右, 6:左下, 7:下, 8:右下
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

		for handleType, handle := range handles {
			if x >= handle.Min.X && x < handle.Max.X && y >= handle.Min.Y && y < handle.Max.Y {
				return i, handleType + 1
			}
		}
	}
	return -1, 0
}
