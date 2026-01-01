package dto

import "time"

// FieldBatchInput はバッチUPSERT用の入力データ
// Consumer側(import機能)のDomain層で型を定義し、Provider側(field機能)が実装する
type FieldBatchInput struct {
	ID          string
	CityCode    string
	Geometry    FieldBatchGeometry
	SoilType    *FieldBatchSoilType
	PinInfoList []FieldBatchPinInfo
}

// FieldBatchGeometry はバッチUPSERT用のジオメトリデータ
type FieldBatchGeometry struct {
	// Coordinates は [[[lng, lat], [lng, lat], ...]] 形式の座標
	// LinearPolygon形式(wagri)またはPolygon形式
	Coordinates [][][]float64
	Type        string
}

// FieldBatchSoilType はバッチUPSERT用の土壌タイプ情報
type FieldBatchSoilType struct {
	LargeCode  string
	MiddleCode string
	SmallCode  string
	SmallName  string
}

// FieldBatchPinInfo はバッチUPSERT用の農地台帳情報
type FieldBatchPinInfo struct {
	FarmerNumber            string
	Address                 string
	Area                    int
	LandCategoryCode        string
	LandCategory            string
	IdleLandStatusCode      string
	IdleLandStatus          string
	DescriptiveStudyData    *time.Time
	DescriptiveStudyDataRaw *string // "2006-01-02" 形式の文字列
}

// HasSoilType は土壌タイプ情報があるかどうかを判定する
func (f *FieldBatchInput) HasSoilType() bool {
	return f.SoilType != nil && f.SoilType.SmallCode != ""
}

// HasPinInfo は農地台帳情報があるかどうかを判定する
func (f *FieldBatchInput) HasPinInfo() bool {
	return len(f.PinInfoList) > 0
}

// GetFirstCoordinates はジオメトリの最初の座標配列を取得する
func (f *FieldBatchInput) GetFirstCoordinates() [][]float64 {
	if len(f.Geometry.Coordinates) > 0 {
		return f.Geometry.Coordinates[0]
	}
	return nil
}

// ParseDescriptiveStudyData は DescriptiveStudyDataRaw をパースして time.Time を返す
func (p *FieldBatchPinInfo) ParseDescriptiveStudyData() *time.Time {
	if p.DescriptiveStudyData != nil {
		return p.DescriptiveStudyData
	}
	if p.DescriptiveStudyDataRaw == nil || *p.DescriptiveStudyDataRaw == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *p.DescriptiveStudyDataRaw)
	if err != nil {
		return nil
	}
	return &t
}
