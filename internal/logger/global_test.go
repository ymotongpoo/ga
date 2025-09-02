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

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

// テスト前にグローバル状態をリセットする
func resetGlobalLogger() {
	globalLogger = nil
	once = sync.Once{}
}

func TestInitGlobalLogger(t *testing.T) {
	resetGlobalLogger()

	// デバッグモード有効で初期化
	InitGlobalLogger(true)

	if globalLogger == nil {
		t.Fatal("InitGlobalLogger() should initialize globalLogger")
	}

	if !globalLogger.debugMode {
		t.Error("InitGlobalLogger(true) should set debug mode to true")
	}

	// 再度初期化を試行（once.Doにより実行されないはず）
	oldLogger := globalLogger
	InitGlobalLogger(false)

	if globalLogger != oldLogger {
		t.Error("InitGlobalLogger() should only initialize once")
	}

	if globalLogger.debugMode != true {
		t.Error("Second InitGlobalLogger() call should not change debug mode")
	}
}

func TestGetGlobalLogger(t *testing.T) {
	resetGlobalLogger()

	// 初期化前にGetGlobalLoggerを呼び出し
	logger := GetGlobalLogger()

	if logger == nil {
		t.Fatal("GetGlobalLogger() should return a logger instance")
	}

	if logger.debugMode {
		t.Error("GetGlobalLogger() should initialize with debug mode false by default")
	}

	// 同じインスタンスが返されることを確認
	logger2 := GetGlobalLogger()
	if logger != logger2 {
		t.Error("GetGlobalLogger() should return the same instance")
	}
}

func TestGetGlobalLogger_AfterInit(t *testing.T) {
	resetGlobalLogger()

	// 明示的に初期化
	InitGlobalLogger(true)
	logger := GetGlobalLogger()

	if !logger.debugMode {
		t.Error("GetGlobalLogger() should return logger with debug mode true after InitGlobalLogger(true)")
	}
}

func TestGlobalDebug(t *testing.T) {
	resetGlobalLogger()
	InitGlobalLogger(true)

	var buf bytes.Buffer
	GetGlobalLogger().SetOutput(&buf)

	Debug("test debug message: %s", "param")

	output := buf.String()
	if !strings.Contains(output, "DEBUG") {
		t.Error("Global Debug() should output DEBUG level message")
	}
	if !strings.Contains(output, "test debug message: param") {
		t.Error("Global Debug() should format message correctly")
	}
}

func TestGlobalInfo(t *testing.T) {
	resetGlobalLogger()
	InitGlobalLogger(false)

	var buf bytes.Buffer
	GetGlobalLogger().SetOutput(&buf)

	Info("test info message: %d", 123)

	output := buf.String()
	if !strings.Contains(output, "test info message: 123") {
		t.Error("Global Info() should format message correctly")
	}
}

func TestGlobalWarn(t *testing.T) {
	resetGlobalLogger()
	InitGlobalLogger(true)

	var buf bytes.Buffer
	GetGlobalLogger().SetOutput(&buf)

	Warn("test warning message")

	output := buf.String()
	if !strings.Contains(output, "WARN") {
		t.Error("Global Warn() should output WARN level message")
	}
	if !strings.Contains(output, "test warning message") {
		t.Error("Global Warn() should output the message")
	}
}

func TestGlobalError(t *testing.T) {
	resetGlobalLogger()
	InitGlobalLogger(false)

	var buf bytes.Buffer
	GetGlobalLogger().SetErrorOutput(&buf)

	Error("test error message: %v", "error")

	output := buf.String()
	if !strings.Contains(output, "test error message: error") {
		t.Error("Global Error() should format message correctly")
	}
}

func TestGlobalLogAPIRequest(t *testing.T) {
	resetGlobalLogger()
	InitGlobalLogger(true)

	var buf bytes.Buffer
	GetGlobalLogger().SetOutput(&buf)

	headers := map[string]string{
		"Authorization": "Bearer token",
		"Content-Type":  "application/json",
	}

	LogAPIRequest("GET", "https://api.example.com/data", headers)

	output := buf.String()
	if !strings.Contains(output, "API Request") {
		t.Error("Global LogAPIRequest() should log API request")
	}
	if !strings.Contains(output, "GET") {
		t.Error("Global LogAPIRequest() should log HTTP method")
	}
	if !strings.Contains(output, "https://api.example.com/data") {
		t.Error("Global LogAPIRequest() should log URL")
	}
}

func TestGlobalLogAPIResponse(t *testing.T) {
	resetGlobalLogger()
	InitGlobalLogger(true)

	var buf bytes.Buffer
	GetGlobalLogger().SetOutput(&buf)

	LogAPIResponse(200, 1024)

	output := buf.String()
	if !strings.Contains(output, "API Response") {
		t.Error("Global LogAPIResponse() should log API response")
	}
	if !strings.Contains(output, "Status=200") {
		t.Error("Global LogAPIResponse() should log status code")
	}
	if !strings.Contains(output, "Size=1024") {
		t.Error("Global LogAPIResponse() should log response size")
	}
}

