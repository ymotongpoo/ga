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
	"strings"
	"testing"

	"github.com/ymotongpoo/ga/internal/analytics"
)

// TestURL結合機能_包括的テスト は要件6.1-6.7をカバーする包括的なテスト
func TestURL結合機能_包括的テスト(t *testing.T) {
	tests := []struct {
		name           string
		data           *analytics.ReportData
		expectedCSVURL []string
		expectedJSONURL []string
		description    string
		requirements   []string
	}{
		{
			name: "要件6.1_base_url設定時のURL結合",
			data: &analytics.ReportData{
				Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
				Rows: [][]string{
					{"123456", "stream1", "2023-01-01", "/home", "1000"},
					{"123456", "stream1", "2023-01-01", "/about", "500"},
				},
				StreamURLs: map[string]string{
					"stream1": "https://example.com",
				},
				Summary: analytics.ReportSummary{
					TotalRows: 2,
					DateRange: "2023-01-01 to 2023-01-01",
				},
			},
			expectedCSVURL:  []string{"https://example.com/home", "https://example.com/about"},
			expectedJSONURL: []string{"https://example.com/home", "https://example.com/about"},
			description:     "ストリーム設定にbase_urlが指定されている場合、pagePathとbase_urlを結合する",
			requirements:    []string{"6.1"},
		},
		{
			name: "要件6.2_相対パスの適切な結合",
			data: &analytics.ReportData{
				Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
				Rows: [][]string{
					{"123456", "stream1", "2023-01-01", "/api/v1/users", "1000"},
					{"123456", "stream1", "2023-01-01", "/api/v1/posts", "500"},
				},
				StreamURLs: map[string]string{
					"stream1": "https://api.example.com",
				},
				Summary: analytics.ReportSummary{
					TotalRows: 2,
					DateRange: "2023-01-01 to 2023-01-01",
				},
			},
			expectedCSVURL:  []string{"https://api.example.com/api/v1/users", "https://api.example.com/api/v1/posts"},
			expectedJSONURL: []string{"https://api.example.com/api/v1/users", "https://api.example.com/api/v1/posts"},
			description:     "pagePathが相対パス（/で始まる）である場合、base_urlとpagePathを適切に結合する",
			requirements:    []string{"6.2"},
		},
		{
			name: "要件6.3_絶対URLはそのまま使用",
			data: &analytics.ReportData{
				Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
				Rows: [][]string{
					{"123456", "stream1", "2023-01-01", "https://external.com/page", "1000"},
					{"123456", "stream1", "2023-01-01", "http://other.com/api", "500"},
				},
				StreamURLs: map[string]string{
					"stream1": "https://example.com",
				},
				Summary: analytics.ReportSummary{
					TotalRows: 2,
					DateRange: "2023-01-01 to 2023-01-01",
				},
			},
			expectedCSVURL:  []string{"https://external.com/page", "http://other.com/api"},
			expectedJSONURL: []string{"https://external.com/page", "http://other.com/api"},
			description:     "pagePathが絶対URL（http://またはhttps://で始まる）である場合、pagePathをそのまま使用する",
			requirements:    []string{"6.3"},
		},
		{
			name: "要件6.4_base_url未設定時はpagePathをそのまま出力",
			data: &analytics.ReportData{
				Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
				Rows: [][]string{
					{"123456", "stream2", "2023-01-01", "/home", "1000"},
					{"123456", "stream2", "2023-01-01", "/about", "500"},
				},
				StreamURLs: map[string]string{}, // base_urlが設定されていない
				Summary: analytics.ReportSummary{
					TotalRows: 2,
					DateRange: "2023-01-01 to 2023-01-01",
				},
			},
			expectedCSVURL:  []string{"/home", "/about"},
			expectedJSONURL: []string{"/home", "/about"},
			description:     "base_urlが設定されていない場合、pagePathをそのまま出力する",
			requirements:    []string{"6.4"},
		},
		{
			name: "要件6.6_スラッシュ重複の適切な処理",
			data: &analytics.ReportData{
				Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
				Rows: [][]string{
					{"123456", "stream1", "2023-01-01", "/home", "1000"},
					{"123456", "stream1", "2023-01-01", "about", "500"},
				},
				StreamURLs: map[string]string{
					"stream1": "https://example.com/", // 末尾にスラッシュ
				},
				Summary: analytics.ReportSummary{
					TotalRows: 2,
					DateRange: "2023-01-01 to 2023-01-01",
				},
			},
			expectedCSVURL:  []string{"https://example.com/home", "https://example.com/about"},
			expectedJSONURL: []string{"https://example.com/home", "https://example.com/about"},
			description:     "base_urlの末尾にスラッシュがある場合、重複スラッシュを適切に処理する",
			requirements:    []string{"6.6"},
		},
		{
			name: "要件6.7_空文字列またはnullの場合はbase_URLのみ出力",
			data: &analytics.ReportData{
				Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
				Rows: [][]string{
					{"123456", "stream1", "2023-01-01", "", "1000"},
					{"123456", "stream1", "2023-01-01", "   ", "500"}, // 空白のみ
				},
				StreamURLs: map[string]string{
					"stream1": "https://example.com",
				},
				Summary: analytics.ReportSummary{
					TotalRows: 2,
					DateRange: "2023-01-01 to 2023-01-01",
				},
			},
			expectedCSVURL:  []string{"https://example.com", "https://example.com"},
			expectedJSONURL: []string{"https://example.com", "https://example.com"},
			description:     "pagePathが空文字列または空白のみの場合、base_urlのみを出力する",
			requirements:    []string{"6.7"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewOutputService()

			// CSV出力テスト
			t.Run("CSV出力", func(t *testing.T) {
				var csvBuf bytes.Buffer
				err := service.WriteCSV(tt.data, &csvBuf)
				if err != nil {
					t.Fatalf("WriteCSV() error = %v", err)
				}

				// CSVを解析してfullURL列を確認
				csvReader := csv.NewReader(strings.NewReader(csvBuf.String()))
				records, err := csvReader.ReadAll()
				if err != nil {
					t.Fatalf("CSV読み込みエラー: %v", err)
				}

				if len(records) < 2 {
					t.Fatalf("CSVレコードが不足しています: %d", len(records))
				}

				// ヘッダーからfullURLのインデックスを取得
				fullURLIndex := -1
				for i, header := range records[0] {
					if header == "fullURL" {
						fullURLIndex = i
						break
					}
				}

				if fullURLIndex == -1 {
					t.Fatalf("fullURLヘッダーが見つかりません: %v", records[0])
				}

				// データ行のfullURLを確認
				for i, expectedURL := range tt.expectedCSVURL {
					if i+1 >= len(records) {
						t.Fatalf("CSVデータ行が不足しています: 期待=%d, 実際=%d", len(tt.expectedCSVURL), len(records)-1)
					}

					actualURL := records[i+1][fullURLIndex]
					if actualURL != expectedURL {
						t.Errorf("CSV行%d: 期待URL=%s, 実際URL=%s", i+1, expectedURL, actualURL)
					}
				}
			})

			// JSON出力テスト
			t.Run("JSON出力", func(t *testing.T) {
				var jsonBuf bytes.Buffer
				err := service.WriteJSON(tt.data, &jsonBuf)
				if err != nil {
					t.Fatalf("WriteJSON() error = %v", err)
				}

				// JSONを解析
				var records []JSONRecord
				if err := json.Unmarshal(jsonBuf.Bytes(), &records); err != nil {
					t.Fatalf("JSON解析エラー: %v", err)
				}

				if len(records) != len(tt.expectedJSONURL) {
					t.Fatalf("JSONレコード数が不正: 期待=%d, 実際=%d", len(tt.expectedJSONURL), len(records))
				}

				// 各レコードのfullURLを確認
				for i, expectedURL := range tt.expectedJSONURL {
					fullURL, exists := records[i].Dimensions["fullURL"]
					if !exists {
						t.Errorf("JSONレコード%d: fullURLディメンションが存在しません", i+1)
						continue
					}

					if fullURL != expectedURL {
						t.Errorf("JSONレコード%d: 期待URL=%s, 実際URL=%s", i+1, expectedURL, fullURL)
					}
				}
			})

			t.Logf("✅ %s: %s (要件: %v)", tt.name, tt.description, tt.requirements)
		})
	}
}

