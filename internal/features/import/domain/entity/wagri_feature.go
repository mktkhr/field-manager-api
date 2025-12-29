package entity

import (
	"encoding/json"
	"time"
)

// WagriResponse はwagri APIのレスポンス全体を表す
type WagriResponse struct {
	TargetFeatures []WagriFeature `json:"targetFeatures"`
}

// WagriFeature はwagri APIの1つのFeatureを表す
type WagriFeature struct {
	Geometry   WagriGeometry   `json:"geometry"`
	Type       string          `json:"type"`
	Properties WagriProperties `json:"properties"`
}

// WagriGeometry はwagri APIのジオメトリを表す
type WagriGeometry struct {
	Coordinates [][][]float64 `json:"coordinates"`
	Type        string        `json:"type"` // LinearPolygon
}

// WagriProperties はwagri APIのプロパティを表す
type WagriProperties struct {
	ID                  string         `json:"ID"`
	CityCode            string         `json:"CityCode"`
	IssueYear           string         `json:"IssueYear"`
	EditYear            string         `json:"EditYear"`
	PointLat            float64        `json:"PointLat"`
	PointLng            float64        `json:"PointLng"`
	FieldType           string         `json:"FieldType"`
	Number              int            `json:"Number"`
	SoilLargeCode       string         `json:"SoilLargeCode"`
	SoilMiddleCode      string         `json:"SoilMiddleCode"`
	SoilSmallCode       string         `json:"SoilSmallCode"`
	SoilSmallName       string         `json:"SoilSmallName"`
	History             string         `json:"History"` // JSON文字列
	LastPolygonUuid     string         `json:"LastPolygonUuid"`
	PrevLastPolygonUuid *string        `json:"PrevLastPolygonUuid"`
	PinInfo             []WagriPinInfo `json:"PinInfo"`
}

// WagriPinInfo はwagri APIのPinInfo(農地台帳情報)を表す
type WagriPinInfo struct {
	FarmerNumber                   string  `json:"FarmerNumber"`
	Address                        string  `json:"Address"`
	Area                           int     `json:"Area"`
	LandCategoryCode               string  `json:"LandCategoryCode"`
	LandCategory                   string  `json:"LandCategory"`
	IsIdleAgriculturalLandCode     string  `json:"IsIdleAgriculturalLandCode"`
	IsIdleAgriculturalLand         string  `json:"IsIdleAgriculturalLand"`
	DescriptiveStudyData           *string `json:"DescriptiveStudyData"` // "2016-08-31" 形式
	AgricultureCommitteeName       string  `json:"AgricultureCommitteeName"`
	RightClassificationCode        string  `json:"RightClassificationCode"`
	RightClassification            string  `json:"RightClassification"`
	FarmlandManagementStatusCode   string  `json:"FarmlandManagementStatusCode"`
	FarmlandManagementStatus       string  `json:"FarmlandManagementStatus"`
	OwnerAssuranceStatusCode       string  `json:"OwnerAssuranceStatusCode"`
	OwnerAssuranceStatus           string  `json:"OwnerAssuranceStatus"`
	IntentionOwnerAgriLandCode     string  `json:"IntentionOwnerAgriLandCode"`
	IntentionOwnerAgriLand         string  `json:"IntentionOwnerAgriLand"`
	IntentionOwnerIdleAgriLandCode string  `json:"IntentionOwnerIdleAgriLandCode"`
	IntentionOwnerIdleAgriLand     string  `json:"IntentionOwnerIdleAgriLand"`
	CityPlanningActClassCode       string  `json:"CityPlanningActClassCode"`
	CityPlanningActClass           string  `json:"CityPlanningActClass"`
	AgriVibrationMethodClassCode   string  `json:"AgriVibrationMethodClassCode"`
	AgriVibrationMethodClass       string  `json:"AgriVibrationMethodClass"`
	StartDuration                  *string `json:"StartDuration"`
	EndDuration                    *string `json:"EndDuration"`
	UseIntentionSurveyData         *string `json:"UseIntentionSurveyData"`
	OwnerAssurancePublicNoticeDate *string `json:"OwnerAssurancePublicNoticeDate"`
	MeasuresDate                   *string `json:"MeasuresDate"`
	MeasuresPublicNoticeDate       *string `json:"MeasuresPublicNoticeDate"`
	FarmlandRecommendedDate        *string `json:"FarmlandRecommendedDate"`
	FarmlandArbitrationDate        *string `json:"FarmlandArbitrationDate"`
}

// ParseDescriptiveStudyData はDescriptiveStudyDataをtime.Timeにパースする
func (p *WagriPinInfo) ParseDescriptiveStudyData() *time.Time {
	if p.DescriptiveStudyData == nil || *p.DescriptiveStudyData == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *p.DescriptiveStudyData)
	if err != nil {
		return nil
	}
	return &t
}

// ParseWagriResponse はwagri APIレスポンスJSONをパースする
func ParseWagriResponse(data []byte) (*WagriResponse, error) {
	var response WagriResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// HasSoilType は土壌タイプ情報があるかどうかを判定する
func (p *WagriProperties) HasSoilType() bool {
	return p.SoilSmallCode != ""
}

// HasPinInfo はPinInfo(農地台帳情報)があるかどうかを判定する
func (f *WagriFeature) HasPinInfo() bool {
	return len(f.Properties.PinInfo) > 0
}
