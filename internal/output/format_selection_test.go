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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ymotongpoo/ga/internal/analytics"
)

// TestOutputFormatSelectionIntegration は出力形式選択機能の統合テスト
// 要件4.2, 4.3, 4.4, 4.7, 4.8: 出力形式選択、エラーハンドリング、統合出力サービス
func TestOutputFormatSelectionIntegration(t *testing.T) {
	outputService := NewOutputService()

	// 共通テストデータ
	testData := &analytics.ReportData{
		Headers: []string{"property_id", "date", "pagePath", "sessions", "activeUsers", "newUsers"},
		Rows: [][]string{
			{"987654321", "2023-01-01", "/home", "1250", "1100", "850"},
			{"987654321", "2023-01-01", "/about", "450", "420", "380"},
			{"987654321", "2023-01-02", "/home", "1180", "1050", "780"},
		},
		Summary: analytics.ReportSummary{
			TotalRows:  3,
			DateRange:  "2023-01-01 to 2023-01-02",
			Properties: []string{"987654321"},
		},
	}

	t.Run("WriteOutput形式選択テスト", func(t *testing.T) {
		tests := []struct {
			name           string
			format         OutputFormat
			outputPath     string
			validateOutput func(t *testing.T, outputPath string, format OutputFormat)
		}{
			{
				name:       "CSV標準出力",
				format:     FormatCSV,
				outputPath: "",
				validateOutput: func(t *testing.T, outputPath string, format OutputFormat) {
					// 標準出力のテストは別途実装
				},
			},
			{
				name:       "JSON標準出力",
				format:     FormatJSON,
				outputPath: "",
				validateOutput: func(t *testing.T, outputPath string, format OutputFormat) {
					// 標準出力のテストは別途実装
				},
			},
			{
				name:       "CSVファイル出力",
				format:     FormatCSV,
				outputPath: "test_format_csv.csv",
				validateOutput: func(t *testing.T, outputPath string, format OutputFormat) {
					defer os.Remove(outputPath)

					// ファイルが作成されたことを確認
					if _, err := os.Stat(outputPath); os.IsNotExist(err) {
						t.Fatalf("出力ファイルが作成されませんでした: %s", outputPath)
					}

					// ファイル内容を読み込み
					content, err := os.ReadFile(outputPath)
					if err != nil {
						t.Fatalf("ファイル読み込みに失敗しました: %v", err)
					}

					// CSVとして解析
					csvReader := csv.NewReader(strings.NewReader(string(content)))
					records, err := csvReader.ReadAll()
					if err != nil {
						t.Fatalf("CSV解析に失敗しました: %v", err)
					}

					// ヘッダー + データ行の確認
					expectedRows := len(testData.Rows) + 1
					if len(records) != expectedRows {
						t.Errorf("CSV行数が期待値と一致しません: 期待=%d, 実際=%d", expectedRows, len(records))
					}

					// ヘッダー確認
					expectedHeader := "property_id,date,pagePath,sessions,activeUsers,newUsers"
					actualHeader := strings.Join(records[0], ",")
					if actualHeader != expectedHeader {
						t.Errorf("CSVヘッダーが期待値と一致しません: 期待=%s, 実際=%s", expectedHeader, actualHeader)
					}
				},
			},
			{
				name:       "JSONファイル出力",
				format:     FormatJSON,
				outputPath: "test_format_json.json",
				validateOutput: func(t *testing.T, outputPath string, format OutputFormat) {
					defer os.Remove(outputPath)

					// ファイルが作成されたことを確認
					if _, err := os.Stat(outputPath); os.IsNotExist(err) {
						t.Fatalf("出力ファイルが作成されませんでした: %s", outputPath)
					}

					// ファイル内容を読み込み
					content, err := os.ReadFile(outputPath)
					if err != nil {
						t.Fatalf("ファイル読み込みに失敗しました: %v", err)
					}

					// JSONとして解析
					var records []JSONRecord
					if err := json.Unmarshal(content, &records); err != nil {
						t.Fatalf("JSON解析に失敗しました: %v", err)
					}

					// レコード数の確認
					if len(records) != len(testData.Rows) {
						t.Errorf("JSONレコード数が期待値と一致しません: 期待=%d, 実際=%d", len(testData.Rows), len(records))
					}

					// 最初のレコードの構造確認
					if len(records) > 0 {
						record := records[0]
						if record.Dimensions == nil {
							t.Error("Dimensionsフィールドがnilです")
						}
						if record.Metrics == nil {
							t.Error("Metricsフィールドがnilです")
						}
						if record.Metadata.OutputFormat != "json" {
							t.Errorf("OutputFormatが期待値と一致しません: 期待=json, 実際=%s", record.Metadata.OutputFormat)
						}
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := outputService.WriteOutput(testData, tt.outputPath, tt.format)
				if err != nil {
					t.Fatalf("WriteOutput でエラーが発生しました: %v", err)
				}

				if tt.outputPath != "" {
					tt.validateOutput(t, tt.outputPath, tt.format)
				}
			})
		}
	})

	t.Run("WriteWithOptions詳細形式選択テスト", func(t *testing.T) {
		tests := []struct {
			name    string
			options OutputOptions
			verify  func(t *testing.T, outputPath string)
		}{
			{
				name: "CSV詳細オプション",
				options: OutputOptions{
					OutputPath:        "test_csv_detailed.csv",
					Format:            FormatCSV,
					OverwriteExisting: true,
					CreateDirectories: true,
					ShowSummary:       false,
					QuietMode:         true,
					CSVOptions: &CSVWriteOptions{
						Delimiter:     ';',
						IncludeHeader: true,
						Encoding:      "UTF-8",
					},
				},
				verify: func(t *testing.T, outputPath string) {
					defer os.Remove(outputPath)

					content, err := os.ReadFile(outputPath)
					if err != nil {
						t.Fatalf("ファイル読み込みに失敗しました: %v", err)
					}

					// セミコロンデリミタが使用されていることを確認
					if !strings.Contains(string(content), ";") {
						t.Error("セミコロンデリミタが使用されていません")
					}

					// CSVとして解析（セミコロンデリミタ）
					csvReader := csv.NewReader(strings.NewReader(string(content)))
					csvReader.Comma = ';'
					records, err := csvReader.ReadAll()
					if err != nil {
						t.Fatalf("セミコロンCSVの解析に失敗しました: %v", err)
					}

					if len(records) == 0 {
						t.Fatal("CSVレコードが空です")
					}

					// ヘッダーが含まれていることを確認
					if len(records[0]) != len(testData.Headers) {
						t.Errorf("ヘッダー列数が期待値と一致しません: 期待=%d, 実際=%d", len(testData.Headers), len(records[0]))
					}
				},
			},
			{
				name: "JSON詳細オプション",
				options: OutputOptions{
					OutputPath:        "test_json_detailed.json",
					Format:            FormatJSON,
					OverwriteExisting: true,
					CreateDirectories: true,
					ShowSummary:       false,
					QuietMode:         true,
					JSONOptions: &JSONWriteOptions{
						Indent:        stringPtr("\t"),
						CompactOutput: boolPtr(false),
						EscapeHTML:    boolPtr(false),
					},
				},
				verify: func(t *testing.T, outputPath string) {
					defer os.Remove(outputPath)

					content, err := os.ReadFile(outputPath)
					if err != nil {
						t.Fatalf("ファイル読み込みに失敗しました: %v", err)
					}

					// タブインデントが使用されていることを確認
					if !strings.Contains(string(content), "\t") {
						t.Error("タブインデントが使用されていません")
					}

					// JSONとして解析
					var records []JSONRecord
					if err := json.Unmarshal(content, &records); err != nil {
						t.Fatalf("JSON解析に失敗しました: %v", err)
					}

					if len(records) != len(testData.Rows) {
						t.Errorf("JSONレコード数が期待値と一致しません: 期待=%d, 実際=%d", len(testData.Rows), len(records))
					}
				},
			},
			{
				name: "ディレクトリ作成テスト",
				options: OutputOptions{
					OutputPath:        "test_dir/subdir/output.json",
					Format:            FormatJSON,
					OverwriteExisting: true,
					CreateDirectories: true,
					ShowSummary:       false,
					QuietMode:         true,
				},
				verify: func(t *testing.T, outputPath string) {
					defer os.RemoveAll("test_dir")

					// ディレクトリが作成されたことを確認
					dir := filepath.Dir(outputPath)
					if _, err := os.Stat(dir); os.IsNotExist(err) {
						t.Fatalf("ディレクトリが作成されませんでした: %s", dir)
					}

					// ファイルが作成されたことを確認
					if _, err := os.Stat(outputPath); os.IsNotExist(err) {
						t.Fatalf("出力ファイルが作成されませんでした: %s", outputPath)
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := outputService.WriteWithOptions(testData, tt.options)
				if err != nil {
					t.Fatalf("WriteWithOptions でエラーが発生しました: %v", err)
				}

				tt.verify(t, tt.options.OutputPath)
			})
		}
	})

	t.Run("形式選択エラーハンドリングテスト", func(t *testing.T) {
		tests := []struct {
			name        string
			options     OutputOptions
			expectError bool
			errorMsg    string
		}{
			{
				name: "無効な出力形式",
				options: OutputOptions{
					OutputPath: "test.txt",
					Format:     OutputFormat(999),
				},
				expectError: true,
				errorMsg:    "サポートされていない出力形式",
			},
			{
				name: "無効なファイルパス",
				options: OutputOptions{
					OutputPath: "invalid\x00path.csv",
					Format:     FormatCSV,
				},
				expectError: true,
				errorMsg:    "無効な文字が含まれています",
			},
			{
				name: "CSVデリミタ未指定",
				options: OutputOptions{
					OutputPath: "test.csv",
					Format:     FormatCSV,
					CSVOptions: &CSVWriteOptions{
						Delimiter: 0,
					},
				},
				expectError: true,
				errorMsg:    "CSVデリミタが指定されていません",
			},
			{
				name: "サポートされていないエンコーディング",
				options: OutputOptions{
					OutputPath: "test.csv",
					Format:     FormatCSV,
					CSVOptions: &CSVWriteOptions{
						Delimiter: ',',
						Encoding:  "SHIFT_JIS",
					},
				},
				expectError: true,
				errorMsg:    "サポートされていないエンコーディング",
			},
			{
				name: "無効なファイル権限",
				options: OutputOptions{
					OutputPath:      "test.csv",
					Format:          FormatCSV,
					FilePermissions: 0o1000,
				},
				expectError: true,
				errorMsg:    "無効なファイル権限",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := outputService.WriteWithOptions(testData, tt.options)

				if tt.expectError {
					if err == nil {
						t.Error("エラーが期待されましたが、エラーが発生しませんでした")
					} else if !strings.Contains(err.Error(), tt.errorMsg) {
						t.Errorf("期待されたエラーメッセージが含まれていません: 期待=%s, 実際=%s", tt.errorMsg, err.Error())
					}
				} else {
					if err != nil {
						t.Errorf("予期しないエラーが発生しました: %v", err)
					}
				}

				// テストファイルのクリーンアップ
				if tt.options.OutputPath != "" && !strings.Contains(tt.options.OutputPath, "\x00") {
					os.Remove(tt.options.OutputPath)
				}
			})
		}
	})
}

// TestFormatConsistencyAcrossOperations は操作間での形式一貫性テスト
func TestFormatConsistencyAcrossOperations(t *testing.T) {
	outputService := NewOutputService()

	testData := &analytics.ReportData{
		Headers: []string{"date", "pagePath", "sessions", "activeUsers"},
		Rows: [][]string{
			{"2023-01-01", "/home", "1250", "1100"},
			{"2023-01-01", "/about", "450", "420"},
		},
		Summary: analytics.ReportSummary{
			TotalRows: 2,
			DateRange: "2023-01-01 to 2023-01-01",
		},
	}

	t.Run("同一データの複数形式出力一貫性", func(t *testing.T) {
		// CSV出力
		csvFile := "consistency_test.csv"
		defer os.Remove(csvFile)

		err := outputService.WriteToFile(testData, csvFile, FormatCSV)
		if err != nil {
			t.Fatalf("CSV出力でエラーが発生しました: %v", err)
		}

		// JSON出力
		jsonFile := "consistency_test.json"
		defer os.Remove(jsonFile)

		err = outputService.WriteToFile(testData, jsonFile, FormatJSON)
		if err != nil {
			t.Fatalf("JSON出力でエラーが発生しました: %v", err)
		}

		// CSV内容の読み込み
		csvContent, err := os.ReadFile(csvFile)
		if err != nil {
			t.Fatalf("CSVファイル読み込みに失敗しました: %v", err)
		}

		csvReader := csv.NewReader(strings.NewReader(string(csvContent)))
		csvRecords, err := csvReader.ReadAll()
		if err != nil {
			t.Fatalf("CSV解析に失敗しました: %v", err)
		}

		// JSON内容の読み込み
		jsonContent, err := os.ReadFile(jsonFile)
		if err != nil {
			t.Fatalf("JSONファイル読み込みに失敗しました: %v", err)
		}

		var jsonRecords []JSONRecord
		if err := json.Unmarshal(jsonContent, &jsonRecords); err != nil {
			t.Fatalf("JSON解析に失敗しました: %v", err)
		}

		// データ一貫性の確認
		csvHeaders := csvRecords[0]
		csvDataRows := csvRecords[1:]

		if len(csvDataRows) != len(jsonRecords) {
			t.Errorf("データ行数が一致しません: CSV=%d, JSON=%d", len(csvDataRows), len(jsonRecords))
		}

		// 各行のデータ値を比較
		for i := 0; i < len(csvDataRows) && i < len(jsonRecords); i++ {
			csvRow := csvDataRows[i]
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
					t.Errorf("行%d, 列%s の値が一致しません: CSV=%s, JSON=%s", i+1, header, csvValue, jsonValue)
				}
			}
		}
	})

	t.Run("WriteOutput vs WriteWithOptions一貫性", func(t *testing.T) {
		// WriteOutputでの出力
		outputFile1 := "writeoutput_test.json"
		defer os.Remove(outputFile1)

		err := outputService.WriteOutput(testData, outputFile1, FormatJSON)
		if err != nil {
			t.Fatalf("WriteOutput でエラーが発生しました: %v", err)
		}

		// WriteWithOptionsでの出力
		outputFile2 := "writewithoptions_test.json"
		defer os.Remove(outputFile2)

		options := OutputOptions{
			OutputPath:        outputFile2,
			Format:            FormatJSON,
			OverwriteExisting: true,
			QuietMode:         true,
		}

		err = outputService.WriteWithOptions(testData, options)
		if err != nil {
			t.Fatalf("WriteWithOptions でエラーが発生しました: %v", err)
		}

		// 両ファイルの内容を比較
		content1, err := os.ReadFile(outputFile1)
		if err != nil {
			t.Fatalf("ファイル1の読み込みに失敗しました: %v", err)
		}

		content2, err := os.ReadFile(outputFile2)
		if err != nil {
			t.Fatalf("ファイル2の読み込みに失敗しました: %v", err)
		}

		var records1, records2 []JSONRecord
		if err := json.Unmarshal(content1, &records1); err != nil {
			t.Fatalf("ファイル1のJSON解析に失敗しました: %v", err)
		}

		if err := json.Unmarshal(content2, &records2); err != nil {
			t.Fatalf("ファイル2のJSON解析に失敗しました: %v", err)
		}

		// レコード数の比較
		if len(records1) != len(records2) {
			t.Errorf("レコード数が一致しません: WriteOutput=%d, WriteWithOptions=%d", len(records1), len(records2))
		}

		// 各レコードのデータ部分を比較（メタデータの時刻は異なる可能性があるため除外）
		for i := 0; i < len(records1) && i < len(records2); i++ {
			record1 := records1[i]
			record2 := records2[i]

			// Dimensionsの比較
			if len(record1.Dimensions) != len(record2.Dimensions) {
				t.Errorf("レコード%d のディメンション数が一致しません", i+1)
			}

			for key, value1 := range record1.Dimensions {
				if value2, exists := record2.Dimensions[key]; !exists {
					t.Errorf("レコード%d のディメンション %s が存在しません", i+1, key)
				} else if value1 != value2 {
					t.Errorf("レコード%d のディメンション %s の値が一致しません: %s vs %s", i+1, key, value1, value2)
				}
			}

			// Metricsの比較
			if len(record1.Metrics) != len(record2.Metrics) {
				t.Errorf("レコード%d のメトリクス数が一致しません", i+1)
			}

			for key, value1 := range record1.Metrics {
				if value2, exists := record2.Metrics[key]; !exists {
					t.Errorf("レコード%d のメトリクス %s が存在しません", i+1, key)
				} else if value1 != value2 {
					t.Errorf("レコード%d のメトリクス %s の値が一致しません: %s vs %s", i+1, key, value1, value2)
				}
			}
		}
	})
}