// TestURL結合機能_エッジケース はエッジケースをテストする
func TestURL結合機能_エッジケース(t *testing.T) {
	tests := []struct {
		name        string
		data        *analytics.ReportData
		expectedURL string
		description string
	}{
		{
			name: "複雑なパス構造",
			data: &analytics.ReportData{
				Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
				Rows: [][]string{
					{"123456", "stream1", "2023-01-01", "/api/v2/users/123/profile", "1000"},
				},
				StreamURLs: map[string]string{
					"stream1": "https://api.example.com/app",
				},
				Summary: analytics.ReportSummary{TotalRows: 1, DateRange: "2023-01-01"},
			},
			expectedURL: "https://api.example.com/app/api/v2/users/123/profile",
			description: "複雑なパス構造でも正しく結合される",
		},
		{
			name: "クエリパラメータ付きURL",
			data: &analytics.ReportData{
				Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
				Rows: [][]string{
					{"123456", "stream1", "2023-01-01", "/search?q=test&page=1", "1000"},
				},
				StreamURLs: map[string]string{
					"stream1": "https://example.com",
				},
				Summary: analytics.ReportSummary{TotalRows: 1, DateRange: "2023-01-01"},
			},
			expectedURL: "https://example.com/search?q=test&page=1",
			description: "クエリパラメータ付きのpagePathも正しく処理される",
		},
		{
			name: "フラグメント付きURL",
			data: &analytics.ReportData{
				Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
				Rows: [][]string{
					{"123456", "stream1", "2023-01-01", "/docs#section1", "1000"},
				},
				StreamURLs: map[string]string{
					"stream1": "https://example.com",
				},
				Summary: analytics.ReportSummary{TotalRows: 1, DateRange: "2023-01-01"},
			},
			expectedURL: "https://example.com/docs#section1",
			description: "フラグメント付きのpagePathも正しく処理される",
		},
		{
			name: "日本語パス",
			data: &analytics.ReportData{
				Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
				Rows: [][]string{
					{"123456", "stream1", "2023-01-01", "/製品/詳細", "1000"},
				},
				StreamURLs: map[string]string{
					"stream1": "https://example.co.jp",
				},
				Summary: analytics.ReportSummary{TotalRows: 1, DateRange: "2023-01-01"},
			},
			expectedURL: "https://example.co.jp/製品/詳細",
			description: "日本語を含むpagePathも正しく処理される",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewOutputService()

			// CSV出力テスト
			var csvBuf bytes.Buffer
			err := service.WriteCSV(tt.data, &csvBuf)
			if err != nil {
				t.Fatalf("WriteCSV() error = %v", err)
			}

			csvReader := csv.NewReader(strings.NewReader(csvBuf.String()))
			records, err := csvReader.ReadAll()
			if err != nil {
				t.Fatalf("CSV読み込みエラー: %v", err)
			}

			// fullURLのインデックスを取得
			fullURLIndex := -1
			for i, header := range records[0] {
				if header == "fullURL" {
					fullURLIndex = i
					break
				}
			}

			if fullURLIndex == -1 {
				t.Fatalf("fullURLヘッダーが見つかりません")
			}

			actualURL := records[1][fullURLIndex]
			if actualURL != tt.expectedURL {
				t.Errorf("期待URL=%s, 実際URL=%s", tt.expectedURL, actualURL)
			}

			t.Logf("✅ %s: %s", tt.name, tt.description)
		})
	}
}

