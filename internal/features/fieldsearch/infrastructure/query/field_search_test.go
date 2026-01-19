package query

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FieldSearchQueryTestSuite struct {
	suite.Suite
}

func TestFieldSearchQueryTestSuite(t *testing.T) {
	suite.Run(t, new(FieldSearchQueryTestSuite))
}

// TestGetH3ColumnName_Success_Res3 は解像度3でh3_index_res3が返されることをテスト
func (s *FieldSearchQueryTestSuite) TestGetH3ColumnName_Success_Res3() {
	q := &fieldSearchQuery{}
	column, err := q.getH3ColumnName(3)

	require.NoError(s.T(), err, "解像度3でエラーが発生")
	require.Equal(s.T(), "h3_index_res3", column, "カラム名が一致しません")
}

// TestGetH3ColumnName_Success_Res5 は解像度5でh3_index_res5が返されることをテスト
func (s *FieldSearchQueryTestSuite) TestGetH3ColumnName_Success_Res5() {
	q := &fieldSearchQuery{}
	column, err := q.getH3ColumnName(5)

	require.NoError(s.T(), err, "解像度5でエラーが発生")
	require.Equal(s.T(), "h3_index_res5", column, "カラム名が一致しません")
}

// TestGetH3ColumnName_Success_Res7 は解像度7でh3_index_res7が返されることをテスト
func (s *FieldSearchQueryTestSuite) TestGetH3ColumnName_Success_Res7() {
	q := &fieldSearchQuery{}
	column, err := q.getH3ColumnName(7)

	require.NoError(s.T(), err, "解像度7でエラーが発生")
	require.Equal(s.T(), "h3_index_res7", column, "カラム名が一致しません")
}

// TestGetH3ColumnName_Success_Res9 は解像度9でh3_index_res9が返されることをテスト
func (s *FieldSearchQueryTestSuite) TestGetH3ColumnName_Success_Res9() {
	q := &fieldSearchQuery{}
	column, err := q.getH3ColumnName(9)

	require.NoError(s.T(), err, "解像度9でエラーが発生")
	require.Equal(s.T(), "h3_index_res9", column, "カラム名が一致しません")
}

// TestGetH3ColumnName_ValidationError_UnsupportedResolution はサポートされていない解像度でエラーを返すことをテスト
func (s *FieldSearchQueryTestSuite) TestGetH3ColumnName_ValidationError_UnsupportedResolution() {
	q := &fieldSearchQuery{}

	unsupportedResolutions := []int{0, 1, 2, 4, 6, 8, 10, 11, 12, 13, 14, 15}
	for _, res := range unsupportedResolutions {
		column, err := q.getH3ColumnName(res)
		require.Error(s.T(), err, "サポートされていない解像度%dでエラーが発生すべき", res)
		require.Empty(s.T(), column, "カラム名が空ではありません")
		require.Contains(s.T(), err.Error(), "サポートされていないH3解像度")
	}
}
