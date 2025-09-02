package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// LogLevel はログレベルを表す列挙型
type LogLevel int

const (
	// DEBUG は詳細なデバッグ情報
	DEBUG LogLevel = iota
	// INFO は一般的な情報
	INFO
	// WARN は警告メッセージ
	WARN
	// ERROR はエラーメッセージ
	ERROR
)

// String はLogLevelの文字列表現を返す
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger はログ機能を提供する構造体
type Logger struct {
	level      LogLevel
	debugMode  bool
	output     io.Writer
	errorOutput io.Writer
	logger     *log.Logger
	errorLogger *log.Logger
}

// NewLogger は新しいLoggerインスタンスを作成する
func NewLogger(debugMode bool) *Logger {
	level := INFO
	if debugMode {
		level = DEBUG
	}

	logger := &Logger{
		level:       level,
		debugMode:   debugMode,
		output:      os.Stdout,
		errorOutput: os.Stderr,
	}

	// 標準ログとエラーログの設定
	logger.logger = log.New(logger.output, "", 0)
	logger.errorLogger = log.New(logger.errorOutput, "", 0)

	return logger
}

// SetLevel はログレベルを設定する
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetOutput は出力先を設定する
func (l *Logger) SetOutput(output io.Writer) {
	l.output = output
	l.logger = log.New(l.output, "", 0)
}

// SetErrorOutput はエラー出力先を設定する
func (l *Logger) SetErrorOutput(output io.Writer) {
	l.errorOutput = output
	l.errorLogger = log.New(l.errorOutput, "", 0)
}

// Debug はデバッグレベルのログを出力する
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG {
		l.log(DEBUG, format, args...)
	}
}

// Info は情報レベルのログを出力する
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		l.log(INFO, format, args...)
	}
}

// Warn は警告レベルのログを出力する
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WARN {
		l.log(WARN, format, args...)
	}
}

// Error はエラーレベルのログを出力する
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ERROR {
		l.logError(ERROR, format, args...)
	}
}

// log は内部的にログを出力する
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)

	if l.debugMode {
		// デバッグモードでは詳細な情報を含める
		logMessage := fmt.Sprintf("[%s] %s %s", timestamp, level, message)
		l.logger.Println(logMessage)
	} else {
		// 通常モードではシンプルな出力
		if level >= INFO {
			l.logger.Println(message)
		}
	}
}

// logError はエラー出力にログを出力する
func (l *Logger) logError(level LogLevel, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)

	if l.debugMode {
		// デバッグモードでは詳細な情報を含める
		logMessage := fmt.Sprintf("[%s] %s %s", timestamp, level, message)
		l.errorLogger.Println(logMessage)
	} else {
		// 通常モードではシンプルな出力
		l.errorLogger.Println(message)
	}
}

// LogAPIRequest はAPIリクエストの詳細をログに記録する
func (l *Logger) LogAPIRequest(method, url string, headers map[string]string) {
	if l.debugMode {
		l.Debug("API Request: %s %s", method, url)
		for key, value := range headers {
			l.Debug("  Header: %s = %s", key, value)
		}
	}
}

// LogAPIResponse はAPIレスポンスの詳細をログに記録する
func (l *Logger) LogAPIResponse(statusCode int, responseSize int64) {
	if l.debugMode {
		l.Debug("API Response: Status=%d, Size=%d bytes", statusCode, responseSize)
	}
}

// LogConfigLoad は設定ファイル読み込みをログに記録する
func (l *Logger) LogConfigLoad(configPath string, success bool) {
	if success {
		l.Info("設定ファイルを読み込みました: %s", configPath)
		if l.debugMode {
			l.Debug("Config file loaded successfully from: %s", configPath)
		}
	} else {
		l.Error("設定ファイルの読み込みに失敗しました: %s", configPath)
	}
}

// LogDataProcessing はデータ処理の進行状況をログに記録する
func (l *Logger) LogDataProcessing(processed, total int) {
	if l.debugMode {
		l.Debug("データ処理中: %d/%d (%d%%)", processed, total, (processed*100)/total)
	} else if processed%100 == 0 || processed == total {
		// 通常モードでは100件ごと、または完了時のみ表示
		l.Info("データ処理中: %d/%d", processed, total)
	}
}

// IsDebugMode はデバッグモードが有効かどうかを返す
func (l *Logger) IsDebugMode() bool {
	return l.debugMode
}