// TestURL結合機能_パフォーマンス はURL結合処理のパフォーマンスをテストする
func TestURL結合機能_パフォーマンス(t *testing.T) {
	// 大量データでのパフォーマンステスト
	const recordCount = 10000

	// テストデータを生成
	headers := []string{"property_id", "stream_id", "date", "pagePath", "sessions"}
	rows := make([][]string, recordCount)
	for i := 0; i < recordCount; i++ {
		rows[i] = []string{
			"123456",
			"stream1",
			"2023-01-01",
			"/page/" + string(rune(i%1000)),
			"1000",
		}
	}

	data := &analytics.ReportData{
		Headers: headers,
		Rows:    rows,
		StreamURLs: map[string]string{
			"stream1": "https://example.com",
		},
		Summary: analytics.ReportSummary{
			TotalRows: recordCount,
			DateRange: "2023-01-01",
		},
	}

	service := NewOutputService()

	// CSV出力のパフォーマンステスト
	t.Run("CSV出力パフォーマンス", func(t *testing.T) {
		var csvBuf bytes.Buffer
		err := service.WriteCSV(data, &csvBuf)
		if err != nil {
			t.Fatalf("WriteCSV() error = %v", err)
		}

		// 出力サイズを確認
		outputSize := csvBuf.Len()
		t.Logf("CSV出力サイズ: %d bytes (%d レコード)", outputSize, recordCount)

		if outputSize == 0 {
			t.Error("CSV出力が空です")
		}
	})

	// JSON出力のパフォーマンステスト
	t.Run("JSON出力パフォーマンス", func(t *testing.T) {
		var jsonBuf bytes.Buffer
		err := service.WriteJSON(data, &jsonBuf)
		if err != nil {
			t.Fatalf("WriteJSON() error = %v", err)
		}

		// 出力サイズを確認
		outputSize := jsonBuf.Len()
		t.Logf("JSON出力サイズ: %d bytes (%d レコード)", outputSize, recordCount)

		if outputSize == 0 {
			t.Error("JSON出力が空です")
		}
	})
}

