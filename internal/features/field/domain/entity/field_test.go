package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestNewField はNewFieldが正しいID、CityCode、デフォルト名称、タイムスタンプを持つFieldを生成することをテストする
func TestNewField(t *testing.T) {
	id := uuid.New()
	cityCode := "163210"

	field := NewField(id, cityCode)

	if field.ID != id {
		t.Errorf("ID = %v, 期待値 %v", field.ID, id)
	}
	if field.CityCode != cityCode {
		t.Errorf("CityCode = %q, 期待値 %q", field.CityCode, cityCode)
	}
	if field.Name != "名称不明" {
		t.Errorf("Name = %q, 期待値 %q", field.Name, "名称不明")
	}
	if field.CreatedAt.IsZero() {
		t.Error("CreatedAtがゼロ値です")
	}
	if field.UpdatedAt.IsZero() {
		t.Error("UpdatedAtがゼロ値です")
	}
}

// TestFieldCalculateH3Indexes はCalculateH3Indexesが各解像度(3,5,7,9)のH3インデックスを正しく計算することをテストする
func TestFieldCalculateH3Indexes(t *testing.T) {
	field := &Field{}

	// 東京駅付近の座標
	lat := 35.6812
	lng := 139.7671

	err := field.CalculateH3Indexes(lat, lng)
	if err != nil {
		t.Fatalf("CalculateH3Indexesでエラーが発生: %v", err)
	}

	if field.H3IndexRes3 == nil {
		t.Error("H3IndexRes3がnilです")
	}
	if field.H3IndexRes5 == nil {
		t.Error("H3IndexRes5がnilです")
	}
	if field.H3IndexRes7 == nil {
		t.Error("H3IndexRes7がnilです")
	}
	if field.H3IndexRes9 == nil {
		t.Error("H3IndexRes9がnilです")
	}

	// H3インデックスが空でないことを確認
	if *field.H3IndexRes3 == "" {
		t.Error("H3IndexRes3が空です")
	}
	if *field.H3IndexRes5 == "" {
		t.Error("H3IndexRes5が空です")
	}
	if *field.H3IndexRes7 == "" {
		t.Error("H3IndexRes7が空です")
	}
	if *field.H3IndexRes9 == "" {
		t.Error("H3IndexRes9が空です")
	}
}

// TestFieldSetSoilType はSetSoilTypeがSoilTypeIDを正しく設定することをテストする
func TestFieldSetSoilType(t *testing.T) {
	field := &Field{}
	soilTypeID := uuid.New()

	field.SetSoilType(soilTypeID)

	if field.SoilTypeID == nil {
		t.Error("SoilTypeIDがnilです")
	}
	if *field.SoilTypeID != soilTypeID {
		t.Errorf("SoilTypeID = %v, 期待値 %v", *field.SoilTypeID, soilTypeID)
	}
}

// TestSetGeometry はSetGeometryが有効なPolygonでGeometry、Centroid、H3インデックスを設定し、nilの場合はnilを設定することをテストする
func TestSetGeometry(t *testing.T) {
	t.Run("set valid polygon", func(t *testing.T) {
		field := NewField(uuid.New(), "163210")

		coords := [][]float64{
			{139.0, 35.0},
			{139.1, 35.0},
			{139.05, 35.1},
		}
		polygon, err := ConvertLinearPolygonToPolygon(coords)
		require.NoError(t, err, "ConvertLinearPolygonToPolygonでエラーが発生")

		err = field.SetGeometry(polygon)
		require.NoError(t, err, "SetGeometryでエラーが発生")

		if field.Geometry == nil {
			t.Error("Geometryがnilです")
		}
		if field.Centroid == nil {
			t.Error("Centroidがnilです")
		}
		if field.H3IndexRes3 == nil {
			t.Error("H3IndexRes3がnilです")
		}
		if field.H3IndexRes5 == nil {
			t.Error("H3IndexRes5がnilです")
		}
		if field.H3IndexRes7 == nil {
			t.Error("H3IndexRes7がnilです")
		}
		if field.H3IndexRes9 == nil {
			t.Error("H3IndexRes9がnilです")
		}
	})

	t.Run("set nil polygon", func(t *testing.T) {
		field := NewField(uuid.New(), "163210")

		err := field.SetGeometry(nil)
		require.NoError(t, err, "SetGeometry(nil)でエラーが発生")

		if field.Geometry != nil {
			t.Error("Geometryがnilではありません")
		}
		if field.Centroid != nil {
			t.Error("Centroidがnilではありません")
		}
	})
}

