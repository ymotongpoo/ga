package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		want      string
	}{
		{AuthError, "AUTH_ERROR"},
		{ConfigError, "CONFIG_ERROR"},
		{APIError, "API_ERROR"},
		{OutputError, "OUTPUT_ERROR"},
		{NetworkError, "NETWORK_ERROR"},
		{ValidationError, "VALIDATION_ERROR"},
		{ErrorType(999), "UNKNOWN_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.errorType.String(); got != tt.want {
				t.Errorf("ErrorType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGAError_Error(t *testing.T) {
	tests := []struct {
		name     string
		gaError  *GAError
		wantContains []string
	}{
		{
			name: "基本的なエラー",
			gaError: &GAError{
				Type:    AuthError,
				Message: "認証に失敗しました",
				Code:    "AUTH001",
			},
			wantContains: []string{"AUTH_ERROR", "AUTH001", "認証に失敗しました"},
		},
		{
			name: "原因付きエラー",
			gaError: &GAError{
				Type:    ConfigError,
				Message: "設定ファイルが見つかりません",
				Code:    "CONFIG001",
				Cause:   errors.New("file not found"),
			},
			wantContains: []string{"CONFIG_ERROR", "CONFIG001", "設定ファイルが見つかりません", "原因", "file not found"},
		},
		{
			name: "コードなしエラー",
			gaError: &GAError{
				Type:    APIError,
				Message: "APIエラー",
			},
			wantContains: []string{"API_ERROR", "APIエラー"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.gaError.Error()
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("GAError.Error() = %v, should contain %v", got, want)
				}
			}
		})
	}
}

func TestGAError_GetUserFriendlyMessage(t *testing.T) {
	tests := []struct {
		name     string
		gaError  *GAError
		wantContains []string
	}{
		{
			name: "認証エラー",
			gaError: &GAError{
				Type:    AuthError,
				Message: "トークンが無効です",
			},
			wantContains: []string{"認証エラー", "ga --login", "再認証"},
		},
		{
			name: "設定エラー",
			gaError: &GAError{
				Type:    ConfigError,
				Message: "必須フィールドが不足しています",
			},
			wantContains: []string{"設定エラー", "ga.yaml", "設定を確認"},
		},
		{
			name: "APIエラー",
			gaError: &GAError{
				Type:    APIError,
				Message: "リクエストが失敗しました",
			},
			wantContains: []string{"API エラー", "再試行"},
		},
		{
			name: "出力エラー",
			gaError: &GAError{
				Type:    OutputError,
				Message: "ファイル書き込みに失敗しました",
			},
			wantContains: []string{"出力エラー", "ディスクスペース", "権限"},
		},
		{
			name: "ネットワークエラー",
			gaError: &GAError{
				Type:    NetworkError,
				Message: "接続がタイムアウトしました",
			},
			wantContains: []string{"ネットワークエラー", "インターネット接続"},
		},
		{
			name: "バリデーションエラー",
			gaError: &GAError{
				Type:    ValidationError,
				Message: "日付形式が正しくありません",
			},
			wantContains: []string{"バリデーションエラー", "入力データ"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.gaError.GetUserFriendlyMessage()
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("GAError.GetUserFriendlyMessage() = %v, should contain %v", got, want)
				}
			}
		})
	}
}

func TestGAError_WithContext(t *testing.T) {
	gaError := &GAError{
		Type:    AuthError,
		Message: "認証エラー",
	}

	result := gaError.WithContext("user_id", "12345")

	if result.Context == nil {
		t.Error("WithContext() should initialize Context map")
	}

	if result.Context["user_id"] != "12345" {
		t.Errorf("WithContext() Context[user_id] = %v, want %v", result.Context["user_id"], "12345")
	}

	// 追加のコンテキストを設定
	result.WithContext("action", "login")

	if result.Context["action"] != "login" {
		t.Errorf("WithContext() Context[action] = %v, want %v", result.Context["action"], "login")
	}
}

func TestNewAuthError(t *testing.T) {
	cause := errors.New("underlying error")
	gaError := NewAuthError(AuthTokenExpired, cause)

	if gaError.Type != AuthError {
		t.Errorf("NewAuthError() Type = %v, want %v", gaError.Type, AuthError)
	}

	if gaError.Message != AuthTokenExpired {
		t.Errorf("NewAuthError() Message = %v, want %v", gaError.Message, AuthTokenExpired)
	}

	if gaError.Cause != cause {
		t.Errorf("NewAuthError() Cause = %v, want %v", gaError.Cause, cause)
	}

	if gaError.Code != "AUTH001" {
		t.Errorf("NewAuthError() Code = %v, want %v", gaError.Code, "AUTH001")
	}

	if gaError.Context == nil {
		t.Error("NewAuthError() should initialize Context map")
	}
}

func TestNewConfigError(t *testing.T) {
	gaError := NewConfigError(ConfigFileNotFound, nil)

	if gaError.Type != ConfigError {
		t.Errorf("NewConfigError() Type = %v, want %v", gaError.Type, ConfigError)
	}

	if gaError.Code != "CONFIG001" {
		t.Errorf("NewConfigError() Code = %v, want %v", gaError.Code, "CONFIG001")
	}
}

func TestNewAPIError(t *testing.T) {
	gaError := NewAPIError(APIRateLimitExceeded, nil)

	if gaError.Type != APIError {
		t.Errorf("NewAPIError() Type = %v, want %v", gaError.Type, APIError)
	}

	if gaError.Code != "API001" {
		t.Errorf("NewAPIError() Code = %v, want %v", gaError.Code, "API001")
	}
}

func TestNewOutputError(t *testing.T) {
	gaError := NewOutputError(OutputFileCreateFailed, nil)

	if gaError.Type != OutputError {
		t.Errorf("NewOutputError() Type = %v, want %v", gaError.Type, OutputError)
	}

	if gaError.Code != "OUTPUT001" {
		t.Errorf("NewOutputError() Code = %v, want %v", gaError.Code, "OUTPUT001")
	}
}

func TestNewNetworkError(t *testing.T) {
	gaError := NewNetworkError(NetworkConnectionFailed, nil)

	if gaError.Type != NetworkError {
		t.Errorf("NewNetworkError() Type = %v, want %v", gaError.Type, NetworkError)
	}

	if gaError.Code != "NETWORK001" {
		t.Errorf("NewNetworkError() Code = %v, want %v", gaError.Code, "NETWORK001")
	}
}

func TestNewValidationError(t *testing.T) {
	gaError := NewValidationError(ValidationDateInvalid, nil)

	if gaError.Type != ValidationError {
		t.Errorf("NewValidationError() Type = %v, want %v", gaError.Type, ValidationError)
	}

	if gaError.Code != "VALIDATION001" {
		t.Errorf("NewValidationError() Code = %v, want %v", gaError.Code, "VALIDATION001")
	}
}

func TestGenerateErrorCode(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		want      string
	}{
		{AuthError, "AUTH001"},
		{ConfigError, "CONFIG001"},
		{APIError, "API001"},
		{OutputError, "OUTPUT001"},
		{NetworkError, "NETWORK001"},
		{ValidationError, "VALIDATION001"},
		{ErrorType(999), "UNKNOWN001"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := generateErrorCode(tt.errorType); got != tt.want {
				t.Errorf("generateErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}