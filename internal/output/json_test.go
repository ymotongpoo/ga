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
	"strings"
	"testing"

	"github.com/ymotongpoo/ga/internal/analytics"
)

func TestJSONRecord_Structure(t *testing.T) {
	// JSONRecord構造体のフィールドが正しく定義されているかテスト
	record := JSONRecord{
		Dimensions: map[string]string{
			"date":     "2023-01-01",
			"pagePath": "/home",
		},
		Metrics: map[string]string{
			"sessions":    "1250",
			"activeUsers": "1100",
		},
		Metadata: JSONMetadata{
			RetrievedAt:  "2023-02-01T10:30:00Z",
			PropertyID:   "987654321",
			StreamID:     "1234567",
			DateRange:    "2023-01-01 to 2023-01-31",
			RecordIndex:  1,
			TotalRecords: 100,
			OutputFormat: "json",
			ToolVersion:  "ga-tool-v1.0",
		},
	}

	// JSON マーシャリングのテスト
	jsonData, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("JSON マーシャリングに失敗しました: %v", err)
	}

	// JSON構造の検証
	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("JSON アンマーシャリングに失敗しました: %v", err)
	}

	// 必須フィールドの存在確認
	if _, exists := unmarshaled["dimensions"]; !exists {
		t.Error("dimensions フィールドが存在しません")
	}
	if _, exists := unmarshaled["metrics"]; !exists {
		t.Error("metrics フィールドが存在しません")
	}
	if _, exists := unmarshaled["metadata"]; !exists {
		t.Error("metadata フィールドが存在しません")
	}

	// メタデータの詳細確認
	metadata, ok := unmarshaled["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("metadata フィールドがオブジェクトではありません")
	}

	expectedMetadataFields := []string{
		"retrieved_at", "property_id", "stream_id", "date_range",
		"record_index", "total_records", "output_format", "tool_version",
	}

	for _, field := range expectedMetadataFields {
		if _, exists := metadata[field]; !exists {
			t.Errorf("metadata に %s フィールドが存在しません", field)
		}
	}
}

