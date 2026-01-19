package h3util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mktkhr/field-manager-api/internal/features/fieldsearch/domain/entity"
)

type BBoxCellCalculatorTestSuite struct {
	suite.Suite
	calculator *BBoxCellCalculator
}

func (s *BBoxCellCalculatorTestSuite) SetupTest() {
	s.calculator = NewBBoxCellCalculator()
}

func TestBBoxCellCalculatorTestSuite(t *testing.T) {
	suite.Run(t, new(BBoxCellCalculatorTestSuite))
}

// TestNewBBoxCellCalculator_Success はNewBBoxCellCalculatorが正しくインスタンスを生成することをテスト
func (s *BBoxCellCalculatorTestSuite) TestNewBBoxCellCalculator_Success() {
	calc := NewBBoxCellCalculator()
	require.NotNil(s.T(), calc, "BBoxCellCalculatorがnilです")
}

// TestCalculateCells_Success_SmallArea は小さな範囲でH3セルが正しく計算されることをテスト
func (s *BBoxCellCalculatorTestSuite) TestCalculateCells_Success_SmallArea() {
	// 東京周辺の小さな範囲(約1km四方)
	bbox, err := entity.NewBoundingBox(35.6800, 139.7000, 35.6900, 139.7100)
	require.NoError(s.T(), err, "BoundingBox作成に失敗")

	// resolution=9で計算
	cells, err := s.calculator.CalculateCells(bbox, 9)

	require.NoError(s.T(), err, "H3セル計算でエラーが発生")
	require.NotEmpty(s.T(), cells, "H3セルが空です")

	// 小さな範囲なのでセル数は比較的少ないはず
	require.Less(s.T(), len(cells), 1000, "H3セル数が多すぎます")
}

// TestCalculateCells_Success_MediumArea は中程度の範囲でH3セルが正しく計算されることをテスト
func (s *BBoxCellCalculatorTestSuite) TestCalculateCells_Success_MediumArea() {
	// 東京周辺の中程度の範囲(約10km四方)
	bbox, err := entity.NewBoundingBox(35.6500, 139.6500, 35.7500, 139.7500)
	require.NoError(s.T(), err, "BoundingBox作成に失敗")

	// resolution=7で計算
	cells, err := s.calculator.CalculateCells(bbox, 7)

	require.NoError(s.T(), err, "H3セル計算でエラーが発生")
	require.NotEmpty(s.T(), cells, "H3セルが空です")
}

// TestCalculateCells_Success_LargeArea は大きな範囲でH3セルが正しく計算されることをテスト
func (s *BBoxCellCalculatorTestSuite) TestCalculateCells_Success_LargeArea() {
	// 関東圏程度の範囲(約100km四方)
	bbox, err := entity.NewBoundingBox(35.0, 139.0, 36.0, 140.0)
	require.NoError(s.T(), err, "BoundingBox作成に失敗")

	// resolution=5で計算
	cells, err := s.calculator.CalculateCells(bbox, 5)

	require.NoError(s.T(), err, "H3セル計算でエラーが発生")
	require.NotEmpty(s.T(), cells, "H3セルが空です")
}

// TestCalculateCells_Success_Resolution3 は解像度3でH3セルが正しく計算されることをテスト
func (s *BBoxCellCalculatorTestSuite) TestCalculateCells_Success_Resolution3() {
	// 広域(日本全国レベル)
	bbox, err := entity.NewBoundingBox(33.0, 130.0, 40.0, 145.0)
	require.NoError(s.T(), err, "BoundingBox作成に失敗")

	// resolution=3で計算
	cells, err := s.calculator.CalculateCells(bbox, 3)

	require.NoError(s.T(), err, "H3セル計算でエラーが発生")
	require.NotEmpty(s.T(), cells, "H3セルが空です")

	// 解像度3は広い範囲をカバーするのでセル数は比較的少ないはず
	require.Less(s.T(), len(cells), 500, "解像度3のH3セル数が多すぎます")
}

// TestCalculateCells_Success_CellsContainValidH3Index は返されたセルが有効なH3インデックス形式であることをテスト
func (s *BBoxCellCalculatorTestSuite) TestCalculateCells_Success_CellsContainValidH3Index() {
	bbox, err := entity.NewBoundingBox(35.6800, 139.7000, 35.6900, 139.7100)
	require.NoError(s.T(), err, "BoundingBox作成に失敗")

	cells, err := s.calculator.CalculateCells(bbox, 9)
	require.NoError(s.T(), err, "H3セル計算でエラーが発生")
	require.NotEmpty(s.T(), cells, "H3セルが空です")

	// H3インデックスは15文字の16進数文字列
	for _, cell := range cells {
		require.Len(s.T(), cell, 15, "H3インデックスの長さが15文字ではありません: %s", cell)
	}
}

// TestCalculateCells_ValidationError_NilBBox はBoundingBoxがnilの場合にエラーを返すことをテスト
func (s *BBoxCellCalculatorTestSuite) TestCalculateCells_ValidationError_NilBBox() {
	cells, err := s.calculator.CalculateCells(nil, 9)

	require.Error(s.T(), err, "nilのBBoxでエラーが発生すべき")
	require.Nil(s.T(), cells, "セルがnilではありません")
	require.Contains(s.T(), err.Error(), "BoundingBoxがnil")
}

