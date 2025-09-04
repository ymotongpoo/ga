package main

import (
	"fmt"
	"os"

	"github.com/ymotongpoo/ga/internal/analytics"
	"github.com/ymotongpoo/ga/internal/output"
)

func main() {
	// 複数ストリームのテストデータを作成
	testData := &analytics.ReportData{
		Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions", "activeUsers"},
		Rows: [][]string{
			// ストリーム1のデータ
			{"320031301", "3282358539", "2024-01-01", "/entry/2024/01/01/120000", "10", "8"},
			{"320031301", "3282358539", "2024-01-01", "/entry/2024/01/02/120000", "15", "12"},
			// ストリーム2のデータ
			{"321189208", "3803158344", "2024-01-01", "/articles/go-tutorial", "20", "18"},
			{"321189208", "3803158344", "2024-01-01", "/articles/python-basics", "25", "22"},
		},
		StreamURLs: map[string]string{
			"3282358539": "https://ymotongpoo.hatenablog.com",
			"3803158344": "https://zenn.com/ymotongpoo",
		},
		Summary: analytics.ReportSummary{
			TotalRows:  4,
			DateRange:  "2024-01-01 - 2024-01-01",
			Properties: []string{"320031301", "321189208"},
		},
	}

	fmt.Println("=== 複数ストリーム統合テスト ===")
	fmt.Printf("StreamURLs: %v\n", testData.StreamURLs)
	fmt.Printf("Headers: %v\n", testData.Headers)
	fmt.Println("\n元データ:")
	for i, row := range testData.Rows {
		fmt.Printf("行 %d: %v\n", i+1, row)
	}

	// 出力サービスを作成
	outputService := output.NewOutputService()

	// CSV出力をテスト
	fmt.Println("\n=== CSV出力テスト ===")
	err := outputService.WriteCSV(testData, os.Stdout)
	if err != nil {
		fmt.Printf("CSV出力エラー: %v\n", err)
	}
}