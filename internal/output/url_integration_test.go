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
	"github.com/ymotongpoo/ga/internal/url"
)

func TestURLIntegration_CSV(t *testing.T) {
	tests := []struct {
		name        string
		data        *analytics.ReportData
		expectedCSV string
	}{
		{
			name: "base_url付きURL結合",
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
			expectedCSV: `property_id,stream_id,date,fullURL,sessions
123456,stream1,2023-01-01,https://example.com/home,1000
123456,stream1,2023-01-01,https://example.com/about,500
`,
		},
		{
			name: "base_urlなしの場合",
			data: &analytics.ReportData{
				Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
				Rows: [][]string{
					{"123456", "stream2", "2023-01-01", "/home", "1000"},
					{"123456", "stream2", "2023-01-01", "/about", "500"},
				},
				StreamURLs: map[string]string{}, // 空のマッピング
				Summary: analytics.ReportSummary{
					TotalRows: 2,
					DateRange: "2023-01-01 to 2023-01-01",
				},
			},
			expectedCSV: `property_id,stream_id,date,fullURL,sessions
123456,stream2,2023-01-01,/home,1000
123456,stream2,2023-01-01,/about,500
`,
		},
		{
			name: "絶対URLの場合",
			data: &analytics.ReportData{
				Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
				Rows: [][]string{
					{"123456", "stream1", "2023-01-01", "https://other.com/page", "1000"},
					{"123456", "stream1", "2023-01-01", "/home", "500"},
				},
				StreamURLs: map[string]string{
					"stream1": "https://example.com",
				},
				Summary: analytics.ReportSummary{
					TotalRows: 2,
					DateRange: "2023-01-01 to 2023-01-01",
				},
			},
			expectedCSV: `property_id,stream_id,date,fullURL,sessions
123456,stream1,2023-01-01,https://other.com/page,1000
123456,stream1,2023-01-01,https://example.com/home,500
`,
		},
		{
			name: "pagePathがない場合",
			data: &analytics.ReportData{
				Headers: []string{"property_id", "stream_id", "date", "sessions"},
				Rows: [][]string{
					{"123456", "stream1", "2023-01-01", "1000"},
				},
				StreamURLs: map[string]string{
					"stream1": "https://example.com",
				},
				Summary: analytics.ReportSummary{
					TotalRows: 1,
					DateRange: "2023-01-01 to 2023-01-01",
				},
			},
			expectedCSV: `property_id,stream_id,date,sessions
123456,stream1,2023-01-01,1000
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewOutputService()
			var buf bytes.Buffer

			err := service.WriteCSV(tt.data, &buf)
			if err != nil {
				t.Fatalf("WriteCSV() error = %v", err)
			}

			result := buf.String()
			if result != tt.expectedCSV {
				t.Errorf("CSV output mismatch:\nExpected:\n%s\nGot:\n%s", tt.expectedCSV, result)
			}
		})
	}
}

func TestURLIntegration_JSON(t *testing.T) {
	tests := []struct {
		name         string
		data         *analytics.ReportData
		expectedURLs []string
	}{
		{
			name: "base_url付きURL結合",
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
			expectedURLs: []string{
				"https://example.com/home",
				"https://example.com/about",
			},
		},
		{
			name: "base_urlなしの場合",
			data: &analytics.ReportData{
				Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
				Rows: [][]string{
					{"123456", "stream2", "2023-01-01", "/home", "1000"},
				},
				StreamURLs: map[string]string{}, // 空のマッピング
				Summary: analytics.ReportSummary{
					TotalRows: 1,
					DateRange: "2023-01-01 to 2023-01-01",
				},
			},
			expectedURLs: []string{
				"/home",
			},
		},
		{
			name: "絶対URLの場合",
			data: &analytics.ReportData{
				Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
				Rows: [][]string{
					{"123456", "stream1", "2023-01-01", "https://other.com/page", "1000"},
				},
				StreamURLs: map[string]string{
					"stream1": "https://example.com",
				},
				Summary: analytics.ReportSummary{
					TotalRows: 1,
					DateRange: "2023-01-01 to 2023-01-01",
				},
			},
			expectedURLs: []string{
				"https://other.com/page",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewOutputService()
			var buf bytes.Buffer

			err := service.WriteJSON(tt.data, &buf)
			if err != nil {
				t.Fatalf("WriteJSON() error = %v", err)
			}

			// JSONを解析
			var records []JSONRecord
			if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
				t.Fatalf("JSON unmarshal error = %v", err)
			}

			if len(records) != len(tt.expectedURLs) {
				t.Fatalf("Expected %d records, got %d", len(tt.expectedURLs), len(records))
			}

			for i, record := range records {
				fullURL, exists := record.Dimensions["fullURL"]
				if !exists {
					t.Errorf("Record %d: fullURL dimension not found", i)
					continue
				}

				if fullURL != tt.expectedURLs[i] {
					t.Errorf("Record %d: Expected fullURL %s, got %s", i, tt.expectedURLs[i], fullURL)
				}
			}
		})
	}
}

func TestProcessHeaders(t *testing.T) {
	service := &OutputServiceImpl{}

	tests := []struct {
		name               string
		headers            []string
		expectedHeaders    []string
		expectedPageIndex  int
	}{
		{
			name:              "pagePathを含む",
			headers:           []string{"date", "pagePath", "sessions"},
			expectedHeaders:   []string{"date", "fullURL", "sessions"},
			expectedPageIndex: 1,
		},
		{
			name:              "pagePathを含まない",
			headers:           []string{"date", "sessions"},
			expectedHeaders:   []string{"date", "sessions"},
			expectedPageIndex: -1,
		},
		{
			name:              "大文字小文字混合",
			headers:           []string{"date", "PagePath", "sessions"},
			expectedHeaders:   []string{"date", "fullURL", "sessions"}, // 大文字小文字を区別しない
			expectedPageIndex: 1,
		},
		{
			name:              "複数のpagePath",
			headers:           []string{"pagePath", "date", "pagePath", "sessions"},
			expectedHeaders:   []string{"fullURL", "date", "fullURL", "sessions"},
			expectedPageIndex: 2, // 最後のインデックス
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processedHeaders, pagePathIndex := service.processHeaders(tt.headers)

			if len(processedHeaders) != len(tt.expectedHeaders) {
				t.Errorf("Expected %d headers, got %d", len(tt.expectedHeaders), len(processedHeaders))
				return
			}

			for i, expected := range tt.expectedHeaders {
				if processedHeaders[i] != expected {
					t.Errorf("Header %d: Expected %s, got %s", i, expected, processedHeaders[i])
				}
			}

			if pagePathIndex != tt.expectedPageIndex {
				t.Errorf("Expected pagePathIndex %d, got %d", tt.expectedPageIndex, pagePathIndex)
			}
		})
	}
}

func TestProcessRow(t *testing.T) {
	service := &OutputServiceImpl{}

	tests := []struct {
		name         string
		row          []string
		headers      []string
		streamURLs   map[string]string
		expectedRow  []string
	}{
		{
			name:       "通常のURL結合",
			row:        []string{"stream1", "2023-01-01", "/home", "1000"},
			headers:    []string{"stream_id", "date", "pagePath", "sessions"},
			streamURLs: map[string]string{"stream1": "https://example.com"},
			expectedRow: []string{"stream1", "2023-01-01", "https://example.com/home", "1000"},
		},
		{
			name:       "ストリームIDが見つからない",
			row:        []string{"2023-01-01", "/home", "1000"},
			headers:    []string{"date", "pagePath", "sessions"},
			streamURLs: map[string]string{"stream1": "https://example.com"},
			expectedRow: []string{"2023-01-01", "/home", "1000"},
		},
		{
			name:       "pagePathが見つからない",
			row:        []string{"stream1", "2023-01-01", "1000"},
			headers:    []string{"stream_id", "date", "sessions"},
			streamURLs: map[string]string{"stream1": "https://example.com"},
			expectedRow: []string{"stream1", "2023-01-01", "1000"},
		},
		{
			name:       "絶対URLの場合",
			row:        []string{"stream1", "2023-01-01", "https://other.com/page", "1000"},
			headers:    []string{"stream_id", "date", "pagePath", "sessions"},
			streamURLs: map[string]string{"stream1": "https://example.com"},
			expectedRow: []string{"stream1", "2023-01-01", "https://other.com/page", "1000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// processRowを呼び出すためのヘルパー
			pagePathIndex := -1
			for i, header := range tt.headers {
				if strings.ToLower(header) == "pagepath" {
					pagePathIndex = i
					break
				}
			}

			// URLProcessorを作成
			urlProcessor := url.NewURLProcessor(tt.streamURLs)
			processedRow := service.processRow(tt.row, pagePathIndex, urlProcessor, tt.headers)

			if len(processedRow) != len(tt.expectedRow) {
				t.Errorf("Expected %d columns, got %d", len(tt.expectedRow), len(processedRow))
				return
			}

			for i, expected := range tt.expectedRow {
				if processedRow[i] != expected {
					t.Errorf("Column %d: Expected %s, got %s", i, expected, processedRow[i])
				}
			}
		})
	}
}

func TestExtractStreamIDFromRow(t *testing.T) {
	service := &OutputServiceImpl{}

	tests := []struct {
		name       string
		row        []string
		headers    []string
		expectedID string
	}{
		{
			name:       "stream_idが存在する",
			row:        []string{"stream1", "2023-01-01", "/home"},
			headers:    []string{"stream_id", "date", "pagePath"},
			expectedID: "stream1",
		},
		{
			name:       "STREAM_ID（大文字）が存在する",
			row:        []string{"stream1", "2023-01-01", "/home"},
			headers:    []string{"STREAM_ID", "date", "pagePath"},
			expectedID: "stream1",
		},
		{
			name:       "stream_idが存在しない",
			row:        []string{"2023-01-01", "/home"},
			headers:    []string{"date", "pagePath"},
			expectedID: "",
		},
		{
			name:       "空の行",
			row:        []string{},
			headers:    []string{"stream_id", "date", "pagePath"},
			expectedID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.extractStreamIDFromRow(tt.row, tt.headers)
			if result != tt.expectedID {
				t.Errorf("Expected %s, got %s", tt.expectedID, result)
			}
		})
	}
}