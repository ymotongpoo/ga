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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ymotongpoo/ga/internal/analytics"
)

// TestJSONWriter_ComprehensiveUnitTests はJSONライターの包括的な単体テスト
// 要件4.6, 4.9: 構造化されたJSON配列の生成、UTF-8エンコーディング対応
func TestJSONWriter_ComprehensiveUnitTests(t *testing.T) {
	t.Run("JSONWriter初期化テスト", func(t *testing.T) {
		jsonWriter := &JSONWriter{
			encoding:      "UTF-8",
			indent:        "  ",
			escapeHTML:    false,
			sortKeys:      false,
			compactOutput: false,
		}

		if jsonWriter.encoding != "UTF-8" {
			t.Errorf("エンコーディングが期待値と一致しません: 期待=UTF-8, 実際=%s", jsonWriter.encoding)
		}
		if jsonWriter.indent != "  " {
			t.Errorf("インデントが期待値と一致しません: 期待='  ', 実際='%s'", jsonWriter.indent)
		}
		if jsonWriter.escapeHTML {
			t.Error("HTMLエスケープが無効になっていません")
		}
		if jsonWriter.compactOutput {
			t.Error("コンパクト出力が無効になっていません")
		}
	})

	t.Run("writeRecords基本機能テスト", func(t *testing.T) {
		jsonWriter := &JSONWriter{
			encoding:      "UTF-8",
			indent:        "  ",
			escapeHTML:    false,
			compactOutput: false,
		}

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
					StreamID:     "1234567",
					DateRange:    "2023-01-01 to 2023-01-31",
					RecordIndex:  1,
					TotalRecords: 1,
					OutputFormat: "json",
					ToolVersion:  "ga-tool-v1.0",
				},
			},
		}

		var buf bytes.Buffer
		err := jsonWriter.writeRecords(records, &buf)
		if err != nil {
			t.Fatalf("writeRecords でエラーが発生しました: %v", err)
		}

		// 出力されたJSONの妥当性を検証
		var outputRecords []JSONRecord
		if err := json.Unmarshal(buf.Bytes(), &outputRecords); err != nil {
			t.Fatalf("出力されたJSONの解析に失敗しました: %v", err)
		}

		if len(outputRecords) != 1 {
			t.Errorf("レコード数が一致しません: 期待=1, 実際=%d", len(outputRecords))
		}

		// レコード内容の詳細検証
		record := outputRecords[0]
		if record.Dimensions["date"] != "2023-01-01" {
			t.Errorf("date ディメンションが一致しません: 期待=2023-01-01, 実際=%s", record.Dimensions["date"])
		}
		if record.Metrics["sessions"] != "1250" {
			t.Errorf("sessions メトリクスが一致しません: 期待=1250, 実際=%s", record.Metrics["sessions"])
		}
		if record.Metadata.PropertyID != "987654321" {
			t.Errorf("PropertyIDが一致しません: 期待=987654321, 実際=%s", record.Metadata.PropertyID)
		}
	})

	t.Run("空のレコード配列テスト", func(t *testing.T) {
		jsonWriter := &JSONWriter{
			encoding:      "UTF-8",
			indent:        "  ",
			escapeHTML:    false,
			compactOutput: false,
		}

		var buf bytes.Buffer
		err := jsonWriter.writeRecords([]JSONRecord{}, &buf)
		if err != nil {
			t.Fatalf("空のレコード配列でエラーが発生しました: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "[]" {
			t.Errorf("空の配列が期待されましたが、実際の出力: %s", output)
		}
	})

	t.Run("コンパクト出力テスト", func(t *testing.T) {
		jsonWriter := &JSONWriter{
			encoding:      "UTF-8",
			indent:        "",
			escapeHTML:    false,
			compactOutput: true,
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

		var buf bytes.Buffer
		err := jsonWriter.writeRecords(records, &buf)
		if err != nil {
			t.Fatalf("コンパクト出力でエラーが発生しました: %v", err)
		}

		output := buf.String()
		// コンパクト出力では余分な空白がないことを確認
		if strings.Contains(output, "  ") {
			t.Error("コンパクト出力に余分な空白が含まれています")
		}
	})
}

// TestJSONStructureAndSchemaValidation はJSON構造とスキーマの検証テスト
// 要件4.6, 4.12: ディメンションとメトリクスのキー・バリューペア、メタデータを含む
func TestJSONStructureAndSchemaValidation(t *testing.T) {
	outputService := NewOutputService()

	t.Run("JSON構造の完全性テスト", func(t *testing.T) {
		testData := &analytics.ReportData{
			Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions", "activeUsers", "newUsers"},
			Rows: [][]string{
				{"987654321", "1234567", "2023-01-01", "/home", "1250", "1100", "850"},
				{"987654321", "1234567", "2023-01-01", "/about", "450", "420", "380"},
			},
			Summary: analytics.ReportSummary{
				TotalRows:  2,
				DateRange:  "2023-01-01 to 2023-01-31",
				Properties: []string{"987654321"},
			},
		}

		var buf bytes.Buffer
		err := outputService.WriteJSON(testData, &buf)
		if err != nil {
			t.Fatalf("WriteJSON でエラーが発生しました: %v", err)
		}

		// JSONスキーマの検証
		var records []JSONRecord
		if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
			t.Fatalf("JSON解析に失敗しました: %v", err)
		}

		// 必須フィールドの存在確認
		for i, record := range records {
			t.Run(fmt.Sprintf("レコード%d", i+1), func(t *testing.T) {
				// Dimensionsフィールドの検証
				if record.Dimensions == nil {
					t.Fatal("Dimensionsフィールドがnilです")
				}

				expectedDimensions := []string{"property_id", "stream_id", "date", "pagePath"}
				for _, dim := range expectedDimensions {
					if _, exists := record.Dimensions[dim]; !exists {
						t.Errorf("ディメンション '%s' が存在しません", dim)
					}
				}

				// Metricsフィールドの検証
				if record.Metrics == nil {
					t.Fatal("Metricsフィールドがnilです")
				}

				expectedMetrics := []string{"sessions", "activeUsers", "newUsers"}
				for _, metric := range expectedMetrics {
					if _, exists := record.Metrics[metric]; !exists {
						t.Errorf("メトリクス '%s' が存在しません", metric)
					}
				}

				// Metadataフィールドの検証
				metadata := record.Metadata
				if metadata.RetrievedAt == "" {
					t.Error("RetrievedAtが空です")
				}
				if metadata.PropertyID == "" {
					t.Error("PropertyIDが空です")
				}
				if metadata.DateRange == "" {
					t.Error("DateRangeが空です")
				}
				if metadata.RecordIndex <= 0 {
					t.Error("RecordIndexが無効です")
				}
				if metadata.TotalRecords <= 0 {
					t.Error("TotalRecordsが無効です")
				}
				if metadata.OutputFormat != "json" {
					t.Errorf("OutputFormatが期待値と一致しません: 期待=json, 実際=%s", metadata.OutputFormat)
				}
			})
		}
	})

	t.Run("メタデータの詳細検証", func(t *testing.T) {
		testData := &analytics.ReportData{
			Headers: []string{"property_id", "date", "sessions"},
			Rows: [][]string{
				{"987654321", "2023-01-01", "1250"},
				{"987654321", "2023-01-02", "1180"},
				{"987654321", "2023-01-03", "1320"},
			},
			Summary: analytics.ReportSummary{
				TotalRows:  3,
				DateRange:  "2023-01-01 to 2023-01-03",
				Properties: []string{"987654321"},
			},
		}

		var buf bytes.Buffer
		err := outputService.WriteJSON(testData, &buf)
		if err != nil {
			t.Fatalf("WriteJSON でエラーが発生しました: %v", err)
		}

		var records []JSONRecord
		if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
			t.Fatalf("JSON解析に失敗しました: %v", err)
		}

		// レコードインデックスの連続性確認
		for i, record := range records {
			expectedIndex := i + 1
			if record.Metadata.RecordIndex != expectedIndex {
				t.Errorf("レコード%dのインデックスが不正: 期待=%d, 実際=%d", i, expectedIndex, record.Metadata.RecordIndex)
			}

			// 全レコードで共通のメタデータ確認
			if record.Metadata.TotalRecords != 3 {
				t.Errorf("TotalRecordsが不正: 期待=3, 実際=%d", record.Metadata.TotalRecords)
			}
			if record.Metadata.DateRange != "2023-01-01 to 2023-01-03" {
				t.Errorf("DateRangeが不正: 期待=2023-01-01 to 2023-01-03, 実際=%s", record.Metadata.DateRange)
			}
		}
	})

	t.Run("日時フォーマットの検証", func(t *testing.T) {
		testData := &analytics.ReportData{
			Headers: []string{"date", "sessions"},
			Rows:    [][]string{{"2023-01-01", "1250"}},
			Summary: analytics.ReportSummary{
				TotalRows: 1,
				DateRange: "2023-01-01 to 2023-01-01",
			},
		}

		var buf bytes.Buffer
		err := outputService.WriteJSON(testData, &buf)
		if err != nil {
			t.Fatalf("WriteJSON でエラーが発生しました: %v", err)
		}

		var records []JSONRecord
		if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
			t.Fatalf("JSON解析に失敗しました: %v", err)
		}

		// RFC3339形式の日時検証
		retrievedAt := records[0].Metadata.RetrievedAt
		if _, err := time.Parse(time.RFC3339, retrievedAt); err != nil {
			t.Errorf("RetrievedAtがRFC3339形式ではありません: %s, エラー: %v", retrievedAt, err)
		}
	})
}

