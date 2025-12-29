package entity

import "testing"

// TestNewLandCategory はNewLandCategoryがCodeとNameを正しく設定したLandCategoryを生成することをテストする
func TestNewLandCategory(t *testing.T) {
	code := "01"
	name := "田"

	landCategory := NewLandCategory(code, name)

	if landCategory.Code != code {
		t.Errorf("Code = %q, want %q", landCategory.Code, code)
	}
	if landCategory.Name != name {
		t.Errorf("Name = %q, want %q", landCategory.Name, name)
	}
}
