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

func TestJSONWriter_WriteRecords(t *testing.T) {
	jsonWriter := &JSONWriter{
		encoding:      "UTF-8",
		indent:        "  ",
		escapeHTML:    false,
		compactOutput: false,
	}

	// テストレコードの作成
	records := []JSONRecord{
		{
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
				DateRange:    "2023-01-01 to 2023-01-31",
				RecordIndex:  1,
				TotalRecords: 1,
				OutputFormat: "json",
			},
		},
	}

	// 通常の出力テスト
	var buf bytes.Buffer
	err := jsonWriter.writeRecords(records, &buf)
	if err != nil {
		t.Fatalf("writeRecords でエラーが発生しました: %v", err)
	}

	// 出力されたJSONの検証
	var outputRecords []JSONRecord
	if err := json.Unmarshal(buf.Bytes(), &outputRecords); err != nil {
		t.Fatalf("出力されたJSONの解析に失敗しました: %v", err)
	}

	if len(outputRecords) != 1 {
		t.Errorf("レコード数が一致しません: 期待値=1, 実際=%d", len(outputRecords))
	}

	// レコード内容の検証
	record := outputRecords[0]
	if record.Dimensions["date"] != "2023-01-01" {
		t.Errorf("date ディメンションが一致しません: 期待値=2023-01-01, 実際=%s", record.Dimensions["date"])
	}
	if record.Metrics["sessions"] != "1250" {
		t.Errorf("sessions メトリクスが一致しません: 期待値=1250, 実際=%s", record.Metrics["sessions"])
	}
}

func TestJSONWriter_WriteRecordsWithOptions(t *testing.T) {
	jsonWriter := &JSONWriter{
		encoding:      "UTF-8",
		indent:        "  ",
		escapeHTML:    false,
		compactOutput: false,
	}

	records := []JSONRecord{
		{
			Dimensions: map[string]string{"date": "2023-01-01"},
			Metrics:    map[string]string{"sessions": "1250"},
			Metadata: JSONMetadata{
				RetrievedAt:  "2023-02-01T10:30:00Z",
				RecordIndex:  1,
				TotalRecords: 1,
				OutputFormat: "json",
			},
		},
	}

	tests := []struct {
		name     string
		options  JSONWriteOptions
		validate func(t *testing.T, output string)
	}{
		{
			name: "コンパクト出力",
			options: JSONWriteOptions{
				CompactOutput: boolPtr(true),
			},
			validate: func(t *testing.T, output string) {
				// コンパクト出力では改行やインデントが最小限
				if strings.Contains(output, "  ") {
					t.Error("コンパクト出力にインデントが含まれています")
				}
			},
		},
		{
			name: "カスタムインデント",
			options: JSONWriteOptions{
				Indent: stringPtr("    "), // 4スペース
			},
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "    ") {
					t.Error("カスタムインデント（4スペース）が適用されていません")
				}
			},
		},
		{
			name: "HTMLエスケープ有効",
			options: JSONWriteOptions{
				EscapeHTML: boolPtr(true),
			},
			validate: func(t *testing.T, output string) {
				// HTMLエスケープが有効な場合の検証
				// この場合は特別な文字がないので、エラーが発生しないことを確認
				var records []JSONRecord
				if err := json.Unmarshal([]byte(output), &records); err != nil {
					t.Errorf("HTMLエスケープ有効時のJSON解析に失敗: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := jsonWriter.writeRecordsWithOptions(records, &buf, tt.options)
			if err != nil {
				t.Fatalf("writeRecordsWithOptions でエラーが発生しました: %v", err)
			}

			output := buf.String()
			tt.validate(t, output)

			// 基本的なJSON妥当性の確認
			var outputRecords []JSONRecord
			if err := json.Unmarshal([]byte(output), &outputRecords); err != nil {
				t.Errorf("出力されたJSONが無効です: %v", err)
			}
		})
	}
}

func TestJSONWriter_ValidateJSONOutput(t *testing.T) {
	jsonWriter := &JSONWriter{}

	tests := []struct {
		name      string
		jsonData  string
		expectErr bool
		errMsg    string
	}{
		{
			name: "有効なJSON",
			jsonData: `[{
				"dimensions": {"date": "2023-01-01"},
				"metrics": {"sessions": "1250"},
				"metadata": {
					"retrieved_at": "2023-02-01T10:30:00Z",
					"record_index": 1,
					"total_records": 1,
					"output_format": "json"
				}
			}]`,
			expectErr: false,
		},
		{
			name:      "無効なJSON",
			jsonData:  `[{"invalid": json}]`,
			expectErr: true,
			errMsg:    "出力されたJSONが無効です",
		},
		{
			name: "dimensions が nil",
			jsonData: `[{
				"dimensions": null,
				"metrics": {"sessions": "1250"},
				"metadata": {"retrieved_at": "2023-02-01T10:30:00Z"}
			}]`,
			expectErr: true,
			errMsg:    "dimensions が nil です",
		},
		{
			name: "metrics が nil",
			jsonData: `[{
				"dimensions": {"date": "2023-01-01"},
				"metrics": null,
				"metadata": {"retrieved_at": "2023-02-01T10:30:00Z"}
			}]`,
			expectErr: true,
			errMsg:    "metrics が nil です",
		},
		{
			name: "retrieved_at が空",
			jsonData: `[{
				"dimensions": {"date": "2023-01-01"},
				"metrics": {"sessions": "1250"},
				"metadata": {"retrieved_at": ""}
			}]`,
			expectErr: true,
			errMsg:    "retrieved_at が空です",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := jsonWriter.validateJSONOutput([]byte(tt.jsonData))

			if tt.expectErr {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("期待されたエラーメッセージが含まれていません: 期待=%s, 実際=%s", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラーが発生しました: %v", err)
				}
			}
		})
	}
}