// TestCSVJSONOutputComparison はCSVとJSON出力の比較テスト
// 要件4.2, 4.6: デフォルト形式（CSV）の設定、JSON形式での出力
func TestCSVJSONOutputComparison(t *testing.T) {
	outputService := NewOutputService()

	t.Run("同一データでのCSVとJSON出力比較", func(t *testing.T) {
		testData := &analytics.ReportData{
			Headers: []string{"property_id", "date", "pagePath", "sessions", "activeUsers"},
			Rows: [][]string{
				{"987654321", "2023-01-01", "/home", "1250", "1100"},
				{"987654321", "2023-01-01", "/about", "450", "420"},
				{"987654321", "2023-01-02", "/home", "1180", "1050"},
			},
			Summary: analytics.ReportSummary{
				TotalRows:  3,
				DateRange:  "2023-01-01 to 2023-01-02",
				Properties: []string{"987654321"},
			},
		}

		// CSV出力
		var csvBuf bytes.Buffer
		err := outputService.WriteCSV(testData, &csvBuf)
		if err != nil {
			t.Fatalf("CSV出力でエラーが発生しました: %v", err)
		}

		// JSON出力
		var jsonBuf bytes.Buffer
		err = outputService.WriteJSON(testData, &jsonBuf)
		if err != nil {
			t.Fatalf("JSON出力でエラーが発生しました: %v", err)
		}

		// CSVデータの解析
		csvReader := csv.NewReader(strings.NewReader(csvBuf.String()))
		csvRecords, err := csvReader.ReadAll()
		if err != nil {
			t.Fatalf("CSV解析でエラーが発生しました: %v", err)
		}

		// JSONデータの解析
		var jsonRecords []JSONRecord
		if err := json.Unmarshal(jsonBuf.Bytes(), &jsonRecords); err != nil {
			t.Fatalf("JSON解析でエラーが発生しました: %v", err)
		}

		// データ行数の比較（CSVはヘッダー行を含む）
		expectedCSVRows := len(testData.Rows) + 1 // ヘッダー + データ行
		if len(csvRecords) != expectedCSVRows {
			t.Errorf("CSVレコード数が不正: 期待=%d, 実際=%d", expectedCSVRows, len(csvRecords))
		}

		if len(jsonRecords) != len(testData.Rows) {
			t.Errorf("JSONレコード数が不正: 期待=%d, 実際=%d", len(testData.Rows), len(jsonRecords))
		}

		// データ内容の比較
		csvHeaders := csvRecords[0]
		for i := range testData.Rows {
			csvRow := csvRecords[i+1] // ヘッダー行をスキップ
			jsonRecord := jsonRecords[i]

			// 各フィールドの値を比較
			for j, header := range csvHeaders {
				csvValue := csvRow[j]

				var jsonValue string
				if isDimension(header) {
					jsonValue = jsonRecord.Dimensions[header]
				} else {
					jsonValue = jsonRecord.Metrics[header]
				}

				if csvValue != jsonValue {
					t.Errorf("行%d, 列%s の値が一致しません: CSV=%s, JSON=%s", i+1, header, csvValue, jsonValue)
				}
			}
		}
	})

	t.Run("データ型の一貫性テスト", func(t *testing.T) {
		testData := &analytics.ReportData{
			Headers: []string{"date", "sessions", "activeUsers", "averageSessionDuration"},
			Rows: [][]string{
				{"2023-01-01", "1250", "1100", "120.5"},
				{"2023-01-02", "0", "0", "0.0"},
				{"2023-01-03", "999999", "888888", "999.99"},
			},
			Summary: analytics.ReportSummary{
				TotalRows: 3,
				DateRange: "2023-01-01 to 2023-01-03",
			},
		}

		// JSON出力でのデータ型確認
		var jsonBuf bytes.Buffer
		err := outputService.WriteJSON(testData, &jsonBuf)
		if err != nil {
			t.Fatalf("JSON出力でエラーが発生しました: %v", err)
		}

		var jsonRecords []JSONRecord
		if err := json.Unmarshal(jsonBuf.Bytes(), &jsonRecords); err != nil {
			t.Fatalf("JSON解析でエラーが発生しました: %v", err)
		}

		// 全ての値が文字列として保存されていることを確認
		for i, record := range jsonRecords {
			for key, value := range record.Dimensions {
				if reflect.TypeOf(value).Kind() != reflect.String {
					t.Errorf("レコード%d のディメンション %s が文字列ではありません: %T", i+1, key, value)
				}
			}
			for key, value := range record.Metrics {
				if reflect.TypeOf(value).Kind() != reflect.String {
					t.Errorf("レコード%d のメトリクス %s が文字列ではありません: %T", i+1, key, value)
				}
			}
		}
	})

	t.Run("特殊文字とエスケープの比較", func(t *testing.T) {
		testData := &analytics.ReportData{
			Headers: []string{"pagePath", "pageTitle", "sessions"},
			Rows: [][]string{
				{"/path/with,comma", "Title with \"quotes\"", "100"},
				{"/path/with\nnewline", "Title with\ttab", "200"},
				{"/path/with'apostrophe", "Title with & ampersand", "300"},
			},
			Summary: analytics.ReportSummary{
				TotalRows: 3,
				DateRange: "2023-01-01 to 2023-01-01",
			},
		}

		// CSV出力
		var csvBuf bytes.Buffer
		err := outputService.WriteCSV(testData, &csvBuf)
		if err != nil {
			t.Fatalf("CSV出力でエラーが発生しました: %v", err)
		}

		// JSON出力
		var jsonBuf bytes.Buffer
		err = outputService.WriteJSON(testData, &jsonBuf)
		if err != nil {
			t.Fatalf("JSON出力でエラーが発生しました: %v", err)
		}

		// 両方の出力が有効であることを確認
		csvReader := csv.NewReader(strings.NewReader(csvBuf.String()))
		_, err = csvReader.ReadAll()
		if err != nil {
			t.Errorf("特殊文字を含むCSVの解析に失敗しました: %v", err)
		}

		var jsonRecords []JSONRecord
		if err := json.Unmarshal(jsonBuf.Bytes(), &jsonRecords); err != nil {
			t.Errorf("特殊文字を含むJSONの解析に失敗しました: %v", err)
		}

		// 特殊文字が正しく保持されていることを確認
		if len(jsonRecords) > 0 {
			firstRecord := jsonRecords[0]
			if !strings.Contains(firstRecord.Dimensions["pagePath"], ",") {
				t.Error("カンマが正しく保持されていません")
			}
			if !strings.Contains(firstRecord.Dimensions["pageTitle"], "\"") {
				t.Error("引用符が正しく保持されていません")
			}
		}
	})
}

