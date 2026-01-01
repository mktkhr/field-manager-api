package h3util

import (
	"testing"

	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
	"github.com/stretchr/testify/require"
)

// TestNewBoundingBox はNewBoundingBoxが正しい値でBoundingBoxを生成することをテストする
func TestNewBoundingBox(t *testing.T) {
	swLat, swLng := 35.0, 139.0
	neLat, neLng := 36.0, 140.0

	bb := NewBoundingBox(swLat, swLng, neLat, neLng)

	require.NotNil(t, bb, "BoundingBoxがnilです")

	if bb.SWLat != swLat {
		t.Errorf("SWLat = %f, 期待値 %f", bb.SWLat, swLat)
	}
	if bb.SWLng != swLng {
		t.Errorf("SWLng = %f, 期待値 %f", bb.SWLng, swLng)
	}
	if bb.NELat != neLat {
		t.Errorf("NELat = %f, 期待値 %f", bb.NELat, neLat)
	}
	if bb.NELng != neLng {
		t.Errorf("NELng = %f, 期待値 %f", bb.NELng, neLng)
	}
}

// TestBoundingBox_IsValid はIsValidメソッドがBoundingBoxの有効性を正しく判定することをテストする
func TestBoundingBox_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		bb    *BoundingBox
		valid bool
	}{
		// 正常系: 有効なBoundingBox
		{
			name:  "通常の有効なBoundingBox",
			bb:    NewBoundingBox(35.0, 139.0, 36.0, 140.0),
			valid: true,
		},
		{
			name:  "赤道付近のBoundingBox",
			bb:    NewBoundingBox(-1.0, 100.0, 1.0, 102.0),
			valid: true,
		},
		{
			name:  "南半球のBoundingBox",
			bb:    NewBoundingBox(-45.0, 170.0, -40.0, 175.0),
			valid: true,
		},
		// 境界値: 最大/最小値
		{
			name:  "緯度の最大値",
			bb:    NewBoundingBox(89.0, 0.0, 90.0, 10.0),
			valid: true,
		},
		{
			name:  "緯度の最小値",
			bb:    NewBoundingBox(-90.0, 0.0, -89.0, 10.0),
			valid: true,
		},
		{
			name:  "経度の最大値",
			bb:    NewBoundingBox(0.0, 170.0, 10.0, 180.0),
			valid: true,
		},
		{
			name:  "経度の最小値",
			bb:    NewBoundingBox(0.0, -180.0, 10.0, -170.0),
			valid: true,
		},
		// 異常系: 無効なBoundingBox
		{
			name:  "南西緯度が範囲外(91)",
			bb:    NewBoundingBox(91.0, 139.0, 92.0, 140.0),
			valid: false,
		},
		{
			name:  "南西緯度が範囲外(-91)",
			bb:    NewBoundingBox(-91.0, 139.0, 36.0, 140.0),
			valid: false,
		},
		{
			name:  "北東緯度が範囲外(91)",
			bb:    NewBoundingBox(35.0, 139.0, 91.0, 140.0),
			valid: false,
		},
		{
			name:  "北東緯度が範囲外(-91)",
			bb:    NewBoundingBox(35.0, 139.0, -91.0, 140.0),
			valid: false,
		},
		{
			name:  "南西経度が範囲外(181)",
			bb:    NewBoundingBox(35.0, 181.0, 36.0, 140.0),
			valid: false,
		},
		{
			name:  "南西経度が範囲外(-181)",
			bb:    NewBoundingBox(35.0, -181.0, 36.0, 140.0),
			valid: false,
		},
		{
			name:  "北東経度が範囲外(181)",
			bb:    NewBoundingBox(35.0, 139.0, 36.0, 181.0),
			valid: false,
		},
		{
			name:  "北東経度が範囲外(-181)",
			bb:    NewBoundingBox(35.0, 139.0, 36.0, -181.0),
			valid: false,
		},
		{
			name:  "南西緯度が北東緯度より大きい",
			bb:    NewBoundingBox(37.0, 139.0, 36.0, 140.0),
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bb.IsValid()
			if got != tt.valid {
				t.Errorf("IsValid() = %v, 期待値 %v", got, tt.valid)
			}
		})
	}
}

