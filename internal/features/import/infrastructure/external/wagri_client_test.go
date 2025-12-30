package external

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mktkhr/field-manager-api/internal/config"
)

// OAuth2トークンレスポンスのモック
const mockTokenResponse = `{"access_token":"mock-access-token","token_type":"Bearer","expires_in":3600}`

// TestNewWagriClient はNewWagriClient関数がWagriClientを正しく生成することをテストする
func TestNewWagriClient(t *testing.T) {
	cfg := &config.WagriConfig{
		BaseURL:      "https://api.wagri.net",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	client := NewWagriClient(cfg)

	if client == nil {
		t.Error("NewWagriClient() が nil を返しました")
	}
}

// TestWagriClient_FetchFieldsByCityCodeToStream_Success は正常系でデータを取得できることをテストする
func TestWagriClient_FetchFieldsByCityCodeToStream_Success(t *testing.T) {
	tests := []struct {
		name           string
		cityCode       string
		serverResponse string
	}{
		{
			name:           "空のフィーチャー配列でのレスポンス",
			cityCode:       "163210",
			serverResponse: `{"targetFeatures":[]}`,
		},
		{
			name:           "フィーチャーを含むレスポンス",
			cityCode:       "163210",
			serverResponse: `{"targetFeatures":[{"type":"Feature","geometry":{"type":"LinearPolygon","coordinates":[[[139.0,35.0]]]},"properties":{"ID":"test-id"}}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedAuthHeader string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// トークンエンドポイント
				if r.URL.Path == "/Token" {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(mockTokenResponse))
					return
				}

				// フィールドAPIエンドポイント
				receivedAuthHeader = r.Header.Get("Authorization")

				if r.URL.Query().Get("cityCode") != tt.cityCode {
					t.Errorf("cityCode クエリパラメータ = %q, 期待値 %q", r.URL.Query().Get("cityCode"), tt.cityCode)
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			cfg := &config.WagriConfig{
				BaseURL:      server.URL,
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
			}
			client := NewWagriClient(cfg)

			data, err := client.FetchFieldsByCityCodeToStream(context.Background(), tt.cityCode)

			if err != nil {
				t.Errorf("FetchFieldsByCityCodeToStream() エラー = %v", err)
				return
			}

			if len(data) == 0 {
				t.Error("FetchFieldsByCityCodeToStream() が空のデータを返しました")
			}

			// Bearer認証ヘッダの確認
			if !strings.HasPrefix(receivedAuthHeader, "Bearer ") {
				t.Errorf("Authorization ヘッダ = %q, 期待値 'Bearer' プレフィックス", receivedAuthHeader)
			}
		})
	}
}

// TestWagriClient_FetchFieldsByCityCodeToStream_APIError はAPIエラーを正しく処理することをテストする
func TestWagriClient_FetchFieldsByCityCodeToStream_APIError(t *testing.T) {
	tests := []struct {
		name           string
		cityCode       string
		serverResponse string
		statusCode     int
	}{
		{
			name:           "サーバーエラー",
			cityCode:       "163210",
			serverResponse: `{"error":"internal error"}`,
			statusCode:     http.StatusInternalServerError,
		},
		{
			name:           "Not Found",
			cityCode:       "999999",
			serverResponse: `{"error":"not found"}`,
			statusCode:     http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// トークンエンドポイント
				if r.URL.Path == "/Token" {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(mockTokenResponse))
					return
				}

				// フィールドAPIエンドポイント
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			cfg := &config.WagriConfig{
				BaseURL:      server.URL,
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
			}
			client := NewWagriClient(cfg)

			_, err := client.FetchFieldsByCityCodeToStream(context.Background(), tt.cityCode)

			if err == nil {
				t.Error("FetchFieldsByCityCodeToStream() エラーが期待されましたが、nil が返されました")
			}
		})
	}
}

// TestWagriClient_FetchFieldsByCityCodeToStream_InvalidJSON は無効なJSONレスポンスを正しく処理することをテストする
func TestWagriClient_FetchFieldsByCityCodeToStream_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// トークンエンドポイント
		if r.URL.Path == "/Token" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(mockTokenResponse))
			return
		}

		// 無効なJSONを返す
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	cfg := &config.WagriConfig{
		BaseURL:      server.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewWagriClient(cfg)

	_, err := client.FetchFieldsByCityCodeToStream(context.Background(), "163210")

	if err == nil {
		t.Error("FetchFieldsByCityCodeToStream() 無効なJSONに対してエラーが期待されましたが、nil が返されました")
	}
}

// TestWagriClient_FetchFieldsByCityCode はFetchFieldsByCityCodeメソッドが正常系を正しく処理することをテストする
func TestWagriClient_FetchFieldsByCityCode(t *testing.T) {
	tests := []struct {
		name           string
		cityCode       string
		serverResponse string
		wantLen        int
	}{
		{
			name:           "空のフィーチャー配列",
			cityCode:       "163210",
			serverResponse: `{"targetFeatures":[]}`,
			wantLen:        0,
		},
		{
			name:     "1件のフィーチャー",
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
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// トークンエンドポイント
				if r.URL.Path == "/Token" {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(mockTokenResponse))
					return
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			cfg := &config.WagriConfig{
				BaseURL:      server.URL,
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
			}
			client := NewWagriClient(cfg)

			response, err := client.FetchFieldsByCityCode(context.Background(), tt.cityCode)

			if err != nil {
				t.Errorf("FetchFieldsByCityCode() エラー = %v", err)
				return
			}

			if len(response.TargetFeatures) != tt.wantLen {
				t.Errorf("FetchFieldsByCityCode() フィーチャー数 = %d, 期待値 %d", len(response.TargetFeatures), tt.wantLen)
			}
		})
	}
}

// TestWagriClient_TokenFetch はトークン取得が正しく行われることをテストする
func TestWagriClient_TokenFetch(t *testing.T) {
	var tokenRequestReceived bool
	var receivedClientID, receivedClientSecret, receivedGrantType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// トークンエンドポイント
		if r.URL.Path == "/Token" {
			tokenRequestReceived = true

			if err := r.ParseForm(); err == nil {
				receivedGrantType = r.FormValue("grant_type")
				receivedClientID = r.FormValue("client_id")
				receivedClientSecret = r.FormValue("client_secret")
			}

			// Content-Typeの確認
			if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				t.Errorf("トークンリクエストのContent-Type = %q, 期待値 %q", r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(mockTokenResponse))
			return
		}

		// フィールドAPIエンドポイント
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"targetFeatures":[]}`))
	}))
	defer server.Close()

	cfg := &config.WagriConfig{
		BaseURL:      server.URL,
		ClientID:     "my-client-id",
		ClientSecret: "my-client-secret",
	}
	client := NewWagriClient(cfg)

	_, err := client.FetchFieldsByCityCodeToStream(context.Background(), "163210")
	if err != nil {
		t.Errorf("FetchFieldsByCityCodeToStream() エラー = %v", err)
		return
	}

	if !tokenRequestReceived {
		t.Error("トークンリクエストが送信されませんでした")
	}

	if receivedGrantType != "client_credentials" {
		t.Errorf("grant_type = %q, 期待値 %q", receivedGrantType, "client_credentials")
	}

	if receivedClientID != "my-client-id" {
		t.Errorf("client_id = %q, 期待値 %q", receivedClientID, "my-client-id")
	}

	if receivedClientSecret != "my-client-secret" {
		t.Errorf("client_secret = %q, 期待値 %q", receivedClientSecret, "my-client-secret")
	}
}