// TestOutputFormatSelection は出力形式選択機能のテスト
// 要件4.2, 4.3, 4.4: デフォルト形式（CSV）の設定、無効な形式指定時のエラーハンドリング
func TestOutputFormatSelection(t *testing.T) {
	t.Run("ParseOutputFormat包括テスト", func(t *testing.T) {
		tests := []struct {
			name           string
			input          string
			expectedFormat OutputFormat
			expectError    bool
			errorContains  string
		}{
			// 有効な形式
			{"小文字CSV", "csv", FormatCSV, false, ""},
			{"大文字CSV", "CSV", FormatCSV, false, ""},
			{"混合ケースCSV", "Csv", FormatCSV, false, ""},
			{"小文字JSON", "json", FormatJSON, false, ""},
			{"大文字JSON", "JSON", FormatJSON, false, ""},
			{"混合ケースJSON", "Json", FormatJSON, false, ""},

			// 空文字列とスペース（デフォルト）
			{"空文字列", "", FormatCSV, false, ""},
			{"スペースのみ", "   ", FormatCSV, false, ""},
			{"前後スペース付きCSV", "  csv  ", FormatCSV, false, ""},
			{"前後スペース付きJSON", "  json  ", FormatJSON, false, ""},

			// 無効な形式
			{"XML形式", "xml", FormatCSV, true, "無効な出力形式"},
			{"TXT形式", "txt", FormatCSV, true, "無効な出力形式"},
			{"数字", "123", FormatCSV, true, "無効な出力形式"},
			{"特殊文字", "csv!", FormatCSV, true, "無効な出力形式"},
			{"部分一致", "csvformat", FormatCSV, true, "無効な出力形式"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				format, err := ParseOutputFormat(tt.input)

				if tt.expectError {
					if err == nil {
						t.Errorf("エラーが期待されましたが、エラーが発生しませんでした")
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
	})

	t.Run("IsValidOutputFormat包括テスト", func(t *testing.T) {
		validFormats := []string{"csv", "CSV", "json", "JSON", "Csv", "Json", "", "  csv  "}
		invalidFormats := []string{"xml", "txt", "pdf", "123", "csv!", "jsonformat"}

		for _, format := range validFormats {
			if !IsValidOutputFormat(format) {
				t.Errorf("有効な形式 '%s' が無効と判定されました", format)
			}
		}

		for _, format := range invalidFormats {
			if IsValidOutputFormat(format) {
				t.Errorf("無効な形式 '%s' が有効と判定されました", format)
			}
		}
	})

	t.Run("GetSupportedFormats一貫性テスト", func(t *testing.T) {
		supportedFormats := GetSupportedFormats()

		// 期待される形式が全て含まれているか確認
		expectedFormats := []string{"csv", "json"}
		for _, expected := range expectedFormats {
			found := false
			for _, supported := range supportedFormats {
				if supported == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("期待される形式 '%s' がサポートされている形式に含まれていません", expected)
			}
		}

		// サポートされている形式が全て有効であることを確認
		for _, format := range supportedFormats {
			if !IsValidOutputFormat(format) {
				t.Errorf("サポートされている形式 '%s' が IsValidOutputFormat で無効と判定されました", format)
			}
		}
	})

	t.Run("OutputFormat.String一貫性テスト", func(t *testing.T) {
		// 定義されている形式の文字列表現確認
		if FormatCSV.String() != "csv" {
			t.Errorf("FormatCSV.String() = %s, 期待値 csv", FormatCSV.String())
		}
		if FormatJSON.String() != "json" {
			t.Errorf("FormatJSON.String() = %s, 期待値 json", FormatJSON.String())
		}

		// 未定義の形式
		unknownFormat := OutputFormat(999)
		if unknownFormat.String() != "unknown" {
			t.Errorf("未定義形式の文字列表現が期待値と一致しません: 期待=unknown, 実際=%s", unknownFormat.String())
		}
	})

	t.Run("デフォルト形式の一貫性テスト", func(t *testing.T) {
		defaultFormat := GetDefaultOutputFormat()

		// デフォルトがCSVであることを確認
		if defaultFormat != FormatCSV {
			t.Errorf("デフォルト形式が期待値と一致しません: 期待=%v, 実際=%v", FormatCSV, defaultFormat)
		}

		// 空文字列での解析結果がデフォルトと一致することを確認
		parsedDefault, err := ParseOutputFormat("")
		if err != nil {
			t.Errorf("空文字列の解析でエラーが発生しました: %v", err)
		}
		if parsedDefault != defaultFormat {
			t.Errorf("空文字列の解析結果がデフォルトと一致しません: 期待=%v, 実際=%v", defaultFormat, parsedDefault)
		}
	})
}

// TestJSONOutputFileOperations はJSON出力のファイル操作テスト
func TestJSONOutputFileOperations(t *testing.T) {
	outputService := NewOutputService()

	t.Run("JSONファイル出力テスト", func(t *testing.T) {
		testData := &analytics.ReportData{
			Headers: []string{"date", "sessions"},
			Rows:    [][]string{{"2023-01-01", "1250"}},
			Summary: analytics.ReportSummary{
				TotalRows: 1,
				DateRange: "2023-01-01 to 2023-01-01",
			},
		}

		// 一時ファイル名を生成
		tempFile := "test_json_output.json"
		defer os.Remove(tempFile)

		err := outputService.WriteToFile(testData, tempFile, FormatJSON)
		if err != nil {
			t.Fatalf("JSONファイル出力でエラーが発生しました: %v", err)
		}

		// ファイルが作成されたことを確認
		if _, err := os.Stat(tempFile); os.IsNotExist(err) {
			t.Fatal("JSONファイルが作成されませんでした")
		}

		// ファイル内容を読み込んで検証
		content, err := os.ReadFile(tempFile)
		if err != nil {
			t.Fatalf("JSONファイルの読み込みに失敗しました: %v", err)
		}

		var records []JSONRecord
		if err := json.Unmarshal(content, &records); err != nil {
			t.Fatalf("JSONファイルの解析に失敗しました: %v", err)
		}

		if len(records) != 1 {
			t.Errorf("JSONレコード数が期待値と一致しません: 期待=1, 実際=%d", len(records))
		}
	})

	t.Run("WriteWithOptionsでのJSON出力テスト", func(t *testing.T) {
		testData := &analytics.ReportData{
			Headers: []string{"date", "pagePath", "sessions"},
			Rows: [][]string{
				{"2023-01-01", "/home", "1250"},
				{"2023-01-01", "/about", "450"},
			},
			Summary: analytics.ReportSummary{
				TotalRows: 2,
				DateRange: "2023-01-01 to 2023-01-01",
			},
		}

		tempFile := "test_json_options.json"
		defer os.Remove(tempFile)

		options := OutputOptions{
			OutputPath:        tempFile,
			Format:           FormatJSON,
			OverwriteExisting: true,
			CreateDirectories: true,
			ShowSummary:      false,
			QuietMode:        true,
			JSONOptions: &JSONWriteOptions{
				Indent:        stringPtr("    "), // 4スペースインデント
				CompactOutput: boolPtr(false),
				EscapeHTML:    boolPtr(false),
			},
		}

		err := outputService.WriteWithOptions(testData, options)
		if err != nil {
			t.Fatalf("WriteWithOptionsでエラーが発生しました: %v", err)
		}

		// ファイル内容を確認
		content, err := os.ReadFile(tempFile)
		if err != nil {
			t.Fatalf("ファイル読み込みに失敗しました: %v", err)
		}

		// 4スペースインデントが適用されていることを確認
		if !strings.Contains(string(content), "    ") {
			t.Error("4スペースインデントが適用されていません")
		}

		// JSONとして有効であることを確認
		var records []JSONRecord
		if err := json.Unmarshal(content, &records); err != nil {
			t.Fatalf("JSON解析に失敗しました: %v", err)
		}

		if len(records) != 2 {
			t.Errorf("レコード数が期待値と一致しません: 期待=2, 実際=%d", len(records))
		}
	})
}

// TestJSONErrorHandling はJSONエラーハンドリングのテスト
func TestJSONErrorHandling(t *testing.T) {
	outputService := NewOutputService()

	t.Run("nilデータでのJSON出力エラー", func(t *testing.T) {
		var buf bytes.Buffer
		err := outputService.WriteJSON(nil, &buf)
		if err == nil {
			t.Error("nilデータでエラーが発生しませんでした")
		}

		expectedMsg := "出力データがnilです"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("期待されたエラーメッセージが含まれていません: 期待=%s, 実際=%s", expectedMsg, err.Error())
		}
	})

	t.Run("不正な列数のデータでのJSON出力", func(t *testing.T) {
		invalidData := &analytics.ReportData{
			Headers: []string{"date", "sessions"},
			Rows: [][]string{
				{"2023-01-01", "1250"},      // 正常
				{"2023-01-02"},              // 列数不足
				{"2023-01-03", "1180", "extra"}, // 列数過多
			},
			Summary: analytics.ReportSummary{
				TotalRows: 3,
				DateRange: "2023-01-01 to 2023-01-03",
			},
		}

		var buf bytes.Buffer
		err := outputService.WriteJSON(invalidData, &buf)
		if err != nil {
			t.Fatalf("不正な列数のデータでエラーが発生しました: %v", err)
		}

		// 不正な行はスキップされ、正常な行のみが出力されることを確認
		var records []JSONRecord
		if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
			t.Fatalf("JSON解析に失敗しました: %v", err)
		}

		// 正常な行（1行目）のみが出力されることを確認
		if len(records) != 1 {
			t.Errorf("出力レコード数が期待値と一致しません: 期待=1, 実際=%d", len(records))
		}
	})

	t.Run("JSONWriter.validateJSONOutput詳細テスト", func(t *testing.T) {
		jsonWriter := &JSONWriter{}

		// 有効なJSONデータ
		validJSON := `[{
			"dimensions": {"date": "2023-01-01"},
			"metrics": {"sessions": "1250"},
			"metadata": {
				"retrieved_at": "2023-02-01T10:30:00Z",
				"record_index": 1,
				"total_records": 1,
				"output_format": "json"
			}
		}]`

		err := jsonWriter.validateJSONOutput([]byte(validJSON))
		if err != nil {
			t.Errorf("有効なJSONでエラーが発生しました: %v", err)
		}

		// 無効なJSONテストケース
		invalidCases := []struct {
			name     string
			jsonData string
			errorMsg string
		}{
			{
				name:     "構文エラー",
				jsonData: `[{"invalid": json}]`,
				errorMsg: "出力されたJSONが無効です",
			},
			{
				name: "dimensions null",
				jsonData: `[{
					"dimensions": null,
					"metrics": {"sessions": "1250"},
					"metadata": {"retrieved_at": "2023-02-01T10:30:00Z"}
				}]`,
				errorMsg: "dimensions が nil です",
			},
			{
				name: "metrics null",
				jsonData: `[{
					"dimensions": {"date": "2023-01-01"},
					"metrics": null,
					"metadata": {"retrieved_at": "2023-02-01T10:30:00Z"}
				}]`,
				errorMsg: "metrics が nil です",
			},
			{
				name: "retrieved_at 空文字",
				jsonData: `[{
					"dimensions": {"date": "2023-01-01"},
					"metrics": {"sessions": "1250"},
					"metadata": {"retrieved_at": ""}
				}]`,
				errorMsg: "retrieved_at が空です",
			},
		}

		for _, tc := range invalidCases {
			t.Run(tc.name, func(t *testing.T) {
				err := jsonWriter.validateJSONOutput([]byte(tc.jsonData))
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				} else if !strings.Contains(err.Error(), tc.errorMsg) {
					t.Errorf("期待されたエラーメッセージが含まれていません: 期待=%s, 実際=%s", tc.errorMsg, err.Error())
				}
			})
		}
	})
}

// TestJSONUTF8Encoding はUTF-8エンコーディングの包括的なテスト
// 要件4.9: UTF-8エンコーディング対応
func TestJSONUTF8Encoding(t *testing.T) {
	outputService := NewOutputService()

	t.Run("多言語文字のJSON出力テスト", func(t *testing.T) {
		testData := &analytics.ReportData{
			Headers: []string{"pagePath", "pageTitle", "country", "sessions"},
			Rows: [][]string{
				{"/ホーム", "ホームページ", "日本", "1250"},
				{"/关于", "关于我们", "中国", "450"},
				{"/о-нас", "О нас", "Россия", "320"},
				{"/à-propos", "À propos", "France", "280"},
				{"/acerca-de", "Acerca de", "España", "190"},
			},
			Summary: analytics.ReportSummary{
				TotalRows: 5,
				DateRange: "2023-01-01 to 2023-01-01",
			},
		}

		var buf bytes.Buffer
		err := outputService.WriteJSON(testData, &buf)
		if err != nil {
			t.Fatalf("多言語文字のJSON出力でエラーが発生しました: %v", err)
		}

		// UTF-8として正しく出力されていることを確認
		output := buf.String()

		// 各言語の文字が含まれていることを確認
		expectedStrings := []string{
			"ホーム", "ホームページ", "日本",
			"关于", "关于我们", "中国",
			"о-нас", "О нас", "Россия",
			"à-propos", "À propos", "France",
			"acerca-de", "Acerca de", "España",
		}

		for _, expected := range expectedStrings {
			if !strings.Contains(output, expected) {
				t.Errorf("期待される文字列 '%s' が出力に含まれていません", expected)
			}
		}

		// JSONとして正しく解析できることを確認
		var records []JSONRecord
		if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
			t.Fatalf("多言語JSONの解析に失敗しました: %v", err)
		}

		// 文字が正しく保持されていることを確認
		if len(records) > 0 {
			firstRecord := records[0]
			if firstRecord.Dimensions["pagePath"] != "/ホーム" {
				t.Errorf("日本語文字が正しく保持されていません: 期待=/ホーム, 実際=%s", firstRecord.Dimensions["pagePath"])
			}
		}
	})

	t.Run("特殊Unicode文字のテスト", func(t *testing.T) {
		testData := &analytics.ReportData{
			Headers: []string{"pagePath", "pageTitle", "sessions"},
			Rows: [][]string{
				{"/emoji", "😀😃😄😁😆😅😂🤣", "100"},
				{"/symbols", "★☆♠♣♥♦♪♫", "200"},
				{"/math", "∑∏∫∆∇∂∞±", "300"},
				{"/arrows", "←↑→↓↔↕⇐⇑⇒⇓", "400"},
			},
			Summary: analytics.ReportSummary{
				TotalRows: 4,
				DateRange: "2023-01-01 to 2023-01-01",
			},
		}

		var buf bytes.Buffer
		err := outputService.WriteJSON(testData, &buf)
		if err != nil {
			t.Fatalf("特殊Unicode文字のJSON出力でエラーが発生しました: %v", err)
		}

		// JSONとして正しく解析できることを確認
		var records []JSONRecord
		if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
			t.Fatalf("特殊Unicode文字JSONの解析に失敗しました: %v", err)
		}

		// 特殊文字が正しく保持されていることを確認
		if len(records) >= 1 {
			emojiRecord := records[0]
			if !strings.Contains(emojiRecord.Dimensions["pageTitle"], "😀") {
				t.Error("絵文字が正しく保持されていません")
			}
		}
	})
}