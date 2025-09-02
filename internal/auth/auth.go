// Copyright 2025 Yoshi Yamaguchi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ymotongpoo/ga/internal/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/analyticsreporting/v4"
)

// AuthService は認証機能を提供するインターフェース
type AuthService interface {
	// Login はOAuth2認証フローを開始する
	Login(ctx context.Context) error

	// GetCredentials は有効な認証トークンを取得する
	GetCredentials(ctx context.Context) (*oauth2.Token, error)

	// RefreshToken は期限切れのトークンを更新する
	RefreshToken(ctx context.Context) error

	// IsAuthenticated は認証状態を確認する
	IsAuthenticated(ctx context.Context) bool

	// ClearToken は保存されたトークンを削除する
	ClearToken() error

	// GetTokenInfo はトークンの情報を取得する
	GetTokenInfo() (*TokenInfo, error)
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
	config     *OAuth2Config
	oauth2Conf *oauth2.Config
	tokenFile  string
}

// TokenInfo はトークンの情報を表す構造体
type TokenInfo struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
	Valid        bool      `json:"valid"`
}

// NewAuthService は新しいAuthServiceを作成する
func NewAuthService(config *OAuth2Config) AuthService {
	oauth2Conf := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		Endpoint:     google.Endpoint,
	}

	// ホームディレクトリにトークンファイルを保存
	homeDir, _ := os.UserHomeDir()
	tokenFile := filepath.Join(homeDir, ".ga_token.json")

	return &AuthServiceImpl{
		config:     config,
		oauth2Conf: oauth2Conf,
		tokenFile:  tokenFile,
	}
}

// Login はOAuth2認証フローを開始する
func (a *AuthServiceImpl) Login(ctx context.Context) error {
	// OAuth2設定の検証
	if a.oauth2Conf.ClientID == "" || a.oauth2Conf.ClientSecret == "" {
		return errors.NewAuthError("OAuth2クライアント設定が不正です", nil)
	}

	// 認証URLを生成
	authURL := a.oauth2Conf.AuthCodeURL("state", oauth2.AccessTypeOffline)

	fmt.Printf("以下のURLをブラウザで開いて認証を完了してください:\n%s\n", authURL)
	fmt.Print("認証コードを入力してください: ")

	var authCode string
	if _, err := fmt.Scanln(&authCode); err != nil {
		return errors.NewAuthError("認証コードの読み取りに失敗しました", err)
	}

	if authCode == "" {
		return errors.NewAuthError("認証コードが入力されていません", nil)
	}

	// 認証コードをトークンに交換
	token, err := a.oauth2Conf.Exchange(ctx, authCode)
	if err != nil {
		return errors.NewAuthError("認証コードが無効です。再度認証を行ってください", err)
	}

	// トークンの有効性を検証
	if token.AccessToken == "" {
		return errors.NewAuthError("有効なアクセストークンを取得できませんでした", nil)
	}

	// トークンを保存
	if err := a.saveToken(token); err != nil {
		return errors.NewAuthError("認証トークンの保存に失敗しました", err)
	}

	fmt.Println("認証が完了しました。")
	return nil
}

// GetCredentials は有効な認証トークンを取得する
func (a *AuthServiceImpl) GetCredentials(ctx context.Context) (*oauth2.Token, error) {
	token, err := a.loadToken()
	if err != nil {
		return nil, errors.NewAuthError("認証トークンが見つかりません。'ga --login'で認証を行ってください", err)
	}

	// トークンが期限切れの場合は更新を試行
	if !token.Valid() {
		if err := a.RefreshToken(ctx); err != nil {
			return nil, errors.NewAuthError("認証トークンの更新に失敗しました。再度認証を行ってください", err)
		}
		// 更新後のトークンを再読み込み
		token, err = a.loadToken()
		if err != nil {
			return nil, errors.NewAuthError("更新後のトークンの読み込みに失敗しました", err)
		}
	}

	return token, nil
}

// RefreshToken は期限切れのトークンを更新する
func (a *AuthServiceImpl) RefreshToken(ctx context.Context) error {
	token, err := a.loadToken()
	if err != nil {
		return errors.NewAuthError("既存のトークンの読み込みに失敗しました", err)
	}

	if token.RefreshToken == "" {
		return errors.NewAuthError("リフレッシュトークンが利用できません。再度認証を行ってください", nil)
	}

	// トークンソースを作成してトークンを更新
	tokenSource := a.oauth2Conf.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return errors.NewAuthError("トークンの更新に失敗しました。認証が無効になっている可能性があります", err)
	}

	// 更新されたトークンを保存
	if err := a.saveToken(newToken); err != nil {
		return errors.NewAuthError("更新されたトークンの保存に失敗しました", err)
	}

	return nil
}

// IsAuthenticated は認証状態を確認する
func (a *AuthServiceImpl) IsAuthenticated(ctx context.Context) bool {
	token, err := a.loadToken()
	if err != nil {
		return false
	}

	// トークンが有効かチェック
	if token.Valid() {
		return true
	}

	// 期限切れの場合は更新を試行
	if err := a.RefreshToken(ctx); err != nil {
		return false
	}

	return true
}

// ClearToken は保存されたトークンを削除する
func (a *AuthServiceImpl) ClearToken() error {
	if _, err := os.Stat(a.tokenFile); os.IsNotExist(err) {
		return nil // ファイルが存在しない場合は何もしない
	}

	if err := os.Remove(a.tokenFile); err != nil {
		return errors.NewAuthError("トークンファイルの削除に失敗しました", err)
	}

	return nil
}

// GetTokenInfo はトークンの情報を取得する
func (a *AuthServiceImpl) GetTokenInfo() (*TokenInfo, error) {
	token, err := a.loadToken()
	if err != nil {
		return nil, errors.NewAuthError("トークンの読み込みに失敗しました", err)
	}

	return &TokenInfo{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
		Valid:        token.Valid(),
	}, nil
}

// saveToken はトークンをファイルに保存する
func (a *AuthServiceImpl) saveToken(token *oauth2.Token) error {
	file, err := os.OpenFile(a.tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("トークンファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(token); err != nil {
		return fmt.Errorf("トークンのエンコードに失敗しました: %w", err)
	}

	return nil
}

// loadToken はファイルからトークンを読み込む
func (a *AuthServiceImpl) loadToken() (*oauth2.Token, error) {
	file, err := os.Open(a.tokenFile)
	if err != nil {
		return nil, fmt.Errorf("トークンファイルの読み込みに失敗しました: %w", err)
	}
	defer file.Close()

	var token oauth2.Token
	if err := json.NewDecoder(file).Decode(&token); err != nil {
		return nil, fmt.Errorf("トークンのデコードに失敗しました: %w", err)
	}

	return &token, nil
}

// NewGoogleAnalyticsAuthService はGoogle Analytics用の認証サービスを作成する
func NewGoogleAnalyticsAuthService(clientID, clientSecret string) AuthService {
	config := &OAuth2Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes:       []string{analyticsreporting.AnalyticsReadonlyScope},
	}
	return NewAuthService(config)
}
