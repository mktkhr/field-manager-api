package external

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewWagriClient はNewWagriClient関数がWagriClientを正しく生成することをテストする
func TestNewWagriClient(t *testing.T) {
	cfg := &WagriConfig{
		BaseURL: "https://api.wagri.net",
		APIKey:  "test-api-key",
	}

	client := NewWagriClient(cfg)

	if client == nil {
		t.Error("NewWagriClient() returned nil")
	}
}

// TestWagriClient_FetchFieldsByCityCodeToStream はFetchFieldsByCityCodeToStreamメソッドが正常系、サーバーエラー、無効なJSONを正しく処理することをテストする
func TestWagriClient_FetchFieldsByCityCodeToStream(t *testing.T) {
	tests := []struct {
		name           string
		cityCode       string
		serverResponse string
		statusCode     int
		wantErr        bool
		checkAPIKey    bool
	}{
		{
			name:           "success with valid response",
			cityCode:       "163210",
			serverResponse: `{"targetFeatures":[]}`,
			statusCode:     http.StatusOK,
			wantErr:        false,
			checkAPIKey:    true,
		},
		{
			name:           "success with features",
			cityCode:       "163210",
			serverResponse: `{"targetFeatures":[{"type":"Feature","geometry":{"type":"LinearPolygon","coordinates":[[[139.0,35.0]]]},"properties":{"ID":"test-id"}}]}`,
			statusCode:     http.StatusOK,
			wantErr:        false,
			checkAPIKey:    true,
		},
		{
			name:           "server error",
			cityCode:       "163210",
			serverResponse: `{"error":"internal error"}`,
			statusCode:     http.StatusInternalServerError,
			wantErr:        true,
			checkAPIKey:    false,
		},
		{
			name:           "not found",
			cityCode:       "999999",
			serverResponse: `{"error":"not found"}`,
			statusCode:     http.StatusNotFound,
			wantErr:        true,
			checkAPIKey:    false,
		},
		{
			name:           "invalid JSON response",
			cityCode:       "163210",
			serverResponse: `{invalid json`,
			statusCode:     http.StatusOK,
			wantErr:        true,
			checkAPIKey:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedAPIKey string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedAPIKey = r.Header.Get("X-API-Key")

				if r.URL.Query().Get("cityCode") != tt.cityCode {
					t.Errorf("cityCode query param = %q, want %q", r.URL.Query().Get("cityCode"), tt.cityCode)
				}

				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			cfg := &WagriConfig{
				BaseURL: server.URL,
				APIKey:  "test-api-key",
			}
			client := NewWagriClient(cfg)

			data, err := client.(*wagriClient).FetchFieldsByCityCodeToStream(context.Background(), tt.cityCode)

			if tt.wantErr {
				if err == nil {
					t.Error("FetchFieldsByCityCodeToStream() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("FetchFieldsByCityCodeToStream() error = %v", err)
				return
			}

			if len(data) == 0 {
				t.Error("FetchFieldsByCityCodeToStream() returned empty data")
			}

			if tt.checkAPIKey && receivedAPIKey != "test-api-key" {
				t.Errorf("X-API-Key header = %q, want %q", receivedAPIKey, "test-api-key")
			}
		})
	}
}

// TestWagriClient_FetchFieldsByCityCode はFetchFieldsByCityCodeメソッドが正常系とサーバーエラーを正しく処理することをテストする
func TestWagriClient_FetchFieldsByCityCode(t *testing.T) {
	tests := []struct {
		name           string
		cityCode       string
		serverResponse string
		statusCode     int
		wantLen        int
		wantErr        bool
	}{
		{
			name:           "success with empty features",
			cityCode:       "163210",
			serverResponse: `{"targetFeatures":[]}`,
			statusCode:     http.StatusOK,
			wantLen:        0,
			wantErr:        false,
		},
		{
			name:     "success with one feature",
			cityCode: "163210",
			serverResponse: `{
				"targetFeatures": [{
					"type": "Feature",
					"geometry": {
						"type": "LinearPolygon",
						"coordinates": [[[139.0, 35.0], [139.1, 35.0], [139.05, 35.1]]]
					},
					"properties": {
						"ID": "test-id-001",
						"CityCode": "163210",
						"IssueYear": "2024",
						"EditYear": "2024",
						"PointLat": 35.05,
						"PointLng": 139.05,
						"FieldType": "1",
						"Number": 1,
						"SoilLargeCode": "",
						"SoilMiddleCode": "",
						"SoilSmallCode": "",
						"SoilSmallName": "",
						"History": "{}",
						"LastPolygonUuid": "uuid-123",
						"PinInfo": []
					}
				}]
			}`,
			statusCode: http.StatusOK,
			wantLen:    1,
			wantErr:    false,
		},
		{
			name:           "server error",
			cityCode:       "163210",
			serverResponse: `{"error":"internal error"}`,
			statusCode:     http.StatusInternalServerError,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			cfg := &WagriConfig{
				BaseURL: server.URL,
				APIKey:  "test-api-key",
			}
			client := NewWagriClient(cfg)

			response, err := client.(*wagriClient).FetchFieldsByCityCode(context.Background(), tt.cityCode)

			if tt.wantErr {
				if err == nil {
					t.Error("FetchFieldsByCityCode() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("FetchFieldsByCityCode() error = %v", err)
				return
			}

			if len(response.TargetFeatures) != tt.wantLen {
				t.Errorf("FetchFieldsByCityCode() features count = %d, want %d", len(response.TargetFeatures), tt.wantLen)
			}
		})
	}
}

// TestWagriClient_RequestHeaders はWagriClientがリクエストに正しいヘッダーを設定することをテストする
func TestWagriClient_RequestHeaders(t *testing.T) {
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"targetFeatures":[]}`))
	}))
	defer server.Close()

	cfg := &WagriConfig{
		BaseURL: server.URL,
		APIKey:  "my-secret-key",
	}
	client := NewWagriClient(cfg)

	_, err := client.(*wagriClient).FetchFieldsByCityCodeToStream(context.Background(), "163210")
	if err != nil {
		t.Errorf("FetchFieldsByCityCodeToStream() error = %v", err)
		return
	}

	if receivedHeaders.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want %q", receivedHeaders.Get("Content-Type"), "application/json")
	}

	if receivedHeaders.Get("X-API-Key") != "my-secret-key" {
		t.Errorf("X-API-Key = %q, want %q", receivedHeaders.Get("X-API-Key"), "my-secret-key")
	}
}

// TestWagriClient_NoAPIKey はAPIキーが設定されていない場合にヘッダーが空になることをテストする
func TestWagriClient_NoAPIKey(t *testing.T) {
	var receivedAPIKey string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAPIKey = r.Header.Get("X-API-Key")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"targetFeatures":[]}`))
	}))
	defer server.Close()

	cfg := &WagriConfig{
		BaseURL: server.URL,
		APIKey:  "", // 空のAPIキー
	}
	client := NewWagriClient(cfg)

	_, err := client.(*wagriClient).FetchFieldsByCityCodeToStream(context.Background(), "163210")
	if err != nil {
		t.Errorf("FetchFieldsByCityCodeToStream() error = %v", err)
		return
	}

	if receivedAPIKey != "" {
		t.Errorf("X-API-Key should be empty when no API key is configured, got %q", receivedAPIKey)
	}
}

