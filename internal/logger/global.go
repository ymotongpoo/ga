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

package logger

import "sync"

var (
	// globalLogger はグローバルなロガーインスタンス
	globalLogger *Logger
	// once はロガーの初期化を一度だけ実行するため
	once sync.Once
)

// InitGlobalLogger はグローバルロガーを初期化する
func InitGlobalLogger(debugMode bool) {
	once.Do(func() {
		globalLogger = NewLogger(debugMode)
	})
}

// GetGlobalLogger はグローバルロガーを取得する
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		// デフォルトでデバッグモードOFFで初期化
		InitGlobalLogger(false)
	}
	return globalLogger
}

// Debug はグローバルロガーでデバッグログを出力する
func Debug(format string, args ...interface{}) {
	GetGlobalLogger().Debug(format, args...)
}

// Info はグローバルロガーで情報ログを出力する
func Info(format string, args ...interface{}) {
	GetGlobalLogger().Info(format, args...)
}

// Warn はグローバルロガーで警告ログを出力する
func Warn(format string, args ...interface{}) {
	GetGlobalLogger().Warn(format, args...)
}

// Error はグローバルロガーでエラーログを出力する
func Error(format string, args ...interface{}) {
	GetGlobalLogger().Error(format, args...)
}

// LogAPIRequest はグローバルロガーでAPIリクエストをログに記録する
func LogAPIRequest(method, url string, headers map[string]string) {
	GetGlobalLogger().LogAPIRequest(method, url, headers)
}

// LogAPIResponse はグローバルロガーでAPIレスポンスをログに記録する
func LogAPIResponse(statusCode int, responseSize int64) {
	GetGlobalLogger().LogAPIResponse(statusCode, responseSize)
}

// LogConfigLoad はグローバルロガーで設定ファイル読み込みをログに記録する
func LogConfigLoad(configPath string, success bool) {
	GetGlobalLogger().LogConfigLoad(configPath, success)
}

// LogDataProcessing はグローバルロガーでデータ処理をログに記録する
func LogDataProcessing(processed, total int) {
	GetGlobalLogger().LogDataProcessing(processed, total)
}

// IsDebugMode はグローバルロガーのデバッグモードを確認する
func IsDebugMode() bool {
	return GetGlobalLogger().IsDebugMode()
}
