package errors

import "fmt"

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
	default:
		return "UNKNOWN_ERROR"
	}
}

// GAError はGoogle Analytics ツール固有のエラー
type GAError struct {
	Type    ErrorType
	Message string
	Cause   error
}

// Error はerrorインターフェースの実装
func (e *GAError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// NewAuthError は認証エラーを作成する
func NewAuthError(message string, cause error) *GAError {
	return &GAError{
		Type:    AuthError,
		Message: message,
		Cause:   cause,
	}
}

// NewConfigError は設定エラーを作成する
func NewConfigError(message string, cause error) *GAError {
	return &GAError{
		Type:    ConfigError,
		Message: message,
		Cause:   cause,
	}
}

// NewAPIError はAPIエラーを作成する
func NewAPIError(message string, cause error) *GAError {
	return &GAError{
		Type:    APIError,
		Message: message,
		Cause:   cause,
	}
}

// NewOutputError は出力エラーを作成する
func NewOutputError(message string, cause error) *GAError {
	return &GAError{
		Type:    OutputError,
		Message: message,
		Cause:   cause,
	}
}