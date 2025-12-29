package entity

import (
	"testing"
)

func TestParseWagriResponse(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantLen int
		wantErr bool
	}{
		{
			name:    "valid empty response",
			data:    []byte(`{"targetFeatures":[]}`),
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "valid response with one feature",
			data: []byte(`{
				"targetFeatures": [{
					"type": "Feature",
					"geometry": {
						"type": "LinearPolygon",
						"coordinates": [[[139.0, 35.0], [139.1, 35.0], [139.05, 35.1]]]
					},
					"properties": {
						"ID": "test-id-123",
						"CityCode": "163210",
						"IssueYear": "2024",
						"EditYear": "2024",
						"PointLat": 35.05,
						"PointLng": 139.05,
						"FieldType": "1",
						"Number": 1,
						"SoilLargeCode": "A",
						"SoilMiddleCode": "A1",
						"SoilSmallCode": "A1a",
						"SoilSmallName": "黒ボク土",
						"History": "{}",
						"LastPolygonUuid": "uuid-123",
						"PinInfo": []
					}
				}]
			}`),
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "invalid json",
			data:    []byte(`{invalid`),
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := ParseWagriResponse(tt.data)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseWagriResponse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseWagriResponse() error = %v", err)
				return
			}

			if len(response.TargetFeatures) != tt.wantLen {
				t.Errorf("len(TargetFeatures) = %d, want %d", len(response.TargetFeatures), tt.wantLen)
			}
		})
	}
}

func TestWagriPropertiesHasSoilType(t *testing.T) {
	tests := []struct {
		name          string
		soilSmallCode string
		want          bool
	}{
		{
			name:          "has soil type",
			soilSmallCode: "A1a",
			want:          true,
		},
		{
			name:          "no soil type",
			soilSmallCode: "",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := &WagriProperties{
				SoilSmallCode: tt.soilSmallCode,
			}
			if got := props.HasSoilType(); got != tt.want {
				t.Errorf("HasSoilType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWagriFeatureHasPinInfo(t *testing.T) {
	tests := []struct {
		name    string
		pinInfo []WagriPinInfo
		want    bool
	}{
		{
			name:    "has pin info",
			pinInfo: []WagriPinInfo{{FarmerNumber: "12345"}},
			want:    true,
		},
		{
			name:    "no pin info",
			pinInfo: []WagriPinInfo{},
			want:    false,
		},
		{
			name:    "nil pin info",
			pinInfo: nil,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feature := &WagriFeature{
				Properties: WagriProperties{
					PinInfo: tt.pinInfo,
				},
			}
			if got := feature.HasPinInfo(); got != tt.want {
				t.Errorf("HasPinInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWagriPinInfoParseDescriptiveStudyData(t *testing.T) {
	validDate := "2016-08-31"
	invalidDate := "invalid-date"
	emptyDate := ""

	tests := []struct {
		name    string
		data    *string
		wantNil bool
	}{
		{
			name:    "valid date",
			data:    &validDate,
			wantNil: false,
		},
		{
			name:    "invalid date",
			data:    &invalidDate,
			wantNil: true,
		},
		{
			name:    "empty date",
			data:    &emptyDate,
			wantNil: true,
		},
		{
			name:    "nil date",
			data:    nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pinInfo := &WagriPinInfo{
				DescriptiveStudyData: tt.data,
			}
			result := pinInfo.ParseDescriptiveStudyData()
			if tt.wantNil {
				if result != nil {
					t.Errorf("ParseDescriptiveStudyData() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Error("ParseDescriptiveStudyData() = nil, want non-nil")
				}
			}
		})
	}
}