func TestJSONWriter_FormatJSONForDisplay(t *testing.T) {
	jsonWriter := &JSONWriter{
		encoding:      "UTF-8",
		indent:        "\t", // タブインデント
		escapeHTML:    true,
		compactOutput: true, // 元の設定はコンパクト
	}

	records := []JSONRecord{
		{
			Dimensions: map[string]string{
				"date":     "2023-01-01",
				"pagePath": "/home",
			},
			Metrics: map[string]string{
				"sessions": "1250",
			},
			Metadata: JSONMetadata{
				RetrievedAt:  "2023-02-01T10:30:00Z",
				RecordIndex:  1,
				TotalRecords: 1,
				OutputFormat: "json",
			},
		},
	}

	// 表示用フォーマット
	formatted, err := jsonWriter.formatJSONForDisplay(records)
	if err != nil {
		t.Fatalf("formatJSONForDisplay でエラーが発生しました: %v", err)
	}

	// 表示用設定が適用されていることを確認（2スペースインデント）
	if !strings.Contains(formatted, "  ") {
		t.Error("表示用の2スペースインデントが適用されていません")
	}

	// 元の設定が保持されていることを確認
	if jsonWriter.indent != "\t" {
		t.Error("元のインデント設定が変更されています")
	}
	if !jsonWriter.compactOutput {
		t.Error("元のコンパクト出力設定が変更されています")
	}

	// 出力されたJSONが有効であることを確認
	var outputRecords []JSONRecord
	if err := json.Unmarshal([]byte(formatted), &outputRecords); err != nil {
		t.Errorf("表示用JSONが無効です: %v", err)
	}
}

func TestJSONWriter_UTF8Encoding(t *testing.T) {
	jsonWriter := &JSONWriter{
		encoding:      "UTF-8",
		indent:        "  ",
		escapeHTML:    false,
		compactOutput: false,
	}

	// 日本語を含むテストデータ
	records := []JSONRecord{
		{
			Dimensions: map[string]string{
				"pagePath":  "/ホーム",
				"pageTitle": "ホームページ",
			},
			Metrics: map[string]string{
				"sessions": "1250",
			},
			Metadata: JSONMetadata{
				RetrievedAt:  "2023-02-01T10:30:00Z",
				RecordIndex:  1,
				TotalRecords: 1,
				OutputFormat: "json",
			},
		},
	}

	var buf bytes.Buffer
	err := jsonWriter.writeRecords(records, &buf)
	if err != nil {
		t.Fatalf("UTF-8エンコーディングテストでエラーが発生しました: %v", err)
	}

	// UTF-8として正しく出力されていることを確認
	output := buf.String()
	if !strings.Contains(output, "ホーム") {
		t.Error("日本語文字が正しく出力されていません")
	}

	// JSONとして正しく解析できることを確認
	var outputRecords []JSONRecord
	if err := json.Unmarshal(buf.Bytes(), &outputRecords); err != nil {
		t.Fatalf("UTF-8 JSONの解析に失敗しました: %v", err)
	}

	// 日本語文字が正しく保持されていることを確認
	if outputRecords[0].Dimensions["pagePath"] != "/ホーム" {
		t.Errorf("日本語文字が正しく保持されていません: 期待値=/ホーム, 実際=%s", outputRecords[0].Dimensions["pagePath"])
	}
}