// TestCalculateCentroid はCalculateCentroidがnilの場合はnilを、有効なPolygonの場合は正しい重心座標を返すことをテストする
func TestCalculateCentroid(t *testing.T) {
	t.Run("nil polygon", func(t *testing.T) {
		result := CalculateCentroid(nil)
		if result != nil {
			t.Errorf("CalculateCentroid() = %v, 期待値 nil", result)
		}
	})

	t.Run("valid polygon", func(t *testing.T) {
		coords := [][]float64{
			{139.0, 35.0},
			{140.0, 35.0},
			{140.0, 36.0},
			{139.0, 36.0},
		}
		polygon, err := ConvertLinearPolygonToPolygon(coords)
		require.NoError(t, err, "ConvertLinearPolygonToPolygonでエラーが発生")

		result := CalculateCentroid(polygon)

		if result == nil {
			t.Error("CalculateCentroid()がnilです、非nilを期待")
			return
		}

		// 重心は正方形の中心付近にあるはず
		x := result.X()
		y := result.Y()
		// 139.0-140.0の中心は139.5, 35.0-36.0の中心は35.5付近
		if x < 139.0 || x > 140.0 {
			t.Errorf("Centroid X = %v, 期待範囲 139.0-140.0", x)
		}
		if y < 35.0 || y > 36.0 {
			t.Errorf("Centroid Y = %v, 期待範囲 35.0-36.0", y)
		}
	})
}

// TestConvertLinearPolygonToPolygon はConvertLinearPolygonToPolygonが空座標、有効な三角形、閉じたポリゴンを正しく処理することをテストする
func TestConvertLinearPolygonToPolygon(t *testing.T) {
	tests := []struct {
		name        string
		coordinates [][]float64
		wantNil     bool
		wantErr     bool
	}{
		// 正常系: 空の座標はnilを返す
		{
			name:        "empty coordinates",
			coordinates: [][]float64{},
			wantNil:     true,
			wantErr:     false,
		},
		// 正常系: 有効な三角形(3点)はポリゴンを返す
		{
			name: "valid triangle",
			coordinates: [][]float64{
				{139.0, 35.0},
				{139.1, 35.0},
				{139.05, 35.1},
			},
			wantNil: false,
			wantErr: false,
		},
		// 正常系: 閉じたポリゴンも正常に処理する
		{
			name: "already closed polygon",
			coordinates: [][]float64{
				{139.0, 35.0},
				{139.1, 35.0},
				{139.05, 35.1},
				{139.0, 35.0},
			},
			wantNil: false,
			wantErr: false,
		},
		// 異常系: 2点のみはエラーを返す
		{
			name: "insufficient points - 2 points",
			coordinates: [][]float64{
				{139.0, 35.0},
				{139.1, 35.0},
			},
			wantNil: false,
			wantErr: true,
		},
		// 異常系: 1点のみはエラーを返す
		{
			name: "insufficient points - 1 point",
			coordinates: [][]float64{
				{139.0, 35.0},
			},
			wantNil: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			polygon, err := ConvertLinearPolygonToPolygon(tt.coordinates)

			if tt.wantErr {
				if err == nil {
					t.Error("ConvertLinearPolygonToPolygon()でエラーを期待したがnilが返された")
				}
				return
			}

			if err != nil {
				t.Errorf("ConvertLinearPolygonToPolygon()でエラー発生 = %v", err)
				return
			}

			if tt.wantNil {
				if polygon != nil {
					t.Errorf("ConvertLinearPolygonToPolygon() = %v, 期待値 nil", polygon)
				}
			} else {
				if polygon == nil {
					t.Error("ConvertLinearPolygonToPolygon()がnilです、非nilを期待")
				}
			}
		})
	}
}

// TestCoordsEqual はcoordsEqualが同一座標、異なる座標、不完全なスライスに対して正しい判定結果を返すことをテストする
func TestCoordsEqual(t *testing.T) {
	tests := []struct {
		name string
		a    []float64
		b    []float64
		want bool
	}{
		{
			name: "equal coords",
			a:    []float64{139.0, 35.0},
			b:    []float64{139.0, 35.0},
			want: true,
		},
		{
			name: "different coords",
			a:    []float64{139.0, 35.0},
			b:    []float64{139.1, 35.0},
			want: false,
		},
		{
			name: "short slice a",
			a:    []float64{139.0},
			b:    []float64{139.0, 35.0},
			want: false,
		},
		{
			name: "short slice b",
			a:    []float64{139.0, 35.0},
			b:    []float64{139.0},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// geom.Coordは[]float64の型エイリアス
			if got := coordsEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("coordsEqual() = %v, 期待値 %v", got, tt.want)
			}
		})
	}
}
