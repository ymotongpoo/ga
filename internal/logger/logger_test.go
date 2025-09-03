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
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name      string
		debugMode bool
		wantLevel LogLevel
	}{
		{
			name:      "デバッグモード有効",
			debugMode: true,
			wantLevel: DEBUG,
		},
		{
			name:      "デバッグモード無効",
			debugMode: false,
			wantLevel: INFO,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.debugMode)
			if logger.level != tt.wantLevel {
				t.Errorf("NewLogger() level = %v, want %v", logger.level, tt.wantLevel)
			}
			if logger.debugMode != tt.debugMode {
				t.Errorf("NewLogger() debugMode = %v, want %v", logger.debugMode, tt.debugMode)
			}
		})
	}
}

func TestLogger_Debug(t *testing.T) {
	tests := []struct {
		name      string
		debugMode bool
		wantEmpty bool
	}{
		{
			name:      "デバッグモード有効時はログ出力",
			debugMode: true,
			wantEmpty: false,
		},
		{
			name:      "デバッグモード無効時はログ出力なし",
			debugMode: false,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(tt.debugMode)
			logger.SetOutput(&buf)

			logger.Debug("test debug message")

			output := buf.String()
			isEmpty := output == ""

			if isEmpty != tt.wantEmpty {
				t.Errorf("Logger.Debug() output empty = %v, want %v", isEmpty, tt.wantEmpty)
			}

			if !tt.wantEmpty && !strings.Contains(output, "DEBUG") {
				t.Errorf("Logger.Debug() output should contain DEBUG level")
			}
		})
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.SetOutput(&buf)

	logger.Info("test info message")

	output := buf.String()
	if output == "" {
		t.Error("Logger.Info() should produce output")
	}
	if !strings.Contains(output, "test info message") {
		t.Error("Logger.Info() output should contain the message")
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.SetErrorOutput(&buf)

	logger.Error("test error message")

	output := buf.String()
	if output == "" {
		t.Error("Logger.Error() should produce output")
	}
	if !strings.Contains(output, "test error message") {
		t.Error("Logger.Error() output should contain the message")
	}
}

func TestLogger_LogAPIRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(true) // デバッグモード有効
	logger.SetOutput(&buf)

	headers := map[string]string{
		"Authorization": "Bearer token",
		"Content-Type":  "application/json",
	}

	logger.LogAPIRequest("GET", "https://api.example.com/data", headers)

	output := buf.String()
	if !strings.Contains(output, "API Request") {
		t.Error("LogAPIRequest() should log API request")
	}
	if !strings.Contains(output, "GET") {
		t.Error("LogAPIRequest() should log HTTP method")
	}
	if !strings.Contains(output, "https://api.example.com/data") {
		t.Error("LogAPIRequest() should log URL")
	}
}

func TestLogger_LogDataProcessing(t *testing.T) {
	tests := []struct {
		name      string
		debugMode bool
		processed int
		total     int
		wantLog   bool
	}{
		{
			name:      "デバッグモードでは常にログ出力",
			debugMode: true,
			processed: 50,
			total:     100,
			wantLog:   true,
		},
		{
			name:      "通常モードでは100件ごとにログ出力",
			debugMode: false,
			processed: 100,
			total:     200,
			wantLog:   true,
		},
		{
			name:      "通常モードでは100件未満はログ出力なし",
			debugMode: false,
			processed: 50,
			total:     200,
			wantLog:   false,
		},
		{
			name:      "完了時は常にログ出力",
			debugMode: false,
			processed: 150,
			total:     150,
			wantLog:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(tt.debugMode)
			logger.SetOutput(&buf)

			logger.LogDataProcessing(tt.processed, tt.total)

			output := buf.String()
			hasLog := output != ""

			if hasLog != tt.wantLog {
				t.Errorf("LogDataProcessing() hasLog = %v, want %v", hasLog, tt.wantLog)
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level LogLevel
		want  string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