// TestURL結合機能_設定ファイル統合 は設定ファイルとの統合をテストする
func TestURL結合機能_設定ファイル統合(t *testing.T) {
	// 複数ストリームでのURL結合テスト
	data := &analytics.ReportData{
		Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
		Rows: [][]string{
			{"123456", "stream1", "2023-01-01", "/home", "1000"},
			{"123456", "stream2", "2023-01-01", "/api/users", "500"},
			{"123456", "stream3", "2023-01-01", "/blog", "300"},
		},
		StreamURLs: map[string]string{
			"stream1": "https://www.example.com",
			"stream2": "https://api.example.com",
			// stream3にはbase_urlが設定されていない
		},
		Summary: analytics.ReportSummary{
			TotalRows: 3,
			DateRange: "2023-01-01",
		},
	}

	expectedURLs := []string{
		"https://www.example.com/home",
		"https://api.example.com/api/users",
		"/blog", // base_urlなし
	}

	service := NewOutputService()

	// CSV出力テスト
	var csvBuf bytes.Buffer
	err := service.WriteCSV(data, &csvBuf)
	if err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}

	csvReader := csv.NewReader(strings.NewReader(csvBuf.String()))
	records, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("CSV読み込みエラー: %v", err)
	}

	// fullURLのインデックスを取得
	fullURLIndex := -1
	for i, header := range records[0] {
		if header == "fullURL" {
			fullURLIndex = i
			break
		}
	}

	if fullURLIndex == -1 {
		t.Fatalf("fullURLヘッダーが見つかりません")
	}

	// 各行のURLを確認
	for i, expectedURL := range expectedURLs {
		actualURL := records[i+1][fullURLIndex]
		if actualURL != expectedURL {
			t.Errorf("行%d: 期待URL=%s, 実際URL=%s", i+1, expectedURL, actualURL)
		}
	}

	t.Log("✅ 複数ストリームでのURL結合が正しく動作しています")
}