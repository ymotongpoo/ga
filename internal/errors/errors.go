package errors

import (
	"fmt"
	"strings"
)

// ErrorType はエラーの種類を表す列挙型
type ErrorType int

const (
	// AuthError は認証関連のエラー
	AuthError ErrorType = iota
	// ConfigError は設定関連のエラー
	ConfigError
	// APIError はAPI関連のエラー
	APIError
	// OutputError は出力関連のエラー
	OutputError
	// NetworkError はネットワーク関連のエラー
	NetworkError
	// ValidationError はバリデーション関連のエラー
	ValidationError
)

// String はErrorTypeの文字列表現を返す
func (e ErrorType) String() string {
	switch e {
	case AuthError:
		return "AUTH_ERROR"
	case ConfigError:
		return "CONFIG_ERROR"
	case APIError:
		return "API_ERROR"
	case OutputError:
		return "OUTPUT_ERROR"
	case NetworkError:
		return "NETWORK_ERROR"
	case ValidationError:
		return "VALIDATION_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}

// GAError はGoogle Analytics ツール固有のエラー
type GAError struct {
	Type    ErrorType
	Message string
	Cause   error
	Code    string // エラーコード（例: "AUTH001", "CONFIG002"）
	Context map[string]interface{} // 追加のコンテキスト情報
}

// Error はerrorインターフェースの実装
func (e *GAError) Error() string {
	var parts []string

	if e.Code != "" {
		parts = append(parts, fmt.Sprintf("[%s:%s]", e.Type, e.Code))
	} else {
		parts = append(parts, fmt.Sprintf("[%s]", e.Type))
	}

	parts = append(parts, e.Message)

	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("原因: %v", e.Cause))
	}

	return strings.Join(parts, " ")
}

// GetUserFriendlyMessage はユーザー向けの分かりやすいメッセージを返す
func (e *GAError) GetUserFriendlyMessage() string {
	switch e.Type {
	case AuthError:
		return fmt.Sprintf("認証エラー: %s\n解決方法: 'ga --login' コマンドを実行して再認証してください。", e.Message)
	case ConfigError:
		return fmt.Sprintf("設定エラー: %s\n解決方法: ga.yaml ファイルの設定を確認してください。", e.Message)
	case APIError:
		return fmt.Sprintf("API エラー: %s\n解決方法: しばらく待ってから再試行してください。", e.Message)
	case OutputError:
		return fmt.Sprintf("出力エラー: %s\n解決方法: 出力先のディスクスペースや権限を確認してください。", e.Message)
	case NetworkError:
		return fmt.Sprintf("ネットワークエラー: %s\n解決方法: インターネット接続を確認してください。", e.Message)
	case ValidationError:
		return fmt.Sprintf("バリデーションエラー: %s\n解決方法: 入力データを確認してください。", e.Message)
	default:
		return fmt.Sprintf("エラー: %s", e.Message)
	}
}

// WithContext はコンテキスト情報を追加する
func (e *GAError) WithContext(key string, value interface{}) *GAError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// 標準化されたエラーメッセージ
var (
	// 認証関連のエラーメッセージ
	AuthTokenExpired    = "認証トークンの有効期限が切れています"
	AuthTokenInvalid    = "認証トークンが無効です"
	AuthCredentialsMissing = "認証情報が見つかりません"

	// 設定関連のエラーメッセージ
	ConfigFileNotFound  = "設定ファイルが見つかりません"
	ConfigInvalidFormat = "設定ファイルの形式が正しくありません"
	ConfigMissingField  = "必須フィールドが設定されていません"

	// API関連のエラーメッセージ
	APIRateLimitExceeded = "API制限に達しました"
	APIRequestFailed     = "APIリクエストが失敗しました"
	APIInvalidResponse   = "APIレスポンスが無効です"

	// 出力関連のエラーメッセージ
	OutputFileCreateFailed = "出力ファイルの作成に失敗しました"
	OutputWriteFailed      = "データの書き込みに失敗しました"

	// ネットワーク関連のエラーメッセージ
	NetworkConnectionFailed = "ネットワーク接続に失敗しました"
	NetworkTimeout         = "ネットワークタイムアウトが発生しました"

	// バリデーション関連のエラーメッセージ
	ValidationDateInvalid = "日付形式が正しくありません"
	ValidationFieldEmpty  = "必須フィールドが空です"
)

// NewAuthError は認証エラーを作成する
func NewAuthError(message string, cause error) *GAError {
	return &GAError{
		Type:    AuthError,
		Message: message,
		Cause:   cause,
		Code:    generateErrorCode(AuthError),
		Context: make(map[string]interface{}),
	}
}

// NewConfigError は設定エラーを作成する
func NewConfigError(message string, cause error) *GAError {
	return &GAError{
		Type:    ConfigError,
		Message: message,
		Cause:   cause,
		Code:    generateErrorCode(ConfigError),
		Context: make(map[string]interface{}),
	}
}

// NewAPIError はAPIエラーを作成する
func NewAPIError(message string, cause error) *GAError {
	return &GAError{
		Type:    APIError,
		Message: message,
		Cause:   cause,
		Code:    generateErrorCode(APIError),
		Context: make(map[string]interface{}),
	}
}

// NewOutputError は出力エラーを作成する
func NewOutputError(message string, cause error) *GAError {
	return &GAError{
		Type:    OutputError,
		Message: message,
		Cause:   cause,
		Code:    generateErrorCode(OutputError),
		Context: make(map[string]interface{}),
	}
}

// NewNetworkError はネットワークエラーを作成する
func NewNetworkError(message string, cause error) *GAError {
	return &GAError{
		Type:    NetworkError,
		Message: message,
		Cause:   cause,
		Code:    generateErrorCode(NetworkError),
		Context: make(map[string]interface{}),
	}
}

// NewValidationError はバリデーションエラーを作成する
func NewValidationError(message string, cause error) *GAError {
	return &GAError{
		Type:    ValidationError,
		Message: message,
		Cause:   cause,
		Code:    generateErrorCode(ValidationError),
		Context: make(map[string]interface{}),
	}
}

// generateErrorCode はエラータイプに基づいてエラーコードを生成する
func generateErrorCode(errorType ErrorType) string {
	switch errorType {
	case AuthError:
		return "AUTH001"
	case ConfigError:
		return "CONFIG001"
	case APIError:
		return "API001"
	case OutputError:
		return "OUTPUT001"
	case NetworkError:
		return "NETWORK001"
	case ValidationError:
		return "VALIDATION001"
	default:
		return "UNKNOWN001"
	}
}