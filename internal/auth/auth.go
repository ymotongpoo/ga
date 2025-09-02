package auth

import (
	"context"

	"golang.org/x/oauth2"
)

// AuthService は認証機能を提供するインターフェース
type AuthService interface {
	// Login はOAuth2認証フローを開始する
	Login(ctx context.Context) error

	// GetCredentials は有効な認証トークンを取得する
	GetCredentials(ctx context.Context) (*oauth2.Token, error)

	// RefreshToken は期限切れのトークンを更新する
	RefreshToken(ctx context.Context) error
}

// OAuth2Config はOAuth2設定を表す構造体
type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// AuthServiceImpl はAuthServiceの実装
type AuthServiceImpl struct {
	config *OAuth2Config
}

// NewAuthService は新しいAuthServiceを作成する
func NewAuthService(config *OAuth2Config) AuthService {
	return &AuthServiceImpl{
		config: config,
	}
}

// Login はOAuth2認証フローを開始する
func (a *AuthServiceImpl) Login(ctx context.Context) error {
	// TODO: 実装予定
	return nil
}

// GetCredentials は有効な認証トークンを取得する
func (a *AuthServiceImpl) GetCredentials(ctx context.Context) (*oauth2.Token, error) {
	// TODO: 実装予定
	return nil, nil
}

// RefreshToken は期限切れのトークンを更新する
func (a *AuthServiceImpl) RefreshToken(ctx context.Context) error {
	// TODO: 実装予定
	return nil
}