// TestBoundingBox_Contains はContainsメソッドが座標の包含判定を正しく行うことをテストする
func TestBoundingBox_Contains(t *testing.T) {
	tests := []struct {
		name     string
		bb       *BoundingBox
		lat, lng float64
		contains bool
	}{
		// 通常のケース
		{
			name:     "中心点を含む",
			bb:       NewBoundingBox(35.0, 139.0, 36.0, 140.0),
			lat:      35.5,
			lng:      139.5,
			contains: true,
		},
		{
			name:     "南西角を含む",
			bb:       NewBoundingBox(35.0, 139.0, 36.0, 140.0),
			lat:      35.0,
			lng:      139.0,
			contains: true,
		},
		{
			name:     "北東角を含む",
			bb:       NewBoundingBox(35.0, 139.0, 36.0, 140.0),
			lat:      36.0,
			lng:      140.0,
			contains: true,
		},
		{
			name:     "北にある座標を含まない",
			bb:       NewBoundingBox(35.0, 139.0, 36.0, 140.0),
			lat:      37.0,
			lng:      139.5,
			contains: false,
		},
		{
			name:     "南にある座標を含まない",
			bb:       NewBoundingBox(35.0, 139.0, 36.0, 140.0),
			lat:      34.0,
			lng:      139.5,
			contains: false,
		},
		{
			name:     "東にある座標を含まない",
			bb:       NewBoundingBox(35.0, 139.0, 36.0, 140.0),
			lat:      35.5,
			lng:      141.0,
			contains: false,
		},
		{
			name:     "西にある座標を含まない",
			bb:       NewBoundingBox(35.0, 139.0, 36.0, 140.0),
			lat:      35.5,
			lng:      138.0,
			contains: false,
		},
		// 日付変更線をまたぐケース
		{
			name:     "日付変更線をまたぐ場合: 東側を含む",
			bb:       NewBoundingBox(35.0, 170.0, 36.0, -170.0),
			lat:      35.5,
			lng:      175.0,
			contains: true,
		},
		{
			name:     "日付変更線をまたぐ場合: 西側を含む",
			bb:       NewBoundingBox(35.0, 170.0, 36.0, -170.0),
			lat:      35.5,
			lng:      -175.0,
			contains: true,
		},
		{
			name:     "日付変更線をまたぐ場合: 中間を含まない",
			bb:       NewBoundingBox(35.0, 170.0, 36.0, -170.0),
			lat:      35.5,
			lng:      0.0,
			contains: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bb.Contains(tt.lat, tt.lng)
			if got != tt.contains {
				t.Errorf("Contains(%f, %f) = %v, 期待値 %v", tt.lat, tt.lng, got, tt.contains)
			}
		})
	}
}

// TestZoomToResolution はZoomToResolutionがズームレベルから正しい解像度を返すことをテストする
func TestZoomToResolution(t *testing.T) {
	tests := []struct {
		name       string
		zoom       float64
		resolution entity.Resolution
	}{
		// zoom < 6 -> Res3
		{"zoom 0はRes3", 0.0, entity.Res3},
		{"zoom 1はRes3", 1.0, entity.Res3},
		{"zoom 5はRes3", 5.0, entity.Res3},
		{"zoom 5.9はRes3", 5.9, entity.Res3},
		// 6 <= zoom < 10 -> Res5
		{"zoom 6はRes5", 6.0, entity.Res5},
		{"zoom 7.5はRes5", 7.5, entity.Res5},
		{"zoom 9.9はRes5", 9.9, entity.Res5},
		// 10 <= zoom < 14 -> Res7
		{"zoom 10はRes7", 10.0, entity.Res7},
		{"zoom 12はRes7", 12.0, entity.Res7},
		{"zoom 13.9はRes7", 13.9, entity.Res7},
		// zoom >= 14 -> Res9
		{"zoom 14はRes9", 14.0, entity.Res9},
		{"zoom 18はRes9", 18.0, entity.Res9},
		{"zoom 22はRes9", 22.0, entity.Res9},
		// 境界値
		{"zoom 5.999999はRes3", 5.999999, entity.Res3},
		{"zoom 9.999999はRes5", 9.999999, entity.Res5},
		{"zoom 13.999999はRes7", 13.999999, entity.Res7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ZoomToResolution(tt.zoom)
			if got != tt.resolution {
				t.Errorf("ZoomToResolution(%f) = %v, 期待値 %v", tt.zoom, got, tt.resolution)
			}
		})
	}
}

