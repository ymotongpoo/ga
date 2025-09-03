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
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// LocalServer はOAuth2認証用のローカルHTTPサーバー
type LocalServer struct {
	server   *http.Server
	authCode chan string
	errChan  chan error
	state    string
	port     int
}

// AuthServiceImpl はAuthServiceの実装
type AuthServiceImpl struct {
	config     *OAuth2Config
	oauth2Conf *oauth2.Config
	tokenFile  string
	server     *LocalServer
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

	// ローカルサーバーを初期化（ポート8080を使用）
	server := NewLocalServer(8080)

	return &AuthServiceImpl{
		config:     config,
		oauth2Conf: oauth2Conf,
		tokenFile:  tokenFile,
		server:     server,
	}
}

// Login はOAuth2認証フローを開始する
func (a *AuthServiceImpl) Login(ctx context.Context) error {
	// OAuth2設定の検証
	if a.oauth2Conf.ClientID == "" || a.oauth2Conf.ClientSecret == "" {
		return errors.NewAuthError("OAuth2クライアント設定が不正です", nil)
	}

	// ローカルサーバーを起動
	fmt.Println("認証用ローカルサーバーを起動しています...")
	if err := a.server.Start(ctx); err != nil {
		// ポート使用中エラーの場合は、より具体的なメッセージを表示
		if isPortInUseError(err) {
			return errors.NewAuthError(
				fmt.Sprintf("ポート8080が既に使用されています。他のアプリケーションを終了してから再試行してください: %v", err),
				err)
		}
		return errors.NewAuthError("ローカルサーバーの起動に失敗しました", err)
	}

	// サーバー停止を確実に実行
	defer func() {
		if err := a.server.Stop(ctx); err != nil {
			fmt.Printf("警告: ローカルサーバーの停止に失敗しました: %v\n", err)
		}
	}()

	// OAuth2 stateパラメータを取得
	state := a.server.GetState()

	// 認証URLを生成（オフラインアクセスとstateパラメータを含む）
	authURL := a.oauth2Conf.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce) // 常にリフレッシュトークンを取得

	fmt.Printf("ブラウザで認証ページを開いています...\n")
	fmt.Printf("自動で開かない場合は、以下のURLを手動で開いてください:\n%s\n", authURL)

	// ブラウザを自動で開く
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("ブラウザの自動起動に失敗しました: %v\n", err)
		fmt.Printf("上記のURLを手動でブラウザで開いてください。\n")
	}

	// 認証コードの受信を待機（タイムアウト: 5分）
	fmt.Println("認証の完了を待機しています...")
	authCode, err := a.server.WaitForAuthCode(ctx, 5*time.Minute)
	if err != nil {
		return errors.NewAuthError("認証の完了に失敗しました", err)
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

	// リフレッシュトークンの確認
	if token.RefreshToken == "" {
		fmt.Println("警告: リフレッシュトークンが取得できませんでした。トークンの有効期限が切れた際は再認証が必要です。")
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
	// トークンファイルを安全な権限で作成（所有者のみ読み書き可能）
	file, err := os.OpenFile(a.tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("トークンファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

	// ファイル権限を確実に設定（既存ファイルの場合も）
	if err := file.Chmod(0600); err != nil {
		return fmt.Errorf("トークンファイルの権限設定に失敗しました: %w", err)
	}

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

// NewLocalServer は新しいLocalServerを作成する
func NewLocalServer(port int) *LocalServer {
	state, _ := generateRandomState()
	return &LocalServer{
		authCode: make(chan string, 1),
		errChan:  make(chan error, 1),
		state:    state,
		port:     port,
	}
}

// Start はローカルHTTPサーバーを起動する
func (ls *LocalServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", ls.handleCallback)

	ls.server = &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", ls.port),
		Handler: mux,
	}

	// サーバー起動の成功/失敗を通知するチャネル
	startChan := make(chan error, 1)

	go func() {
		if err := ls.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			startChan <- fmt.Errorf("ローカルサーバーの起動に失敗しました (ポート %d): %w", ls.port, err)
		}
	}()

	// サーバーが起動するまで少し待機してからポートの使用状況を確認
	time.Sleep(100 * time.Millisecond)

	// ポートが使用中かどうかを確認
	select {
	case err := <-startChan:
		return err
	default:
		// サーバーが正常に起動した
		return nil
	}
}

// Stop はローカルHTTPサーバーを停止する
func (ls *LocalServer) Stop(ctx context.Context) error {
	if ls.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return ls.server.Shutdown(ctx)
}

// WaitForAuthCode は認証コードの受信を待機する
func (ls *LocalServer) WaitForAuthCode(ctx context.Context, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case authCode := <-ls.authCode:
		return authCode, nil
	case err := <-ls.errChan:
		return "", err
	case <-ctx.Done():
		return "", fmt.Errorf("認証がタイムアウトしました")
	}
}

// GetState はOAuth2 stateパラメータを取得する
func (ls *LocalServer) GetState() string {
	return ls.state
}

// handleCallback は/callbackエンドポイントのハンドラー
func (ls *LocalServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	// CSRF攻撃防止のためのstate検証
	receivedState := r.URL.Query().Get("state")
	if receivedState != ls.state {
		ls.errChan <- fmt.Errorf("無効なstateパラメータです。CSRF攻撃の可能性があります")
		http.Error(w, "無効なリクエストです", http.StatusBadRequest)
		return
	}

	// エラーチェック
	if errorCode := r.URL.Query().Get("error"); errorCode != "" {
		errorDesc := r.URL.Query().Get("error_description")
		ls.errChan <- fmt.Errorf("認証エラー: %s - %s", errorCode, errorDesc)
		http.Error(w, "認証に失敗しました", http.StatusBadRequest)
		return
	}

	// 認証コードを取得
	authCode := r.URL.Query().Get("code")
	if authCode == "" {
		ls.errChan <- fmt.Errorf("認証コードが取得できませんでした")
		http.Error(w, "認証コードが見つかりません", http.StatusBadRequest)
		return
	}

	// 認証コードをチャネルに送信
	select {
	case ls.authCode <- authCode:
		// 成功レスポンスを返す
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>認証完了</title>
</head>
<body>
    <h1>認証が完了しました</h1>
    <p>このタブを閉じてください。</p>
    <script>
        setTimeout(function() {
            window.close();
        }, 3000);
    </script>
</body>
</html>`)
	default:
		ls.errChan <- fmt.Errorf("認証コードの処理に失敗しました")
		http.Error(w, "内部エラーが発生しました", http.StatusInternalServerError)
	}
}

// generateRandomState はランダムなstateパラメータを生成する
func generateRandomState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// openBrowser はデフォルトブラウザでURLを開く
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// isPortInUseError はポート使用中エラーかどうかを判定する
func isPortInUseError(err error) bool {
	return err != nil && (
	// Linux/Unix系
	fmt.Sprintf("%v", err) == "listen tcp 127.0.0.1:8080: bind: address already in use" ||
		// Windows
		fmt.Sprintf("%v", err) == "listen tcp 127.0.0.1:8080: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted." ||
		// 一般的なパターン
		fmt.Sprintf("%v", err) == "address already in use")
}

// NewGoogleAnalyticsAuthService はGoogle Analytics用の認証サービスを作成する
func NewGoogleAnalyticsAuthService(clientID, clientSecret string) AuthService {
	config := &OAuth2Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{analyticsreporting.AnalyticsReadonlyScope},
	}
	return NewAuthService(config)
}