func TestCreateKeyValuePairs(t *testing.T) {
	outputService := NewOutputService().(*OutputServiceImpl)

	tests := []struct {
		name            string
		headers         []string
		row             []string
		expectedDims    map[string]string
		expectedMetrics map[string]string
	}{
		{
			name:    "基本的なディメンションとメトリクス",
			headers: []string{"date", "pagePath", "sessions", "activeUsers"},
			row:     []string{"2023-01-01", "/home", "1250", "1100"},
			expectedDims: map[string]string{
				"date":     "2023-01-01",
				"pagePath": "/home",
			},
			expectedMetrics: map[string]string{
				"sessions":    "1250",
				"activeUsers": "1100",
			},
		},
		{
			name:    "プロパティIDを含む",
			headers: []string{"property_id", "date", "sessions"},
			row:     []string{"987654321", "2023-01-01", "1250"},
			expectedDims: map[string]string{
				"property_id": "987654321",
				"date":        "2023-01-01",
			},
			expectedMetrics: map[string]string{
				"sessions": "1250",
			},
		},
		{
			name:    "空の値を含む",
			headers: []string{"date", "pagePath", "sessions"},
			row:     []string{"2023-01-01", "", "1250"},
			expectedDims: map[string]string{
				"date":     "2023-01-01",
				"pagePath": "",
			},
			expectedMetrics: map[string]string{
				"sessions": "1250",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dimensions, metrics := outputService.createKeyValuePairs(tt.headers, tt.row)

			// ディメンションの検証
			if len(dimensions) != len(tt.expectedDims) {
				t.Errorf("ディメンション数が一致しません: 期待値=%d, 実際=%d", len(tt.expectedDims), len(dimensions))
			}

			for key, expectedValue := range tt.expectedDims {
				if actualValue, exists := dimensions[key]; !exists {
					t.Errorf("ディメンション %s が存在しません", key)
				} else if actualValue != expectedValue {
					t.Errorf("ディメンション %s の値が一致しません: 期待値=%s, 実際=%s", key, expectedValue, actualValue)
				}
			}

			// メトリクスの検証
			if len(metrics) != len(tt.expectedMetrics) {
				t.Errorf("メトリクス数が一致しません: 期待値=%d, 実際=%d", len(tt.expectedMetrics), len(metrics))
			}

			for key, expectedValue := range tt.expectedMetrics {
				if actualValue, exists := metrics[key]; !exists {
					t.Errorf("メトリクス %s が存在しません", key)
				} else if actualValue != expectedValue {
					t.Errorf("メトリクス %s の値が一致しません: 期待値=%s, 実際=%s", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestIsDimension(t *testing.T) {
	tests := []struct {
		header   string
		expected bool
	}{
		// 明確なディメンション
		{"date", true},
		{"pagePath", true},
		{"fullURL", true},
		{"country", true},
		{"property_id", true},
		{"stream_id", true},

		// 明確なメトリクス
		{"sessions", false},
		{"activeUsers", false},
		{"newUsers", false},
		{"averageSessionDuration", false},
		{"bounceRate", false},
		{"pageviews", false},

		// 大文字小文字の違い
		{"DATE", true},
		{"SESSIONS", false},
		{"ActiveUsers", false},

		// アンダースコア区切り
		{"page_path", true},
		{"active_users", false},
		{"session_duration", false},

		// 数値的なパターン（メトリクス）
		{"customCount", false},
		{"conversionRate", false},
		{"engagementDuration", false},
		{"totalRevenue", false},

		// 不明なフィールド（デフォルトはディメンション）
		{"unknownField", true},
		{"customDimension", true},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			result := isDimension(tt.header)
			if result != tt.expected {
				t.Errorf("isDimension(%s) = %v, 期待値 %v", tt.header, result, tt.expected)
			}
		})
	}
}

func TestExtractPropertyID(t *testing.T) {
	outputService := NewOutputService().(*OutputServiceImpl)

	tests := []struct {
		name     string
		headers  []string
		row      []string
		expected string
	}{
		{
			name:     "property_id が存在する",
			headers:  []string{"property_id", "date", "sessions"},
			row:      []string{"987654321", "2023-01-01", "1250"},
			expected: "987654321",
		},
		{
			name:     "PROPERTY_ID (大文字) が存在する",
			headers:  []string{"PROPERTY_ID", "date", "sessions"},
			row:      []string{"987654321", "2023-01-01", "1250"},
			expected: "987654321",
		},
		{
			name:     "property_id が存在しない",
			headers:  []string{"date", "sessions"},
			row:      []string{"2023-01-01", "1250"},
			expected: "",
		},
		{
			name:     "空の行",
			headers:  []string{"property_id", "date"},
			row:      []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := outputService.extractPropertyID(tt.row, tt.headers)
			if result != tt.expected {
				t.Errorf("extractPropertyID() = %s, 期待値 %s", result, tt.expected)
			}
		})
	}
}

func TestWriteJSON_BasicFunctionality(t *testing.T) {
	outputService := NewOutputService()

	// テストデータの作成
	testData := &analytics.ReportData{
		Headers: []string{"property_id", "date", "pagePath", "sessions", "activeUsers"},
		Rows: [][]string{
			{"987654321", "2023-01-01", "/home", "1250", "1100"},
			{"987654321", "2023-01-01", "/about", "450", "420"},
		},
		Summary: analytics.ReportSummary{
			TotalRows:  2,
			DateRange:  "2023-01-01 to 2023-01-31",
			Properties: []string{"987654321"},
		},
	}

	// JSON出力のテスト
	var buf bytes.Buffer
	err := outputService.WriteJSON(testData, &buf)
	if err != nil {
		t.Fatalf("WriteJSON でエラーが発生しました: %v", err)
	}

	// 出力されたJSONの検証
	var records []JSONRecord
	if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
		t.Fatalf("出力されたJSONの解析に失敗しました: %v", err)
	}

	// レコード数の確認
	if len(records) != 2 {
		t.Errorf("レコード数が一致しません: 期待値=2, 実際=%d", len(records))
	}

	// 最初のレコードの検証
	firstRecord := records[0]

	// ディメンションの確認
	expectedDimensions := map[string]string{
		"property_id": "987654321",
		"date":        "2023-01-01",
		"pagePath":    "/home",
	}

	for key, expectedValue := range expectedDimensions {
		if actualValue, exists := firstRecord.Dimensions[key]; !exists {
			t.Errorf("ディメンション %s が存在しません", key)
		} else if actualValue != expectedValue {
			t.Errorf("ディメンション %s の値が一致しません: 期待値=%s, 実際=%s", key, expectedValue, actualValue)
		}
	}

	// メトリクスの確認
	expectedMetrics := map[string]string{
		"sessions":    "1250",
		"activeUsers": "1100",
	}

	for key, expectedValue := range expectedMetrics {
		if actualValue, exists := firstRecord.Metrics[key]; !exists {
			t.Errorf("メトリクス %s が存在しません", key)
		} else if actualValue != expectedValue {
			t.Errorf("メトリクス %s の値が一致しません: 期待値=%s, 実際=%s", key, expectedValue, actualValue)
		}
	}

	// メタデータの確認
	metadata := firstRecord.Metadata
	if metadata.PropertyID != "987654321" {
		t.Errorf("メタデータのPropertyIDが一致しません: 期待値=987654321, 実際=%s", metadata.PropertyID)
	}
	if metadata.DateRange != "2023-01-01 to 2023-01-31" {
		t.Errorf("メタデータのDateRangeが一致しません: 期待値=2023-01-01 to 2023-01-31, 実際=%s", metadata.DateRange)
	}
	if metadata.RecordIndex != 1 {
		t.Errorf("メタデータのRecordIndexが一致しません: 期待値=1, 実際=%d", metadata.RecordIndex)
	}
	if metadata.TotalRecords != 2 {
		t.Errorf("メタデータのTotalRecordsが一致しません: 期待値=2, 実際=%d", metadata.TotalRecords)
	}
	if metadata.OutputFormat != "json" {
		t.Errorf("メタデータのOutputFormatが一致しません: 期待値=json, 実際=%s", metadata.OutputFormat)
	}

	// 2番目のレコードのインデックス確認
	secondRecord := records[1]
	if secondRecord.Metadata.RecordIndex != 2 {
		t.Errorf("2番目のレコードのインデックスが一致しません: 期待値=2, 実際=%d", secondRecord.Metadata.RecordIndex)
	}
}

func TestWriteJSON_EmptyData(t *testing.T) {
	outputService := NewOutputService()

	// 空のデータでのテスト
	testData := &analytics.ReportData{
		Headers: []string{"date", "sessions"},
		Rows:    [][]string{},
		Summary: analytics.ReportSummary{
			TotalRows: 0,
			DateRange: "2023-01-01 to 2023-01-31",
		},
	}

	var buf bytes.Buffer
	err := outputService.WriteJSON(testData, &buf)
	if err != nil {
		t.Fatalf("空データでのWriteJSONでエラーが発生しました: %v", err)
	}

	// 空の配列が出力されることを確認
	output := strings.TrimSpace(buf.String())
	if output != "[]" {
		t.Errorf("空データの出力が期待値と一致しません: 期待値=[], 実際=%s", output)
	}
}

func TestWriteJSON_NilData(t *testing.T) {
	outputService := NewOutputService()

	var buf bytes.Buffer
	err := outputService.WriteJSON(nil, &buf)
	if err == nil {
		t.Error("nilデータでエラーが発生しませんでした")
	}

	expectedErrorMessage := "出力データがnilです"
	if !strings.Contains(err.Error(), expectedErrorMessage) {
		t.Errorf("エラーメッセージが期待値と一致しません: 期待値に含まれるべき=%s, 実際=%s", expectedErrorMessage, err.Error())
	}
}