// TestCellToLatLng はCellToLatLngがH3インデックスから座標を正しく取得することをテストする
func TestCellToLatLng(t *testing.T) {
	tests := []struct {
		name      string
		h3Index   string
		expectErr bool
	}{
		// 正常系: 有効なH3インデックス
		{
			name:      "有効なH3インデックス(res7)",
			h3Index:   "871f1a4adffffff",
			expectErr: false,
		},
		{
			name:      "有効なH3インデックス(res9)",
			h3Index:   "891f1a4ad83ffff",
			expectErr: false,
		},
		// 異常系: 無効なH3インデックス
		{
			name:      "空文字列",
			h3Index:   "",
			expectErr: true,
		},
		{
			name:      "無効な文字列",
			h3Index:   "invalid",
			expectErr: true,
		},
		{
			name:      "不正な16進数",
			h3Index:   "gggggggggggggg",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lng, err := CellToLatLng(tt.h3Index)

			if tt.expectErr {
				if err == nil {
					t.Error("エラーを期待しましたがnilが返されました")
				}
				return
			}

			if err != nil {
				t.Errorf("予期しないエラー: %v", err)
				return
			}

			// 緯度の範囲チェック
			if lat < -90 || lat > 90 {
				t.Errorf("緯度が範囲外です: %f", lat)
			}

			// 経度の範囲チェック
			if lng < -180 || lng > 180 {
				t.Errorf("経度が範囲外です: %f", lng)
			}
		})
	}
}

// TestFilterCellsInBBox はFilterCellsInBBoxがBoundingBox内のセルを正しくフィルタリングすることをテストする
func TestFilterCellsInBBox(t *testing.T) {
	// 東京周辺のH3インデックス(res7)
	tokyoCell := "871f1a4adffffff"

	tests := []struct {
		name      string
		h3Indexes []string
		bb        *BoundingBox
		wantLen   int
	}{
		{
			name:      "bbがnilの場合は全てのセルを返す",
			h3Indexes: []string{tokyoCell},
			bb:        nil,
			wantLen:   1,
		},
		{
			name:      "空のスライスの場合は空を返す",
			h3Indexes: []string{},
			bb:        NewBoundingBox(35.0, 139.0, 36.0, 140.0),
			wantLen:   0,
		},
		{
			name:      "無効なH3インデックスはスキップされる",
			h3Indexes: []string{"invalid", "also-invalid"},
			bb:        NewBoundingBox(35.0, 139.0, 36.0, 140.0),
			wantLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FilterCellsInBBox(tt.h3Indexes, tt.bb)

			if err != nil {
				t.Errorf("予期しないエラー: %v", err)
				return
			}

			if len(result) != tt.wantLen {
				t.Errorf("結果の長さ = %d, 期待値 %d", len(result), tt.wantLen)
			}
		})
	}
}

// TestCalculateCenterFromH3 はCalculateCenterFromH3がH3インデックスから中心座標を正しく計算することをテストする
func TestCalculateCenterFromH3(t *testing.T) {
	tests := []struct {
		name      string
		h3Index   string
		expectErr bool
	}{
		{
			name:      "有効なH3インデックス",
			h3Index:   "871f1a4adffffff",
			expectErr: false,
		},
		{
			name:      "無効なH3インデックス",
			h3Index:   "invalid",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lng, err := CalculateCenterFromH3(tt.h3Index)

			if tt.expectErr {
				if err == nil {
					t.Error("エラーを期待しましたがnilが返されました")
				}
				return
			}

			if err != nil {
				t.Errorf("予期しないエラー: %v", err)
				return
			}

			// 緯度の範囲チェック
			if lat < -90 || lat > 90 {
				t.Errorf("緯度が範囲外です: %f", lat)
			}

			// 経度の範囲チェック
			if lng < -180 || lng > 180 {
				t.Errorf("経度が範囲外です: %f", lng)
			}
		})
	}
}

// TestIsValidH3Index はIsValidH3IndexがH3インデックスの有効性を正しく判定することをテストする
func TestIsValidH3Index(t *testing.T) {
	tests := []struct {
		name    string
		h3Index string
		valid   bool
	}{
		{"有効なH3インデックス(res7)", "871f1a4adffffff", true},
		{"有効なH3インデックス(res3)", "831f8dfffffffff", true},
		{"有効なH3インデックス(res9)", "891f1a4ad83ffff", true},
		{"空文字列", "", false},
		{"無効な文字列", "invalid", false},
		{"不正な16進数", "gggggggggggggg", false},
		{"短すぎる文字列", "871f1a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidH3Index(tt.h3Index)
			if got != tt.valid {
				t.Errorf("IsValidH3Index(%q) = %v, 期待値 %v", tt.h3Index, got, tt.valid)
			}
		})
	}
}
