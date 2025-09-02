package auth

import (
	"context"
	"testing"
	"time"
)

func TestAuthServiceImpl_Login_InvalidConfig(t *testing.T) {
	// 無効な設定でAuthServiceを作成
	config := &OAuth2Config{
		ClientID:     "", // 空のクライアントID
		ClientSecret: "",
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes:       []string{"https://www.googleapis.com/auth/analytics.readonly"},
	}

	authService := NewAuthService(config)
	ctx := context.Background()

	err := authService.Login(ctx)
	if err == nil {
		t.Error("無効な設定でもエラーが発生しませんでした")
	}

	if err.Error() != "[AUTH_ERROR] OAuth2クライアント設定が不正です" {
		t.Errorf("期待されるエラーメッセージと異なります: %s", err.Error())
	}
}

func TestAuthServiceImpl_GetCredentials_NoToken(t *testing.T) {
	config := &OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes:       []string{"https://www.googleapis.com/auth/analytics.readonly"},
	}

	authService := NewAuthService(config)
	ctx := context.Background()

	_, err := authService.GetCredentials(ctx)
	if err == nil {
		t.Error("トークンが存在しないのにエラーが発生しませんでした")
	}
}

func TestAuthServiceImpl_IsAuthenticated_NoToken(t *testing.T) {
	config := &OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes:       []string{"https://www.googleapis.com/auth/analytics.readonly"},
	}

	authService := NewAuthService(config)
	ctx := context.Background()

	if authService.IsAuthenticated(ctx) {
		t.Error("トークンが存在しないのに認証済みと判定されました")
	}
}

func TestAuthServiceImpl_ClearToken(t *testing.T) {
	config := &OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes:       []string{"https://www.googleapis.com/auth/analytics.readonly"},
	}

	authService := NewAuthService(config)

	// 存在しないトークンファイルの削除（エラーにならないことを確認）
	err := authService.ClearToken()
	if err != nil {
		t.Errorf("存在しないトークンファイルの削除でエラーが発生しました: %s", err.Error())
	}
}

func TestTokenInfo_Structure(t *testing.T) {
	now := time.Now()
	tokenInfo := &TokenInfo{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       now,
		Valid:        true,
	}

	if tokenInfo.AccessToken != "test-access-token" {
		t.Error("AccessTokenが正しく設定されていません")
	}

	if tokenInfo.RefreshToken != "test-refresh-token" {
		t.Error("RefreshTokenが正しく設定されていません")
	}

	if tokenInfo.TokenType != "Bearer" {
		t.Error("TokenTypeが正しく設定されていません")
	}

	if !tokenInfo.Expiry.Equal(now) {
		t.Error("Expiryが正しく設定されていません")
	}

	if !tokenInfo.Valid {
		t.Error("Validが正しく設定されていません")
	}
}

func TestNewGoogleAnalyticsAuthService(t *testing.T) {
	clientID := "test-client-id"
	clientSecret := "test-client-secret"

	authService := NewGoogleAnalyticsAuthService(clientID, clientSecret)
	if authService == nil {
		t.Error("AuthServiceの作成に失敗しました")
	}

	// 型アサーションでAuthServiceImplにキャスト
	impl, ok := authService.(*AuthServiceImpl)
	if !ok {
		t.Error("AuthServiceImplの型アサーションに失敗しました")
	}

	if impl.config.ClientID != clientID {
		t.Error("ClientIDが正しく設定されていません")
	}

	if impl.config.ClientSecret != clientSecret {
		t.Error("ClientSecretが正しく設定されていません")
	}

	if impl.config.RedirectURL != "urn:ietf:wg:oauth:2.0:oob" {
		t.Error("RedirectURLが正しく設定されていません")
	}

	expectedScopes := []string{"https://www.googleapis.com/auth/analytics.readonly"}
	if len(impl.config.Scopes) != len(expectedScopes) || impl.config.Scopes[0] != expectedScopes[0] {
		t.Error("Scopesが正しく設定されていません")
	}
}