package entity

import "testing"

// TestNewIdleLandStatus はNewIdleLandStatusがCodeとNameを正しく設定したIdleLandStatusを生成することをテストする
func TestNewIdleLandStatus(t *testing.T) {
	code := "1"
	name := "遊休農地"

	idleLandStatus := NewIdleLandStatus(code, name)

	if idleLandStatus.Code != code {
		t.Errorf("Code = %q, want %q", idleLandStatus.Code, code)
	}
	if idleLandStatus.Name != name {
		t.Errorf("Name = %q, want %q", idleLandStatus.Name, name)
	}
}
