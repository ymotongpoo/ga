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
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ymotongpoo/ga/internal/analytics"
)

// TestCSVJSONDataConsistency はCSVとJSON出力のデータ一貫性テスト
// 要件4.1, 4.2, 4.6: CSV形式での出力、デフォルト形式（CSV）、JSON形式での出力
func TestCSVJSONDataConsistency(t *testing.T) {
	outputService := NewOutputService()

	t.Run("基本データ一貫性テスト", func(t *testing.T) {
		testData := &analytics.ReportData{
			Headers: []string{"property_id", "date", "pagePath", "sessions", "activeUsers", "newUsers", "averageSessionDuration"},
			Rows: [][]string{
				{"987654321", "2023-01-01", "/home", "1250", "1100", "850", "120.5"},
				{"987654321", "2023-01-01", "/about", "450", "420", "380", "95.2"},
				{"987654321", "2023-01-02", "/home", "1180", "1050", "780", "115.8"},
				{"987654321", "2023-01-02", "/contact", "320", "300", "250", "88.4"},
			},
			Summary: analytics.ReportSummary{
				TotalRows:  4,
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

		// データ行数の検証
		csvDataRows := csvRecords[1:] // ヘッダー行を除く
		if len(csvDataRows) != len(jsonRecords) {
			t.Errorf("データ行数が一致しません: CSV=%d, JSON=%d", len(csvDataRows), len(jsonRecords))
		}

		if len(jsonRecords) != len(testData.Rows) {
			t.Errorf("JSONレコード数が元データと一致しません: 期待=%d, 実際=%d", len(testData.Rows), len(jsonRecords))
		}

		// ヘッダーの検証
		csvHeaders := csvRecords[0]
		if len(csvHeaders) != len(testData.Headers) {
			t.Errorf("CSVヘッダー数が一致しません: 期待=%d, 実際=%d", len(testData.Headers), len(csvHeaders))
		}

		// 各データ行の値を比較
		for i, csvRow := range csvDataRows {
			if i >= len(jsonRecords) {
				break
			}

			jsonRecord := jsonRecords[i]

			for j, header := range csvHeaders {
				csvValue := csvRow[j]

				var jsonValue string
				var exists bool

				if isDimension(header) {
					jsonValue, exists = jsonRecord.Dimensions[header]
				} else {
					jsonValue, exists = jsonRecord.Metrics[header]
				}

				if !exists {
					t.Errorf("行%d: JSONに列 '%s' が存在しません", i+1, header)
					continue
				}

				if csvValue != jsonValue {
					t.Errorf("行%d, 列%s: 値が一致しません: CSV='%s', JSON='%s'", i+1, header, csvValue, jsonValue)
				}
			}
		}
	})

	t.Run("特殊文字データ一貫性テスト", func(t *testing.T) {
		specialCharData := &analytics.ReportData{
			Headers: []string{"pagePath", "pageTitle", "country", "sessions"},
			Rows: [][]string{
				{"/path,with,commas", "Title \"with quotes\"", "Japan", "100"},
				{"/path\nwith\nnewlines", "Title\twith\ttabs", "USA", "200"},
				{"/path'with'apostrophes", "Title & with & ampersands", "UK", "300"},
				{"/path;with;semicolons", "Title | with | pipes", "Germany", "400"},
				{"/path with spaces", "Title (with) parentheses", "France", "500"},
			},
			Summary: analytics.ReportSummary{
				TotalRows: 5,
				DateRange: "2023-01-01 to 2023-01-01",
			},
		}

		// CSV出力
		var csvBuf bytes.Buffer
		err := outputService.WriteCSV(specialCharData, &csvBuf)
		if err != nil {
			t.Fatalf("特殊文字データのCSV出力でエラーが発生しました: %v", err)
		}

		// JSON出力
		var jsonBuf bytes.Buffer
		err = outputService.WriteJSON(specialCharData, &jsonBuf)
		if err != nil {
			t.Fatalf("特殊文字データのJSON出力でエラーが発生しました: %v", err)
		}

		// CSVデータの解析
		csvReader := csv.NewReader(strings.NewReader(csvBuf.String()))
		csvRecords, err := csvReader.ReadAll()
		if err != nil {
			t.Fatalf("特殊文字CSVの解析でエラーが発生しました: %v", err)
		}

		// JSONデータの解析
		var jsonRecords []JSONRecord
		if err := json.Unmarshal(jsonBuf.Bytes(), &jsonRecords); err != nil {
			t.Fatalf("特殊文字JSONの解析でエラーが発生しました: %v", err)
		}

		// 特殊文字が正しく保持されているかを確認
		csvHeaders := csvRecords[0]
		csvDataRows := csvRecords[1:]

		for i, csvRow := range csvDataRows {
			if i >= len(jsonRecords) {
				break
			}

			jsonRecord := jsonRecords[i]

			for j, header := range csvHeaders {
				csvValue := csvRow[j]

				var jsonValue string
				if isDimension(header) {
					jsonValue = jsonRecord.Dimensions[header]
				} else {
					jsonValue = jsonRecord.Metrics[header]
				}

				if csvValue != jsonValue {
					t.Errorf("特殊文字行%d, 列%s: 値が一致しません: CSV='%s', JSON='%s'", i+1, header, csvValue, jsonValue)
				}
			}
		}

		// 特定の特殊文字が保持されていることを確認
		if len(jsonRecords) > 0 {
			firstRecord := jsonRecords[0]

			// カンマが含まれるパス
			if !strings.Contains(firstRecord.Dimensions["pagePath"], ",") {
				t.Error("カンマが正しく保持されていません")
			}

			// 引用符が含まれるタイトル
			if !strings.Contains(firstRecord.Dimensions["pageTitle"], "\"") {
				t.Error("引用符が正しく保持されていません")
			}
		}
	})

	t.Run("数値データ型一貫性テスト", func(t *testing.T) {
		numericData := &analytics.ReportData{
			Headers: []string{"date", "sessions", "activeUsers", "averageSessionDuration", "bounceRate"},
			Rows: [][]string{
				{"2023-01-01", "0", "0", "0.0", "0.00"},
				{"2023-01-02", "1", "1", "1.5", "100.00"},
				{"2023-01-03", "999999", "888888", "999.99", "50.25"},
				{"2023-01-04", "1250", "1100", "120.5", "25.75"},
			},
			Summary: analytics.ReportSummary{
				TotalRows: 4,
				DateRange: "2023-01-01 to 2023-01-04",
			},
		}

		// CSV出力
		var csvBuf bytes.Buffer
		err := outputService.WriteCSV(numericData, &csvBuf)
		if err != nil {
			t.Fatalf("数値データのCSV出力でエラーが発生しました: %v", err)
		}

		// JSON出力
		var jsonBuf bytes.Buffer
		err = outputService.WriteJSON(numericData, &jsonBuf)
		if err != nil {
			t.Fatalf("数値データのJSON出力でエラーが発生しました: %v", err)
		}

		// JSONデータの解析
		var jsonRecords []JSONRecord
		if err := json.Unmarshal(jsonBuf.Bytes(), &jsonRecords); err != nil {
			t.Fatalf("数値JSONの解析でエラーが発生しました: %v", err)
		}

		// 全ての数値が文字列として保存されていることを確認
		for i, record := range jsonRecords {
			// ディメンションの型確認
			for key, value := range record.Dimensions {
				if reflect.TypeOf(value).Kind() != reflect.String {
					t.Errorf("レコード%d のディメンション %s が文字列ではありません: %T", i+1, key, value)
				}
			}

			// メトリクスの型確認
			for key, value := range record.Metrics {
				if reflect.TypeOf(value).Kind() != reflect.String {
					t.Errorf("レコード%d のメトリクス %s が文字列ではありません: %T", i+1, key, value)
				}
			}
		}

		// 数値の精度が保持されていることを確認
		csvReader := csv.NewReader(strings.NewReader(csvBuf.String()))
		csvRecords, err := csvReader.ReadAll()
		if err != nil {
			t.Fatalf("数値CSVの解析でエラーが発生しました: %v", err)
		}

		csvHeaders := csvRecords[0]
		csvDataRows := csvRecords[1:]

		for i, csvRow := range csvDataRows {
			if i >= len(jsonRecords) {
				break
			}

			jsonRecord := jsonRecords[i]

			for j, header := range csvHeaders {
				csvValue := csvRow[j]

				var jsonValue string
				if isDimension(header) {
					jsonValue = jsonRecord.Dimensions[header]
				} else {
					jsonValue = jsonRecord.Metrics[header]
				}

				if csvValue != jsonValue {
					t.Errorf("数値行%d, 列%s: 精度が一致しません: CSV='%s', JSON='%s'", i+1, header, csvValue, jsonValue)
				}
			}
		}
	})

	t.Run("空値とnull値の一貫性テスト", func(t *testing.T) {
		emptyValueData := &analytics.ReportData{
			Headers: []string{"date", "pagePath", "pageTitle", "sessions"},
			Rows: [][]string{
				{"2023-01-01", "/home", "Home Page", "1250"},
				{"2023-01-01", "", "Empty Path", "450"},
				{"2023-01-01", "/about", "", "320"},
				{"2023-01-01", "", "", "180"},
			},
			Summary: analytics.ReportSummary{
				TotalRows: 4,
				DateRange: "2023-01-01 to 2023-01-01",
			},
		}

		// CSV出力
		var csvBuf bytes.Buffer
		err := outputService.WriteCSV(emptyValueData, &csvBuf)
		if err != nil {
			t.Fatalf("空値データのCSV出力でエラーが発生しました: %v", err)
		}

		// JSON出力
		var jsonBuf bytes.Buffer
		err = outputService.WriteJSON(emptyValueData, &jsonBuf)
		if err != nil {
			t.Fatalf("空値データのJSON出力でエラーが発生しました: %v", err)
		}

		// CSVデータの解析
		csvReader := csv.NewReader(strings.NewReader(csvBuf.String()))
		csvRecords, err := csvReader.ReadAll()
		if err != nil {
			t.Fatalf("空値CSVの解析でエラーが発生しました: %v", err)
		}

		// JSONデータの解析
		var jsonRecords []JSONRecord
		if err := json.Unmarshal(jsonBuf.Bytes(), &jsonRecords); err != nil {
			t.Fatalf("空値JSONの解析でエラーが発生しました: %v", err)
		}

		// 空値の処理が一貫していることを確認
		csvHeaders := csvRecords[0]
		csvDataRows := csvRecords[1:]

		for i, csvRow := range csvDataRows {
			if i >= len(jsonRecords) {
				break
			}

			jsonRecord := jsonRecords[i]

			for j, header := range csvHeaders {
				csvValue := csvRow[j]

				var jsonValue string
				if isDimension(header) {
					jsonValue = jsonRecord.Dimensions[header]
				} else {
					jsonValue = jsonRecord.Metrics[header]
				}

				// 空文字列が正しく保持されていることを確認
				if csvValue != jsonValue {
					t.Errorf("空値行%d, 列%s: 値が一致しません: CSV='%s', JSON='%s'", i+1, header, csvValue, jsonValue)
				}
			}
		}
	})
}

// TestCSVJSONStructuralDifferences はCSVとJSONの構造的違いのテスト
func TestCSVJSONStructuralDifferences(t *testing.T) {
	outputService := NewOutputService()

	testData := &analytics.ReportData{
		Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions", "activeUsers"},
		Rows: [][]string{
			{"987654321", "1234567", "2023-01-01", "/home", "1250", "1100"},
			{"987654321", "1234567", "2023-01-01", "/about", "450", "420"},
		},
		Summary: analytics.ReportSummary{
			TotalRows:  2,
			DateRange:  "2023-01-01 to 2023-01-01",
			Properties: []string{"987654321"},
		},
	}

	t.Run("CSV構造特性テスト", func(t *testing.T) {
		var csvBuf bytes.Buffer
		err := outputService.WriteCSV(testData, &csvBuf)
		if err != nil {
			t.Fatalf("CSV出力でエラーが発生しました: %v", err)
		}

		csvOutput := csvBuf.String()
		lines := strings.Split(strings.TrimSpace(csvOutput), "\n")

		// CSVの特性確認
		// 1. ヘッダー行が存在する
		if len(lines) == 0 {
			t.Fatal("CSV出力が空です")
		}

		headerLine := lines[0]
		expectedHeaders := strings.Join(testData.Headers, ",")
		if headerLine != expectedHeaders {
			t.Errorf("CSVヘッダーが期待値と一致しません: 期待=%s, 実際=%s", expectedHeaders, headerLine)
		}

		// 2. データ行数が正しい（ヘッダー + データ行）
		expectedLines := len(testData.Rows) + 1
		if len(lines) != expectedLines {
			t.Errorf("CSV行数が期待値と一致しません: 期待=%d, 実際=%d", expectedLines, len(lines))
		}

		// 3. 各行の列数が一致する
		for i, line := range lines {
			fields := strings.Split(line, ",")
			if len(fields) != len(testData.Headers) {
				t.Errorf("CSV行%d の列数が期待値と一致しません: 期待=%d, 実際=%d", i+1, len(testData.Headers), len(fields))
			}
		}

		// 4. フラットな構造（ネストなし）
		if strings.Contains(csvOutput, "{") || strings.Contains(csvOutput, "}") {
			t.Error("CSV出力にJSON構造が含まれています")
		}
	})

	t.Run("JSON構造特性テスト", func(t *testing.T) {
		var jsonBuf bytes.Buffer
		err := outputService.WriteJSON(testData, &jsonBuf)
		if err != nil {
			t.Fatalf("JSON出力でエラーが発生しました: %v", err)
		}

		jsonOutput := jsonBuf.String()

		// JSONの特性確認
		// 1. 有効なJSON配列である
		var jsonRecords []JSONRecord
		if err := json.Unmarshal(jsonBuf.Bytes(), &jsonRecords); err != nil {
			t.Fatalf("JSON解析に失敗しました: %v", err)
		}

		// 2. 配列の要素数が正しい
		if len(jsonRecords) != len(testData.Rows) {
			t.Errorf("JSONレコード数が期待値と一致しません: 期待=%d, 実際=%d", len(testData.Rows), len(jsonRecords))
		}

		// 3. 構造化されたデータ（dimensions, metrics, metadata）
		for i, record := range jsonRecords {
			if record.Dimensions == nil {
				t.Errorf("レコード%d のDimensionsがnilです", i+1)
			}
			if record.Metrics == nil {
				t.Errorf("レコード%d のMetricsがnilです", i+1)
			}
			if record.Metadata.RetrievedAt == "" {
				t.Errorf("レコード%d のMetadata.RetrievedAtが空です", i+1)
			}
		}

		// 4. メタデータが含まれている
		if len(jsonRecords) > 0 {
			metadata := jsonRecords[0].Metadata
			if metadata.PropertyID == "" {
				t.Error("PropertyIDが設定されていません")
			}
			if metadata.DateRange == "" {
				t.Error("DateRangeが設定されていません")
			}
			if metadata.OutputFormat != "json" {
				t.Errorf("OutputFormatが期待値と一致しません: 期待=json, 実際=%s", metadata.OutputFormat)
			}
		}

		// 5. ネストした構造である
		if !strings.Contains(jsonOutput, "{") || !strings.Contains(jsonOutput, "}") {
			t.Error("JSON出力にオブジェクト構造が含まれていません")
		}
	})

	t.Run("データアクセス方法の違いテスト", func(t *testing.T) {
		// CSV: インデックスベースのアクセス
		var csvBuf bytes.Buffer
		err := outputService.WriteCSV(testData, &csvBuf)
		if err != nil {
			t.Fatalf("CSV出力でエラーが発生しました: %v", err)
		}

		csvReader := csv.NewReader(strings.NewReader(csvBuf.String()))
		csvRecords, err := csvReader.ReadAll()
		if err != nil {
			t.Fatalf("CSV解析でエラーが発生しました: %v", err)
		}

		// JSON: キーベースのアクセス
		var jsonBuf bytes.Buffer
		err = outputService.WriteJSON(testData, &jsonBuf)
		if err != nil {
			t.Fatalf("JSON出力でエラーが発生しました: %v", err)
		}

		var jsonRecords []JSONRecord
		if err := json.Unmarshal(jsonBuf.Bytes(), &jsonRecords); err != nil {
			t.Fatalf("JSON解析でエラーが発生しました: %v", err)
		}

		// 同じデータに異なる方法でアクセス
		if len(csvRecords) > 1 && len(jsonRecords) > 0 {
			csvHeaders := csvRecords[0]
			csvFirstRow := csvRecords[1]
			jsonFirstRecord := jsonRecords[0]

			// CSVでのアクセス（インデックス）
			var csvSessionsValue string
			for i, header := range csvHeaders {
				if header == "sessions" {
					csvSessionsValue = csvFirstRow[i]
					break
				}
			}

			// JSONでのアクセス（キー）
			jsonSessionsValue := jsonFirstRecord.Metrics["sessions"]

			// 値が一致することを確認
			if csvSessionsValue != jsonSessionsValue {
				t.Errorf("アクセス方法による値の違い: CSV=%s, JSON=%s", csvSessionsValue, jsonSessionsValue)
			}
		}
	})

	t.Run("メタデータの有無比較テスト", func(t *testing.T) {
		var csvBuf bytes.Buffer
		err := outputService.WriteCSV(testData, &csvBuf)
		if err != nil {
			t.Fatalf("CSV出力でエラーが発生しました: %v", err)
		}

		var jsonBuf bytes.Buffer
		err = outputService.WriteJSON(testData, &jsonBuf)
		if err != nil {
			t.Fatalf("JSON出力でエラーが発生しました: %v", err)
		}

		// CSVにはメタデータが含まれない
		csvOutput := csvBuf.String()
		metadataKeywords := []string{"retrieved_at", "date_range", "record_index", "total_records"}

		for _, keyword := range metadataKeywords {
			if strings.Contains(csvOutput, keyword) {
				t.Errorf("CSV出力にメタデータキーワード '%s' が含まれています", keyword)
			}
		}

		// JSONにはメタデータが含まれる
		var jsonRecords []JSONRecord
		if err := json.Unmarshal(jsonBuf.Bytes(), &jsonRecords); err != nil {
			t.Fatalf("JSON解析でエラーが発生しました: %v", err)
		}

		if len(jsonRecords) > 0 {
			metadata := jsonRecords[0].Metadata

			// 必須メタデータフィールドの存在確認
			if metadata.RetrievedAt == "" {
				t.Error("JSONにRetrievedAtメタデータが含まれていません")
			}
			if metadata.DateRange == "" {
				t.Error("JSONにDateRangeメタデータが含まれていません")
			}
			if metadata.RecordIndex <= 0 {
				t.Error("JSONにRecordIndexメタデータが含まれていません")
			}
			if metadata.TotalRecords <= 0 {
				t.Error("JSONにTotalRecordsメタデータが含まれていません")
			}
		}
	})
}

// TestCSVJSONPerformanceComparison はCSVとJSON出力のパフォーマンス比較テスト
func TestCSVJSONPerformanceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("パフォーマンステストをスキップします（-short フラグが指定されています）")
	}

	outputService := NewOutputService()

	// 様々なサイズのテストデータを作成
	testSizes := []struct {
		name string
		rows int
		cols int
	}{
		{"小規模", 100, 10},
		{"中規模", 1000, 15},
		{"大規模", 5000, 20},
	}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("%sデータ性能比較", size.name), func(t *testing.T) {
			// テストデータの生成
			headers := make([]string, size.cols)
			for i := 0; i < size.cols; i++ {
				if i < size.cols/2 {
					headers[i] = fmt.Sprintf("dimension_%d", i+1)
				} else {
					headers[i] = fmt.Sprintf("metric_%d", i+1)
				}
			}

			rows := make([][]string, size.rows)
			for i := 0; i < size.rows; i++ {
				row := make([]string, size.cols)
				for j := 0; j < size.cols; j++ {
					row[j] = fmt.Sprintf("value_%d_%d", i+1, j+1)
				}
				rows[i] = row
			}

			testData := &analytics.ReportData{
				Headers: headers,
				Rows:    rows,
				Summary: analytics.ReportSummary{
					TotalRows: size.rows,
					DateRange: "2023-01-01 to 2023-12-31",
				},
			}

			// CSV出力の測定
			csvStart := time.Now()
			var csvBuf bytes.Buffer
			err := outputService.WriteCSV(testData, &csvBuf)
			csvDuration := time.Since(csvStart)

			if err != nil {
				t.Fatalf("CSV出力でエラーが発生しました: %v", err)
			}

			// JSON出力の測定
			jsonStart := time.Now()
			var jsonBuf bytes.Buffer
			err = outputService.WriteJSON(testData, &jsonBuf)
			jsonDuration := time.Since(jsonStart)

			if err != nil {
				t.Fatalf("JSON出力でエラーが発生しました: %v", err)
			}

			// 結果の記録
			t.Logf("%s - CSV出力時間: %v", size.name, csvDuration)
			t.Logf("%s - JSON出力時間: %v", size.name, jsonDuration)
			t.Logf("%s - CSVサイズ: %d bytes", size.name, csvBuf.Len())
			t.Logf("%s - JSONサイズ: %d bytes", size.name, jsonBuf.Len())

			// 性能比較
			ratio := float64(jsonDuration) / float64(csvDuration)
			t.Logf("%s - JSON/CSV時間比: %.2f", size.name, ratio)

			sizeRatio := float64(jsonBuf.Len()) / float64(csvBuf.Len())
			t.Logf("%s - JSON/CSVサイズ比: %.2f", size.name, sizeRatio)

			// 極端な性能差の警告
			if ratio > 5.0 {
				t.Logf("警告: JSONがCSVより5倍以上遅いです（%s）", size.name)
			}
			if ratio < 0.2 {
				t.Logf("警告: CSVがJSONより5倍以上遅いです（%s）", size.name)
			}
		})
	}
}

