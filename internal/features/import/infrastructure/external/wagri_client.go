package external

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/mktkhr/field-manager-api/internal/config"
	"github.com/mktkhr/field-manager-api/internal/features/import/application/port"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
)

const (
	defaultTimeout = 10 * time.Minute
	// トークンの有効期限が切れる前に更新するためのバッファ
	tokenExpiryBuffer = 5 * time.Minute
)

// wagriTokenResponse はOAuth2トークンレスポンス
type wagriTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// wagriClient はWagriClientの実装
type wagriClient struct {
	baseURL      string
	clientID     string
	clientSecret string
	httpClient   *http.Client

	// トークンキャッシュ
	token       string
	tokenExpiry time.Time
	mu          sync.RWMutex
}

// NewWagriClient は新しいWagriClientを作成する
func NewWagriClient(cfg *config.WagriConfig) port.WagriClient {
	return &wagriClient{
		baseURL:      cfg.BaseURL,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// getToken はOAuth2トークンを取得する(キャッシュ有効なら再利用)
func (c *wagriClient) getToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		token := c.token
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	return c.fetchToken(ctx)
}

// fetchToken は/Tokenエンドポイントからトークンを取得する
func (c *wagriClient) fetchToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// ダブルチェック: 他のgoroutineが既にトークンを取得した可能性
	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		return c.token, nil
	}

	tokenURL := fmt.Sprintf("%s/Token", c.baseURL)

	// リクエストボディを作成
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("トークンリクエスト作成に失敗: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("トークン取得リクエストに失敗: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("トークン取得エラー: ステータスコード %d, レスポンス読み取り失敗: %w", resp.StatusCode, err)
		}
		return "", fmt.Errorf("トークン取得エラー: ステータスコード %d, レスポンス: %s", resp.StatusCode, string(body))
	}

	var tokenResp wagriTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("トークンレスポンスのパースに失敗: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("トークンレスポンスにアクセストークンが含まれていません")
	}

	// トークンをキャッシュ(有効期限の5分前に期限切れとする)
	c.token = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn)*time.Second - tokenExpiryBuffer)

	return c.token, nil
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
	// OAuth2トークンを取得
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("トークン取得に失敗: %w", err)
	}

	apiURL := fmt.Sprintf("%s/api/v1/fields?cityCode=%s", c.baseURL, cityCode)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエスト作成に失敗: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

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
