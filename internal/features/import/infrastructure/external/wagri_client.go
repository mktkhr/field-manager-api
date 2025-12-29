package external

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mktkhr/field-manager-api/internal/features/import/application/port"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
)

const (
	defaultTimeout = 10 * time.Minute
)

// WagriConfig はwagri API設定
type WagriConfig struct {
	BaseURL string `env:"WAGRI_BASE_URL" envDefault:"https://api.wagri.net"`
	APIKey  string `env:"WAGRI_API_KEY"`
}

// wagriClient はWagriClientの実装
type wagriClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewWagriClient は新しいWagriClientを作成する
func NewWagriClient(cfg *WagriConfig) port.WagriClient {
	return &wagriClient{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// FetchFieldsByCityCode は市区町村コードで圃場データを取得する
func (c *wagriClient) FetchFieldsByCityCode(ctx context.Context, cityCode string) (*entity.WagriResponse, error) {
	data, err := c.FetchFieldsByCityCodeToStream(ctx, cityCode)
	if err != nil {
		return nil, err
	}

	return entity.ParseWagriResponse(data)
}

// FetchFieldsByCityCodeToStream は市区町村コードで圃場データを取得し、バイト列として返す
func (c *wagriClient) FetchFieldsByCityCodeToStream(ctx context.Context, cityCode string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/fields?cityCode=%s", c.baseURL, cityCode)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエスト作成に失敗: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API呼び出しに失敗: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("APIエラー: ステータスコード %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンス読み取りに失敗: %w", err)
	}

	// レスポンスが有効なJSONかチェック
	var test json.RawMessage
	if err := json.Unmarshal(data, &test); err != nil {
		return nil, fmt.Errorf("無効なJSONレスポンス: %w", err)
	}

	return data, nil
}
