package main

import (
	"bytes"
	"fmt"
	"github.com/ymotongpoo/ga/internal/analytics"
	"github.com/ymotongpoo/ga/internal/output"
)

func main() {
	// 複数ストリームのテストデータを作成
	data := &analytics.ReportData{
		Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
		Rows: [][]string{
			// ストリーム1のデータ
			{"320031301", "3282358539", "2024-01-01", "/entry/2024/01/01/120000", "100"},
			{"320031301", "3282358539", "2024-01-01", "/entry/2024/01/02/120000", "150"},
			// ストリーム2のデータ
			{"321189208", "3803158344", "2024-01-01", "/articles/golang-tips", "200"},
			{"321189208", "3803158344", "2024-01-01", "/articles/web-development", "250"},
		},
		StreamURLs: map[string]string{
			"3282358539": "https://ymotongpoo.hatenablog.com/",
			"3803158344": "https://zenn.com/ymotongpoo/",
		},
		Summary: analytics.ReportSummary{
			TotalRows: 4,
			DateRange: "2024-01-01 to 2024-01-01",
		},
	}

	// 出力サービスを作成
	service := output.NewOutputService()

	// CSV出力をテスト
	var buf bytes.Buffer
	err := service.WriteCSV(data, &buf)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("=== CSV出力結果 ===")
	fmt.Print(buf.String())
	fmt.Println("\n=== 期待される結果 ===")
	fmt.Println("property_id,stream_id,date,fullURL,sessions")
	fmt.Println("320031301,3282358539,2024-01-01,https://ymotongpoo.hatenablog.com/entry/2024/01/01/120000,100")
	fmt.Println("320031301,3282358539,2024-01-01,https://ymotongpoo.hatenablog.com/entry/2024/01/02/120000,150")
	fmt.Println("321189208,3803158344,2024-01-01,https://zenn.com/ymotongpoo/articles/golang-tips,200")
	fmt.Println("321189208,3803158344,2024-01-01,https://zenn.com/ymotongpoo/articles/web-development,250")
}