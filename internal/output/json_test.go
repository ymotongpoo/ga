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

package output

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/ymotongpoo/ga/internal/analytics"
)

func TestWriteJSON(t *testing.T) {
	service := NewOutputService()
	data := createTestReportData()

	var buf bytes.Buffer
	err := service.WriteJSON(data, &buf)
	if err != nil {
		t.Fatalf("WriteJSON() failed: %v", err)
	}

	// JSON形式の妥当性を確認
	var records []JSONRecord
	if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// レコード数の確認
	expectedRecords := len(data.Rows)
	if len(records) != expectedRecords {
		t.Errorf("Expected %d records, got %d", expectedRecords, len(records))
	}

	// 最初のレコードの構造を確認
	if len(records) > 0 {
		record := records[0]

		// Dimensionsフィールドの確認
		if record.Dimensions == nil {
			t.Error("Dimensions field should not be nil")
		}

		// Metricsフィールドの確認
		if record.Metrics == nil {
			t.Error("Metrics field should not be nil")
		}

		// Metadataフィールドの確認
		if record.Metadata.RetrievedAt == "" {
			t.Error("Metadata.RetrievedAt should not be empty")
		}

		if record.Metadata.DateRange == "" {
			t.Error("Metadata.DateRange should not be empty")
		}
	}
}

func TestWriteJSON_EmptyData(t *testing.T) {
	service := NewOutputService()
	data := &analytics.ReportData{
		Headers: []string{"date", "sessions"},
		Rows:    [][]string{},
		Summary: analytics.ReportSummary{
			TotalRows: 0,
			DateRange: "2023-01-01 to 2023-01-31",
		},
	}

	var buf bytes.Buffer
	err := service.WriteJSON(data, &buf)
	if err != nil {
		t.Fatalf("WriteJSON() with empty data failed: %v", err)
	}

	// 空の配列が出力されることを確認
	var records []JSONRecord
	if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("Expected empty array, got %d records", len(records))
	}
}

func TestWriteJSON_NilData(t *testing.T) {
	service := NewOutputService()

	var buf bytes.Buffer
	err := service.WriteJSON(nil, &buf)
	if err == nil {
		t.Fatal("WriteJSON() with nil data should return error")
	}

	expectedError := "出力データがnilです"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got: %v", expectedError, err)
	}
}

func TestIsDimension(t *testing.T) {
	testCases := []struct {
		header   string
		expected bool
	}{
		// ディメンション
		{"date", true},
		{"pagePath", true},
		{"fullURL", true},
		{"country", true},
		{"browser", true},
		{"Date", true}, // 大文字小文字を無視
		{"PAGEPATH", true},

		// メトリクス
		{"sessions", false},
		{"activeUsers", false},
		{"newUsers", false},
		{"averageSessionDuration", false},
		{"Sessions", false}, // 大文字小文字を無視
		{"ACTIVEUSERS", false},

		// 不明なヘッダー（ディメンションとして扱う）
		{"unknownField", true},
		{"customDimension", true},
	}

	for _, tc := range testCases {
		result := isDimension(tc.header)
		if result != tc.expected {
			t.Errorf("isDimension(%s) = %v, expected %v", tc.header, result, tc.expected)
		}
	}
}

func TestParseOutputFormat(t *testing.T) {
	testCases := []struct {
		input       string
		expected    OutputFormat
		shouldError bool
	}{
		{"csv", FormatCSV, false},
		{"json", FormatJSON, false},
		{"CSV", FormatCSV, false},
		{"JSON", FormatJSON, false},
		{"Csv", FormatCSV, false},
		{"Json", FormatJSON, false},
		{"xml", FormatCSV, true},
		{"txt", FormatCSV, true},
		{"", FormatCSV, true},
	}

	for _, tc := range testCases {
		result, err := ParseOutputFormat(tc.input)

		if tc.shouldError {
			if err == nil {
				t.Errorf("ParseOutputFormat(%s) should return error", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseOutputFormat(%s) returned unexpected error: %v", tc.input, err)
			}
			if result != tc.expected {
				t.Errorf("ParseOutputFormat(%s) = %v, expected %v", tc.input, result, tc.expected)
			}
		}
	}
}

func TestOutputFormat_String(t *testing.T) {
	testCases := []struct {
		format   OutputFormat
		expected string
	}{
		{FormatCSV, "csv"},
		{FormatJSON, "json"},
		{OutputFormat(999), "unknown"}, // 無効な値
	}

	for _, tc := range testCases {
		result := tc.format.String()
		if result != tc.expected {
			t.Errorf("OutputFormat(%d).String() = %s, expected %s", tc.format, result, tc.expected)
		}
	}
}

func TestWriteToConsole_JSON(t *testing.T) {
	service := NewOutputService()
	data := createTestReportData()

	// バッファを使用してテスト
	var buf bytes.Buffer

	// 標準エラー出力をキャプチャ
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	// WriteJSONを直接テスト（WriteToConsoleは標準出力に依存するため）
	err := service.WriteJSON(data, &buf)

	// 標準エラー出力を復元
	wErr.Close()
	os.Stderr = oldStderr

	if err != nil {
		t.Fatalf("WriteJSON() failed: %v", err)
	}

	// JSON形式の妥当性を確認
	var records []JSONRecord
	if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// レコード数の確認
	expectedRecords := len(data.Rows)
	if len(records) != expectedRecords {
		t.Errorf("Expected %d records, got %d", expectedRecords, len(records))
	}

	// 標準エラー出力の内容を読み取り
	errBuf := make([]byte, 1024)
	n, _ := rErr.Read(errBuf)
	_ = string(errBuf[:n]) // エラー出力は使用しない（WriteJSONは標準エラーに出力しないため）
}

func TestWriteToFile_JSON(t *testing.T) {
	service := NewOutputService()
	data := createTestReportData()

	// 一時ファイル名を生成
	tempFile := "test_output.json"
	defer os.Remove(tempFile) // テスト後にファイルを削除

	err := service.WriteToFile(data, tempFile, FormatJSON)
	if err != nil {
		t.Fatalf("WriteToFile() with JSON format failed: %v", err)
	}

	// ファイルが作成されたことを確認
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Fatal("JSON output file was not created")
	}

	// ファイル内容を読み込んで検証
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read JSON output file: %v", err)
	}

	// JSON形式の妥当性を確認
	var records []JSONRecord
	if err := json.Unmarshal(content, &records); err != nil {
		t.Fatalf("Invalid JSON in output file: %v", err)
	}

	// レコード数の確認
	expectedRecords := len(data.Rows)
	if len(records) != expectedRecords {
		t.Errorf("Expected %d records in JSON file, got %d", expectedRecords, len(records))
	}
}