// TestWagriClient_URLConstruction はWagriClientがリクエストURLを正しく構築することをテストする
func TestWagriClient_URLConstruction(t *testing.T) {
	var receivedPath string
	var receivedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		receivedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"targetFeatures":[]}`))
	}))
	defer server.Close()

	cfg := &WagriConfig{
		BaseURL: server.URL,
		APIKey:  "test-key",
	}
	client := NewWagriClient(cfg)

	_, err := client.(*wagriClient).FetchFieldsByCityCodeToStream(context.Background(), "163210")
	if err != nil {
		t.Errorf("FetchFieldsByCityCodeToStream() error = %v", err)
		return
	}

	if receivedPath != "/api/v1/fields" {
		t.Errorf("URL path = %q, want %q", receivedPath, "/api/v1/fields")
	}

	if receivedQuery != "cityCode=163210" {
		t.Errorf("URL query = %q, want %q", receivedQuery, "cityCode=163210")
	}
}

// TestWagriClient_ContextCancellation はコンテキストがキャンセルされた場合にエラーを返すことをテストする
func TestWagriClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// このハンドラーは呼ばれないはず
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"targetFeatures":[]}`))
	}))
	defer server.Close()

	cfg := &WagriConfig{
		BaseURL: server.URL,
		APIKey:  "test-key",
	}
	client := NewWagriClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 即座にキャンセル

	_, err := client.(*wagriClient).FetchFieldsByCityCodeToStream(ctx, "163210")
	if err == nil {
		t.Error("FetchFieldsByCityCodeToStream() expected error for cancelled context, got nil")
	}
}