// TestCalculateCells_ValidationError_InvalidResolutionNegative は負の解像度でエラーを返すことをテスト
func (s *BBoxCellCalculatorTestSuite) TestCalculateCells_ValidationError_InvalidResolutionNegative() {
	bbox, err := entity.NewBoundingBox(35.6800, 139.7000, 35.6900, 139.7100)
	require.NoError(s.T(), err, "BoundingBox作成に失敗")

	cells, err := s.calculator.CalculateCells(bbox, -1)

	require.Error(s.T(), err, "負の解像度でエラーが発生すべき")
	require.Nil(s.T(), cells, "セルがnilではありません")
}

// TestCalculateCells_ValidationError_InvalidResolutionTooHigh は16以上の解像度でエラーを返すことをテスト
func (s *BBoxCellCalculatorTestSuite) TestCalculateCells_ValidationError_InvalidResolutionTooHigh() {
	bbox, err := entity.NewBoundingBox(35.6800, 139.7000, 35.6900, 139.7100)
	require.NoError(s.T(), err, "BoundingBox作成に失敗")

	cells, err := s.calculator.CalculateCells(bbox, 16)

	require.Error(s.T(), err, "16以上の解像度でエラーが発生すべき")
	require.Nil(s.T(), cells, "セルがnilではありません")
}

// TestCalculateCells_BoundaryValue_Resolution0 は解像度0でも正常に動作することをテスト
func (s *BBoxCellCalculatorTestSuite) TestCalculateCells_BoundaryValue_Resolution0() {
	// 解像度0は非常に広い範囲をカバーするため、大きなBBoxを使用
	bbox, err := entity.NewBoundingBox(-60.0, -180.0, 60.0, 180.0)
	require.NoError(s.T(), err, "BoundingBox作成に失敗")

	cells, err := s.calculator.CalculateCells(bbox, 0)

	require.NoError(s.T(), err, "解像度0でエラーが発生")
	// 解像度0でも計算自体はエラーにならない
	// 結果が空でも許容(ContainmentCenterの挙動による)
	require.NotNil(s.T(), cells, "セルがnilです")
}

// TestCalculateCells_BoundaryValue_Resolution15 は解像度15でも正常に動作することをテスト
func (s *BBoxCellCalculatorTestSuite) TestCalculateCells_BoundaryValue_Resolution15() {
	// 非常に小さな範囲
	bbox, err := entity.NewBoundingBox(35.6812, 139.7671, 35.6813, 139.7672)
	require.NoError(s.T(), err, "BoundingBox作成に失敗")

	cells, err := s.calculator.CalculateCells(bbox, 15)

	require.NoError(s.T(), err, "解像度15でエラーが発生")
	require.NotEmpty(s.T(), cells, "H3セルが空です")
}

// TestCalculateCells_Success_VerySmallBBox はH3セルより小さいBBoxでもセルが返されることをテスト
// PolygonToCellsExperimentalが空を返す場合のフォールバックロジックを検証
func (s *BBoxCellCalculatorTestSuite) TestCalculateCells_Success_VerySmallBBox() {
	// 非常に小さな範囲(約200m四方) - H3 res9セル(約0.1km²)より小さい
	// 新潟県の圃場データが存在する範囲
	bbox, err := entity.NewBoundingBox(37.084, 138.260, 37.086, 138.263)
	require.NoError(s.T(), err, "BoundingBox作成に失敗")

	// resolution=9で計算(この範囲ではPolygonToCellsExperimentalが空を返す可能性がある)
	cells, err := s.calculator.CalculateCells(bbox, 9)

	require.NoError(s.T(), err, "H3セル計算でエラーが発生")
	require.NotEmpty(s.T(), cells, "小さなBBoxでもH3セルが取得できるべき(フォールバックロジック)")

	// 小さな範囲なので、4隅+中心から計算しても重複排除後は1-5セル程度のはず
	require.LessOrEqual(s.T(), len(cells), 5, "小さなBBoxのH3セル数が多すぎます")
}

// TestCalculateCells_Success_TinyBBox は極小BBoxでもセルが返されることをテスト
func (s *BBoxCellCalculatorTestSuite) TestCalculateCells_Success_TinyBBox() {
	// 極小範囲(約10m四方)
	bbox, err := entity.NewBoundingBox(35.68000, 139.70000, 35.68010, 139.70010)
	require.NoError(s.T(), err, "BoundingBox作成に失敗")

	cells, err := s.calculator.CalculateCells(bbox, 9)

	require.NoError(s.T(), err, "H3セル計算でエラーが発生")
	require.NotEmpty(s.T(), cells, "極小BBoxでもH3セルが取得できるべき")

	// 極小範囲なので、すべての点が同じセルに入る可能性が高い
	require.LessOrEqual(s.T(), len(cells), 2, "極小BBoxのH3セル数は1-2であるべき")
}
