// Package dto はimport機能のデータ転送オブジェクトを定義する
package dto

// H3IndexSet はH3インデックスの重複排除セット
type H3IndexSet struct {
	indexes map[string]struct{}
}

// NewH3IndexSet は新しいH3IndexSetを作成する
func NewH3IndexSet() *H3IndexSet {
	return &H3IndexSet{
		indexes: make(map[string]struct{}),
	}
}

// Add はH3インデックスを追加する(空文字列は無視)
func (s *H3IndexSet) Add(index string) {
	if index != "" {
		s.indexes[index] = struct{}{}
	}
}

// AddAll は複数のH3インデックスを追加する
func (s *H3IndexSet) AddAll(indexes ...string) {
	for _, idx := range indexes {
		s.Add(idx)
	}
}

// ToSlice はセット内の全H3インデックスをスライスとして返す
func (s *H3IndexSet) ToSlice() []string {
	result := make([]string, 0, len(s.indexes))
	for idx := range s.indexes {
		result = append(result, idx)
	}
	return result
}

// Len はセット内の要素数を返す
func (s *H3IndexSet) Len() int {
	return len(s.indexes)
}

// IsEmpty はセットが空かどうかを判定する
func (s *H3IndexSet) IsEmpty() bool {
	return len(s.indexes) == 0
}