// TestCSVJSONUseCaseScenarios はCSVとJSONの使用ケースシナリオテスト
func TestCSVJSONUseCaseScenarios(t *testing.T) {
	outputService := NewOutputService()

	t.Run("データ分析ツール連携シナリオ", func(t *testing.T) {
		// Excel/スプレッドシート向けのCSV出力
		analyticsData := &analytics.ReportData{
			Headers: []string{"date", "pagePath", "sessions", "activeUsers", "bounceRate"},
			Rows: [][]string{
				{"2023-01-01", "/home", "1250", "1100", "25.5"},
				{"2023-01-01", "/about", "450", "420", "30.2"},
				{"2023-01-02", "/home", "1180", "1050", "28.1"},
			},
			Summary: analytics.ReportSummary{
				TotalRows: 3,
				DateRange: "2023-01-01 to 2023-01-02",
			},
		}

		// CSV出力（スプレッドシート向け）
		var csvBuf bytes.Buffer
		err := outputService.WriteCSV(analyticsData, &csvBuf)
		if err != nil {
			t.Fatalf("分析ツール向けCSV出力でエラーが発生しました: %v", err)
		}

		// CSVが標準的な形式であることを確認
		csvOutput := csvBuf.String()
		lines := strings.Split(strings.TrimSpace(csvOutput), "\n")

		// ヘッダー行の確認
		if !strings.Contains(lines[0], "date,pagePath,sessions") {
			t.Error("分析ツール向けCSVヘッダーが正しくありません")
		}

		// 数値データが適切に出力されていることを確認
		if !strings.Contains(csvOutput, "1250") || !strings.Contains(csvOutput, "25.5") {
			t.Error("数値データが正しく出力されていません")
		}
	})

	t.Run("API/プログラム連携シナリオ", func(t *testing.T) {
		// プログラム処理向けのJSON出力
		apiData := &analytics.ReportData{
			Headers: []string{"property_id", "date", "pagePath", "sessions", "activeUsers"},
			Rows: [][]string{
				{"987654321", "2023-01-01", "/api/v1/users", "500", "450"},
				{"987654321", "2023-01-01", "/api/v1/orders", "300", "280"},
			},
			Summary: analytics.ReportSummary{
				TotalRows: 2,
				DateRange: "2023-01-01 to 2023-01-01",
			},
		}

		// JSON出力（API向け）
		var jsonBuf bytes.Buffer
		err := outputService.WriteJSON(apiData, &jsonBuf)
		if err != nil {
			t.Fatalf("API向けJSON出力でエラーが発生しました: %v", err)
		}

		// JSONが構造化されていることを確認
		var jsonRecords []JSONRecord
		if err := json.Unmarshal(jsonBuf.Bytes(), &jsonRecords); err != nil {
			t.Fatalf("API向けJSON解析に失敗しました: %v", err)
		}

		// API向けの構造確認
		if len(jsonRecords) > 0 {
			record := jsonRecords[0]

			// ディメンションとメトリクスが分離されていることを確認
			if len(record.Dimensions) == 0 {
				t.Error("API向けJSONにディメンションが含まれていません")
			}
			if len(record.Metrics) == 0 {
				t.Error("API向けJSONにメトリクスが含まれていません")
			}

			// メタデータが含まれていることを確認
			if record.Metadata.PropertyID == "" {
				t.Error("API向けJSONにプロパティIDが含まれていません")
			}
			if record.Metadata.RetrievedAt == "" {
				t.Error("API向けJSONに取得日時が含まれていません")
			}
		}
	})

	t.Run("レポート生成シナリオ", func(t *testing.T) {
		// レポート向けの詳細データ
		reportData := &analytics.ReportData{
			Headers: []string{"date", "country", "city", "sessions", "activeUsers", "newUsers", "averageSessionDuration"},
			Rows: [][]string{
				{"2023-01-01", "Japan", "Tokyo", "5000", "4500", "3000", "180.5"},
				{"2023-01-01", "Japan", "Osaka", "2000", "1800", "1200", "165.2"},
				{"2023-01-01", "USA", "New York", "3000", "2700", "1800", "195.8"},
			},
			Summary: analytics.ReportSummary{
				TotalRows: 3,
				DateRange: "2023-01-01 to 2023-01-01",
			},
		}

		// 両形式での出力
		var csvBuf, jsonBuf bytes.Buffer

		err := outputService.WriteCSV(reportData, &csvBuf)
		if err != nil {
			t.Fatalf("レポート向けCSV出力でエラーが発生しました: %v", err)
		}

		err = outputService.WriteJSON(reportData, &jsonBuf)
		if err != nil {
			t.Fatalf("レポート向けJSON出力でエラーが発生しました: %v", err)
		}

		// CSV: 表形式での可読性確認
		csvLines := strings.Split(strings.TrimSpace(csvBuf.String()), "\n")
		if len(csvLines) != 4 { // ヘッダー + 3データ行
			t.Errorf("レポート向けCSV行数が期待値と一致しません: 期待=4, 実際=%d", len(csvLines))
		}

		// JSON: 構造化データでの詳細情報確認
		var jsonRecords []JSONRecord
		if err := json.Unmarshal(jsonBuf.Bytes(), &jsonRecords); err != nil {
			t.Fatalf("レポート向けJSON解析に失敗しました: %v", err)
		}

		if len(jsonRecords) != 3 {
			t.Errorf("レポート向けJSONレコード数が期待値と一致しません: 期待=3, 実際=%d", len(jsonRecords))
		}

		// メタデータによる追加情報の確認
		if len(jsonRecords) > 0 {
			metadata := jsonRecords[0].Metadata
			if metadata.DateRange != "2023-01-01 to 2023-01-01" {
				t.Error("レポート向けJSONの期間情報が正しくありません")
			}
			if metadata.TotalRecords != 3 {
				t.Error("レポート向けJSONの総レコード数が正しくありません")
			}
		}
	})
}