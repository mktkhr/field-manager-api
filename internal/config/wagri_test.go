package config

import (
	"testing"
)

// デフォルト値でWagri設定が正しく読み込まれることを確認
func TestLoadWagriConfig_Success_DefaultValues(t *testing.T) {
	// 環境変数をクリア(デフォルト値のテスト)
	t.Setenv("WAGRI_BASE_URL", "")
	t.Setenv("WAGRI_CLIENT_ID", "")
	t.Setenv("WAGRI_CLIENT_SECRET", "")

	cfg, err := LoadWagriConfig()
	if err != nil {
		t.Fatalf("LoadWagriConfig()でエラー発生 = %v", err)
	}

	// デフォルト値を確認
	if cfg.BaseURL != "https://api.wagri.net" {
		t.Errorf("BaseURL = %q, 期待値 %q (デフォルト)", cfg.BaseURL, "https://api.wagri.net")
	}
	if cfg.ClientID != "" {
		t.Errorf("ClientID = %q, 期待値 %q (デフォルト)", cfg.ClientID, "")
	}
	if cfg.ClientSecret != "" {
		t.Errorf("ClientSecret = %q, 期待値 %q (デフォルト)", cfg.ClientSecret, "")
	}
}

// カスタム値でWagri設定が正しく読み込まれることを確認
func TestLoadWagriConfig_Success_CustomValues(t *testing.T) {
	t.Setenv("WAGRI_BASE_URL", "https://custom.wagri.net")
	t.Setenv("WAGRI_CLIENT_ID", "test-client-id")
	t.Setenv("WAGRI_CLIENT_SECRET", "test-client-secret")

	cfg, err := LoadWagriConfig()
	if err != nil {
		t.Fatalf("LoadWagriConfig()でエラー発生 = %v", err)
	}

	if cfg.BaseURL != "https://custom.wagri.net" {
		t.Errorf("BaseURL = %q, 期待値 %q", cfg.BaseURL, "https://custom.wagri.net")
	}
	if cfg.ClientID != "test-client-id" {
		t.Errorf("ClientID = %q, 期待値 %q", cfg.ClientID, "test-client-id")
	}
	if cfg.ClientSecret != "test-client-secret" {
		t.Errorf("ClientSecret = %q, 期待値 %q", cfg.ClientSecret, "test-client-secret")
	}
}

// BaseURLのみカスタム値で、他はデフォルト値の場合のテスト
func TestLoadWagriConfig_Success_PartialCustomValues(t *testing.T) {
	t.Setenv("WAGRI_BASE_URL", "https://staging.wagri.net")
	t.Setenv("WAGRI_CLIENT_ID", "")
	t.Setenv("WAGRI_CLIENT_SECRET", "")

	cfg, err := LoadWagriConfig()
	if err != nil {
		t.Fatalf("LoadWagriConfig()でエラー発生 = %v", err)
	}

	if cfg.BaseURL != "https://staging.wagri.net" {
		t.Errorf("BaseURL = %q, 期待値 %q", cfg.BaseURL, "https://staging.wagri.net")
	}
	if cfg.ClientID != "" {
		t.Errorf("ClientID = %q, 期待値 %q (デフォルト)", cfg.ClientID, "")
	}
	if cfg.ClientSecret != "" {
		t.Errorf("ClientSecret = %q, 期待値 %q (デフォルト)", cfg.ClientSecret, "")
	}
}