// TestOutputFormatPerformance は出力形式のパフォーマンステスト
func TestOutputFormatPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("パフォーマンステストをスキップします（-short フラグが指定されています）")
	}

	outputService := NewOutputService()

	// 大量データの作成
	largeTestData := &analytics.ReportData{
		Headers: []string{"property_id", "date", "pagePath", "sessions", "activeUsers", "newUsers"},
		Rows:    make([][]string, 1000), // 1000行のデータ
		Summary: analytics.ReportSummary{
			TotalRows: 1000,
			DateRange: "2023-01-01 to 2023-12-31",
		},
	}

	// テストデータの生成
	for i := 0; i < 1000; i++ {
		largeTestData.Rows[i] = []string{
			"987654321",
			fmt.Sprintf("2023-01-%02d", (i%31)+1),
			fmt.Sprintf("/page-%d", i),
			fmt.Sprintf("%d", 1000+i),
			fmt.Sprintf("%d", 900+i),
			fmt.Sprintf("%d", 800+i),
		}
	}

	t.Run("CSV vs JSON出力パフォーマンス比較", func(t *testing.T) {
		// CSV出力のベンチマーク
		csvFile := "performance_test.csv"
		defer os.Remove(csvFile)

		csvStart := time.Now()
		err := outputService.WriteToFile(largeTestData, csvFile, FormatCSV)
		csvDuration := time.Since(csvStart)

		if err != nil {
			t.Fatalf("CSV出力でエラーが発生しました: %v", err)
		}

		// JSON出力のベンチマーク
		jsonFile := "performance_test.json"
		defer os.Remove(jsonFile)

		jsonStart := time.Now()
		err = outputService.WriteToFile(largeTestData, jsonFile, FormatJSON)
		jsonDuration := time.Since(jsonStart)

		if err != nil {
			t.Fatalf("JSON出力でエラーが発生しました: %v", err)
		}

		t.Logf("CSV出力時間: %v", csvDuration)
		t.Logf("JSON出力時間: %v", jsonDuration)

		// ファイルサイズの比較
		csvInfo, _ := os.Stat(csvFile)
		jsonInfo, _ := os.Stat(jsonFile)

		t.Logf("CSVファイルサイズ: %d bytes", csvInfo.Size())
		t.Logf("JSONファイルサイズ: %d bytes", jsonInfo.Size())

		// 極端な性能差がないことを確認（10倍以上の差は問題）
		ratio := float64(jsonDuration) / float64(csvDuration)
		if ratio > 10.0 || ratio < 0.1 {
			t.Logf("警告: CSV と JSON の出力時間に大きな差があります（比率: %.2f）", ratio)
		}
	})
}

