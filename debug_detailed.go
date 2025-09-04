package main

import (
	"bytes"
	"fmt"
	"github.com/ymotongpoo/ga/internal/analytics"
	"github.com/ymotongpoo/ga/internal/output"
	"github.com/ymotongpoo/ga/internal/url"
)

func main() {
	fmt.Println("=== URL結合機能の詳細デバッグ ===")

	// StreamURLsマッピングをテスト
	streamURLs := map[string]string{
		"3282358539": "https://ymotongpoo.hatenablog.com/",
		"3803158344": "https://zenn.com/ymotongpoo/",
	}

	fmt.Println("1. StreamURLsマッピング:")
	for streamID, baseURL := range streamURLs {
		fmt.Printf("   %s -> %s\n", streamID, baseURL)
	}

	// URLProcessorを直接テスト
	urlProcessor := url.NewURLProcessor(streamURLs)
	
	fmt.Println("\n2. URLProcessor直接テスト:")
	testCases := []struct {
		streamID string
		pagePath string
	}{
		{"3282358539", "/entry/2024/01/01/120000"},
		{"3803158344", "/articles/golang-tips"},
	}

	for _, tc := range testCases {
		result := urlProcessor.ProcessPagePath(tc.streamID, tc.pagePath)
		fmt.Printf("   Stream %s + %s = %s\n", tc.streamID, tc.pagePath, result)
	}

	// 実際のデータ処理をテスト
	fmt.Println("\n3. 実際のデータ処理テスト:")
	data := &analytics.ReportData{
		Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
		Rows: [][]string{
			{"320031301", "3282358539", "2024-01-01", "/entry/2024/01/01/120000", "100"},
			{"321189208", "3803158344", "2024-01-01", "/articles/golang-tips", "200"},
		},
		StreamURLs: streamURLs,
		Summary: analytics.ReportSummary{
			TotalRows: 2,
			DateRange: "2024-01-01 to 2024-01-01",
		},
	}

	// 各行のストリームIDを確認
	fmt.Println("   データ行の詳細:")
	for i, row := range data.Rows {
		fmt.Printf("   行%d: %v\n", i+1, row)
		if len(row) >= 2 {
			streamID := row[1] // stream_idは2番目の列
			baseURL := streamURLs[streamID]
			fmt.Printf("        ストリームID: %s -> ベースURL: %s\n", streamID, baseURL)
		}
	}

	// CSV出力をテスト
	service := output.NewOutputService()
	var buf bytes.Buffer
	err := service.WriteCSV(data, &buf)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("\n4. CSV出力結果:")
	fmt.Print(buf.String())
}