// TestWagriClient_TokenFetch_Error はトークン取得エラーを正しく処理することをテストする
func TestWagriClient_TokenFetch_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// トークンエンドポイントでエラーを返す
		if r.URL.Path == "/Token" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_client"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"targetFeatures":[]}`))
	}))
	defer server.Close()

	cfg := &config.WagriConfig{
		BaseURL:      server.URL,
		ClientID:     "invalid-client-id",
		ClientSecret: "invalid-client-secret",
	}
	client := NewWagriClient(cfg)

	_, err := client.FetchFieldsByCityCodeToStream(context.Background(), "163210")

	if err == nil {
		t.Error("トークン取得エラー時にエラーが期待されましたが、nil が返されました")
	}
}

// TestWagriClient_TokenCache はトークンがキャッシュされることをテストする
func TestWagriClient_TokenCache(t *testing.T) {
	tokenRequestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// トークンエンドポイント
		if r.URL.Path == "/Token" {
			tokenRequestCount++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(mockTokenResponse))
			return
		}

		// フィールドAPIエンドポイント
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"targetFeatures":[]}`))
	}))
	defer server.Close()

	cfg := &config.WagriConfig{
		BaseURL:      server.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewWagriClient(cfg)

	// 複数回リクエストを実行
	for i := 0; i < 3; i++ {
		_, err := client.FetchFieldsByCityCodeToStream(context.Background(), "163210")
		if err != nil {
			t.Errorf("FetchFieldsByCityCodeToStream() エラー = %v", err)
			return
		}
	}

	// トークンリクエストは1回だけであるべき
	if tokenRequestCount != 1 {
		t.Errorf("トークンリクエスト回数 = %d, 期待値 1 (キャッシュされるべき)", tokenRequestCount)
	}
}

// TestWagriClient_URLConstruction はWagriClientがリクエストURLを正しく構築することをテストする
func TestWagriClient_URLConstruction(t *testing.T) {
	var receivedPath string
	var receivedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// トークンエンドポイント
		if r.URL.Path == "/Token" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(mockTokenResponse))
			return
		}

		receivedPath = r.URL.Path
		receivedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"targetFeatures":[]}`))
	}))
	defer server.Close()

	cfg := &config.WagriConfig{
		BaseURL:      server.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewWagriClient(cfg)

	_, err := client.FetchFieldsByCityCodeToStream(context.Background(), "163210")
	if err != nil {
		t.Errorf("FetchFieldsByCityCodeToStream() エラー = %v", err)
		return
	}

	if receivedPath != "/api/v1/fields" {
		t.Errorf("URLパス = %q, 期待値 %q", receivedPath, "/api/v1/fields")
	}

	if receivedQuery != "cityCode=163210" {
		t.Errorf("URLクエリ = %q, 期待値 %q", receivedQuery, "cityCode=163210")
	}
}

// TestWagriClient_ContextCancellation はコンテキストがキャンセルされた場合にエラーを返すことをテストする
func TestWagriClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"targetFeatures":[]}`))
	}))
	defer server.Close()

	cfg := &config.WagriConfig{
		BaseURL:      server.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewWagriClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 即座にキャンセル

	_, err := client.FetchFieldsByCityCodeToStream(ctx, "163210")
	if err == nil {
		t.Error("キャンセルされたコンテキストに対してエラーが期待されましたが、nil が返されました")
	}
}