// TestOutputFormatEdgeCases は出力形式のエッジケーステスト
func TestOutputFormatEdgeCases(t *testing.T) {
	outputService := NewOutputService()

	t.Run("空データの形式選択", func(t *testing.T) {
		emptyData := &analytics.ReportData{
			Headers: []string{"date", "sessions"},
			Rows:    [][]string{},
			Summary: analytics.ReportSummary{
				TotalRows: 0,
				DateRange: "2023-01-01 to 2023-01-01",
			},
		}

		// CSV出力
		csvFile := "empty_csv.csv"
		defer os.Remove(csvFile)

		err := outputService.WriteToFile(emptyData, csvFile, FormatCSV)
		if err != nil {
			t.Fatalf("空データのCSV出力でエラーが発生しました: %v", err)
		}

		// JSON出力
		jsonFile := "empty_json.json"
		defer os.Remove(jsonFile)

		err = outputService.WriteToFile(emptyData, jsonFile, FormatJSON)
		if err != nil {
			t.Fatalf("空データのJSON出力でエラーが発生しました: %v", err)
		}

		// JSON内容の確認（空配列であることを確認）
		jsonContent, err := os.ReadFile(jsonFile)
		if err != nil {
			t.Fatalf("JSONファイル読み込みに失敗しました: %v", err)
		}

		var records []JSONRecord
		if err := json.Unmarshal(jsonContent, &records); err != nil {
			t.Fatalf("JSON解析に失敗しました: %v", err)
		}

		if len(records) != 0 {
			t.Errorf("空データのJSONレコード数が期待値と一致しません: 期待=0, 実際=%d", len(records))
		}
	})

	t.Run("単一行データの形式選択", func(t *testing.T) {
		singleRowData := &analytics.ReportData{
			Headers: []string{"date", "sessions"},
			Rows:    [][]string{{"2023-01-01", "1250"}},
			Summary: analytics.ReportSummary{
				TotalRows: 1,
				DateRange: "2023-01-01 to 2023-01-01",
			},
		}

		// 両形式での出力テスト
		formats := []struct {
			format OutputFormat
			ext    string
		}{
			{FormatCSV, "csv"},
			{FormatJSON, "json"},
		}

		for _, f := range formats {
			fileName := fmt.Sprintf("single_row.%s", f.ext)
			defer os.Remove(fileName)

			err := outputService.WriteToFile(singleRowData, fileName, f.format)
			if err != nil {
				t.Fatalf("単一行データの%s出力でエラーが発生しました: %v", f.format, err)
			}

			// ファイルが作成されたことを確認
			if _, err := os.Stat(fileName); os.IsNotExist(err) {
				t.Errorf("単一行データの%sファイルが作成されませんでした", f.format)
			}
		}
	})

	t.Run("大量列データの形式選択", func(t *testing.T) {
		// 多数の列を持つデータ
		manyColumns := make([]string, 50)
		for i := 0; i < 50; i++ {
			manyColumns[i] = fmt.Sprintf("column_%d", i+1)
		}

		manyColumnData := &analytics.ReportData{
			Headers: manyColumns,
			Rows: [][]string{
				make([]string, 50),
			},
			Summary: analytics.ReportSummary{
				TotalRows: 1,
				DateRange: "2023-01-01 to 2023-01-01",
			},
		}

		// 行データを生成
		for i := 0; i < 50; i++ {
			manyColumnData.Rows[0][i] = fmt.Sprintf("value_%d", i+1)
		}

		// 両形式での出力テスト
		csvFile := "many_columns.csv"
		jsonFile := "many_columns.json"
		defer os.Remove(csvFile)
		defer os.Remove(jsonFile)

		err := outputService.WriteToFile(manyColumnData, csvFile, FormatCSV)
		if err != nil {
			t.Fatalf("多列データのCSV出力でエラーが発生しました: %v", err)
		}

		err = outputService.WriteToFile(manyColumnData, jsonFile, FormatJSON)
		if err != nil {
			t.Fatalf("多列データのJSON出力でエラーが発生しました: %v", err)
		}

		// JSON構造の確認
		jsonContent, err := os.ReadFile(jsonFile)
		if err != nil {
			t.Fatalf("JSONファイル読み込みに失敗しました: %v", err)
		}

		var records []JSONRecord
		if err := json.Unmarshal(jsonContent, &records); err != nil {
			t.Fatalf("JSON解析に失敗しました: %v", err)
		}

		if len(records) != 1 {
			t.Errorf("JSONレコード数が期待値と一致しません: 期待=1, 実際=%d", len(records))
		}

		// ディメンションとメトリクスの合計が50列であることを確認
		if len(records) > 0 {
			totalFields := len(records[0].Dimensions) + len(records[0].Metrics)
			if totalFields != 50 {
				t.Errorf("総フィールド数が期待値と一致しません: 期待=50, 実際=%d", totalFields)
			}
		}
	})
}
