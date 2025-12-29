package entity

import "testing"

// TestNewSoilType はNewSoilTypeがLargeCode、MiddleCode、SmallCode、SmallNameを正しく設定したSoilTypeを生成することをテストする
func TestNewSoilType(t *testing.T) {
	largeCode := "A"
	middleCode := "A1"
	smallCode := "A1a"
	smallName := "黒ボク土"

	soilType := NewSoilType(largeCode, middleCode, smallCode, smallName)

	if soilType.LargeCode != largeCode {
		t.Errorf("LargeCode = %q, want %q", soilType.LargeCode, largeCode)
	}
	if soilType.MiddleCode != middleCode {
		t.Errorf("MiddleCode = %q, want %q", soilType.MiddleCode, middleCode)
	}
	if soilType.SmallCode != smallCode {
		t.Errorf("SmallCode = %q, want %q", soilType.SmallCode, smallCode)
	}
	if soilType.SmallName != smallName {
		t.Errorf("SmallName = %q, want %q", soilType.SmallName, smallName)
	}
}