func TestGlobalLogConfigLoad(t *testing.T) {
	resetGlobalLogger()
	InitGlobalLogger(false)

	var buf bytes.Buffer
	GetGlobalLogger().SetOutput(&buf)

	LogConfigLoad("/path/to/config.yaml", true)

	output := buf.String()
	if !strings.Contains(output, "設定ファイルを読み込みました") {
		t.Error("Global LogConfigLoad() should log config load success")
	}
	if !strings.Contains(output, "/path/to/config.yaml") {
		t.Error("Global LogConfigLoad() should log config path")
	}
}

func TestGlobalLogConfigLoad_Failure(t *testing.T) {
	resetGlobalLogger()
	InitGlobalLogger(false)

	var buf bytes.Buffer
	GetGlobalLogger().SetErrorOutput(&buf)

	LogConfigLoad("/path/to/config.yaml", false)

	output := buf.String()
	if !strings.Contains(output, "設定ファイルの読み込みに失敗しました") {
		t.Error("Global LogConfigLoad() should log config load failure")
	}
	if !strings.Contains(output, "/path/to/config.yaml") {
		t.Error("Global LogConfigLoad() should log config path")
	}
}

func TestGlobalLogDataProcessing(t *testing.T) {
	resetGlobalLogger()
	InitGlobalLogger(true)

	var buf bytes.Buffer
	GetGlobalLogger().SetOutput(&buf)

	LogDataProcessing(50, 100)

	output := buf.String()
	if !strings.Contains(output, "データ処理中") {
		t.Error("Global LogDataProcessing() should log data processing")
	}
	if !strings.Contains(output, "50/100") {
		t.Error("Global LogDataProcessing() should log progress")
	}
}

func TestGlobalLogDataProcessing_NormalMode(t *testing.T) {
	resetGlobalLogger()
	InitGlobalLogger(false)

	var buf bytes.Buffer
	GetGlobalLogger().SetOutput(&buf)

	// 通常モードでは100件未満は出力されない
	LogDataProcessing(50, 200)
	output1 := buf.String()
	if output1 != "" {
		t.Error("Global LogDataProcessing() should not log in normal mode for non-100 multiples")
	}

	// 100件ちょうどは出力される
	buf.Reset()
	LogDataProcessing(100, 200)
	output2 := buf.String()
	if !strings.Contains(output2, "100/200") {
		t.Error("Global LogDataProcessing() should log in normal mode for 100 multiples")
	}

	// 完了時は出力される
	buf.Reset()
	LogDataProcessing(200, 200)
	output3 := buf.String()
	if !strings.Contains(output3, "200/200") {
		t.Error("Global LogDataProcessing() should log in normal mode when completed")
	}
}

func TestGlobalIsDebugMode(t *testing.T) {
	resetGlobalLogger()
	InitGlobalLogger(true)

	if !IsDebugMode() {
		t.Error("Global IsDebugMode() should return true when initialized with debug mode")
	}

	resetGlobalLogger()
	InitGlobalLogger(false)

	if IsDebugMode() {
		t.Error("Global IsDebugMode() should return false when initialized without debug mode")
	}
}

func TestGlobalIsDebugMode_DefaultInit(t *testing.T) {
	resetGlobalLogger()

	// 明示的に初期化せずにIsDebugModeを呼び出し
	if IsDebugMode() {
		t.Error("Global IsDebugMode() should return false by default")
	}
}

func TestGlobalLogger_ConcurrentAccess(t *testing.T) {
	resetGlobalLogger()

	// 並行してGetGlobalLoggerを呼び出し
	const numGoroutines = 10
	loggers := make([]*Logger, numGoroutines)
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			loggers[index] = GetGlobalLogger()
		}(i)
	}

	wg.Wait()

	// すべて同じインスタンスであることを確認
	firstLogger := loggers[0]
	for i := 1; i < numGoroutines; i++ {
		if loggers[i] != firstLogger {
			t.Errorf("Concurrent GetGlobalLogger() calls should return the same instance")
		}
	}
}

func TestGlobalLogger_ConcurrentInit(t *testing.T) {
	resetGlobalLogger()

	// 並行してInitGlobalLoggerを呼び出し
	const numGoroutines = 10
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			InitGlobalLogger(true)
		}()
	}

	wg.Wait()

	// グローバルロガーが正しく初期化されていることを確認
	logger := GetGlobalLogger()
	if logger == nil {
		t.Fatal("Concurrent InitGlobalLogger() should initialize globalLogger")
	}

	if !logger.debugMode {
		t.Error("Concurrent InitGlobalLogger() should set debug mode correctly")
	}
}

func TestGlobalLogger_MixedConcurrentAccess(t *testing.T) {
	resetGlobalLogger()

	// 初期化とアクセスを並行して実行
	const numGoroutines = 20
	var wg sync.WaitGroup
	results := make([]bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			if index%2 == 0 {
				InitGlobalLogger(true)
			}
			logger := GetGlobalLogger()
			results[index] = logger != nil
		}(i)
	}

	wg.Wait()

	// すべてのgoroutineでロガーが取得できたことを確認
	for i, result := range results {
		if !result {
			t.Errorf("Goroutine %d failed to get global logger", i)
		}
	}

	// 最終的にロガーが初期化されていることを確認
	// 並行実行の結果、デバッグモードの状態は不定なので、初期化されていることのみ確認
	logger := GetGlobalLogger()
	if logger == nil {
		t.Error("Mixed concurrent access should result in logger being initialized")
	}
}