func TestParseOutputFormat(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedFormat OutputFormat
		expectError    bool
		errorContains  string
	}{
		{
			name:           "有効なCSV形式（小文字）",
			input:          "csv",
			expectedFormat: FormatCSV,
			expectError:    false,
		},
		{
			name:           "有効なJSON形式（小文字）",
			input:          "json",
			expectedFormat: FormatJSON,
			expectError:    false,
		},
		{
			name:           "有効なCSV形式（大文字）",
			input:          "CSV",
			expectedFormat: FormatCSV,
			expectError:    false,
		},
		{
			name:           "有効なJSON形式（大文字）",
			input:          "JSON",
			expectedFormat: FormatJSON,
			expectError:    false,
		},
		{
			name:           "有効なCSV形式（混合ケース）",
			input:          "Csv",
			expectedFormat: FormatCSV,
			expectError:    false,
		},
		{
			name:           "空文字列（デフォルト）",
			input:          "",
			expectedFormat: FormatCSV,
			expectError:    false,
		},
		{
			name:           "スペースのみ（デフォルト）",
			input:          "   ",
			expectedFormat: FormatCSV,
			expectError:    false,
		},
		{
			name:           "前後にスペースがあるCSV",
			input:          "  csv  ",
			expectedFormat: FormatCSV,
			expectError:    false,
		},
		{
			name:        "無効な形式",
			input:       "xml",
			expectError: true,
			errorContains: "無効な出力形式",
		},
		{
			name:        "無効な形式（数字）",
			input:       "123",
			expectError: true,
			errorContains: "無効な出力形式",
		},
		{
			name:        "無効な形式（特殊文字）",
			input:       "csv!",
			expectError: true,
			errorContains: "無効な出力形式",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, err := ParseOutputFormat(tt.input)

			if tt.expectError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("期待されたエラーメッセージが含まれていません: 期待=%s, 実際=%s", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラーが発生しました: %v", err)
				} else if format != tt.expectedFormat {
					t.Errorf("出力形式が一致しません: 期待=%v, 実際=%v", tt.expectedFormat, format)
				}
			}
		})
	}
}

func TestGetDefaultOutputFormat(t *testing.T) {
	defaultFormat := GetDefaultOutputFormat()
	if defaultFormat != FormatCSV {
		t.Errorf("デフォルト出力形式が期待値と一致しません: 期待=%v, 実際=%v", FormatCSV, defaultFormat)
	}
}

func TestIsValidOutputFormat(t *testing.T) {
	tests := []struct {
		format   string
		expected bool
	}{
		{"csv", true},
		{"json", true},
		{"CSV", true},
		{"JSON", true},
		{"", true}, // 空文字列はデフォルトとして有効
		{"xml", false},
		{"txt", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := IsValidOutputFormat(tt.format)
			if result != tt.expected {
				t.Errorf("IsValidOutputFormat(%s) = %v, 期待値 %v", tt.format, result, tt.expected)
			}
		})
	}
}

func TestGetSupportedFormats(t *testing.T) {
	formats := GetSupportedFormats()

	expectedFormats := []string{"csv", "json"}
	if len(formats) != len(expectedFormats) {
		t.Errorf("サポートされている形式の数が一致しません: 期待=%d, 実際=%d", len(expectedFormats), len(formats))
	}

	for _, expected := range expectedFormats {
		found := false
		for _, format := range formats {
			if format == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("期待される形式 '%s' がサポートされている形式に含まれていません", expected)
		}
	}
}

func TestOutputFormat_String(t *testing.T) {
	tests := []struct {
		format   OutputFormat
		expected string
	}{
		{FormatCSV, "csv"},
		{FormatJSON, "json"},
		{OutputFormat(999), "unknown"}, // 無効な値
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.format.String()
			if result != tt.expected {
				t.Errorf("OutputFormat.String() = %s, 期待値 %s", result, tt.expected)
			}
		})
	}
}

func TestWriteOutput_FormatSelection(t *testing.T) {
	outputService := NewOutputService()

	// テストデータの作成
	testData := &analytics.ReportData{
		Headers: []string{"date", "pagePath", "sessions"},
		Rows: [][]string{
			{"2023-01-01", "/home", "1250"},
		},
		Summary: analytics.ReportSummary{
			TotalRows: 1,
			DateRange: "2023-01-01 to 2023-01-31",
		},
	}

	tests := []struct {
		name         string
		format       OutputFormat
		expectError  bool
		validateFunc func(t *testing.T, output string)
	}{
		{
			name:        "CSV形式での出力",
			format:      FormatCSV,
			expectError: false,
			validateFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "date,pagePath,sessions") {
					t.Error("CSVヘッダーが含まれていません")
				}
				if !strings.Contains(output, "2023-01-01,/home,1250") {
					t.Error("CSVデータが含まれていません")
				}
			},
		},
		{
			name:        "JSON形式での出力",
			format:      FormatJSON,
			expectError: false,
			validateFunc: func(t *testing.T, output string) {
				var records []JSONRecord
				if err := json.Unmarshal([]byte(output), &records); err != nil {
					t.Errorf("JSON出力の解析に失敗しました: %v", err)
					return
				}
				if len(records) != 1 {
					t.Errorf("JSONレコード数が一致しません: 期待=1, 実際=%d", len(records))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			// 形式に応じて適切なメソッドを呼び出し
			var err error
			switch tt.format {
			case FormatCSV:
				err = outputService.WriteCSV(testData, &buf)
			case FormatJSON:
				err = outputService.WriteJSON(testData, &buf)
			}

			if tt.expectError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラーが発生しました: %v", err)
				} else {
					output := buf.String()
					if tt.validateFunc != nil {
						tt.validateFunc(t, output)
					}
				}
			}
		})
	}
}