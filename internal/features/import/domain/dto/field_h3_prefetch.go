// Package dto はimport機能のデータ転送オブジェクトを定義する
package dto

// FieldH3Prefetch はフィールドの既存H3インデックス情報(差分更新のプリフェッチ用)
type FieldH3Prefetch struct {
	ID          string
	H3IndexRes3 *string
	H3IndexRes5 *string
	H3IndexRes7 *string
	H3IndexRes9 *string
}

// AllIndexes は全解像度のH3インデックスをスライスで返す(nil値は除外)
func (f *FieldH3Prefetch) AllIndexes() []string {
	var result []string
	if f.H3IndexRes3 != nil {
		result = append(result, *f.H3IndexRes3)
	}
	if f.H3IndexRes5 != nil {
		result = append(result, *f.H3IndexRes5)
	}
	if f.H3IndexRes7 != nil {
		result = append(result, *f.H3IndexRes7)
	}
	if f.H3IndexRes9 != nil {
		result = append(result, *f.H3IndexRes9)
	}
	return result
}
