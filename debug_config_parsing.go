package main

import (
	"fmt"
	"log"

	"github.com/ymotongpoo/ga/internal/config"
)

func main() {
	fmt.Println("=== 設定ファイル解析テスト ===")

	// 設定サービスを作成
	configService := config.NewConfigService()

	// 設定ファイルを読み込み
	cfg, err := configService.LoadConfig("ga.yoshi.yaml")
	if err != nil {
		log.Fatalf("設定ファイルの読み込みに失敗: %v", err)
	}

	fmt.Printf("開始日: %s\n", cfg.StartDate)
	fmt.Printf("終了日: %s\n", cfg.EndDate)
	fmt.Printf("アカウント: %s\n", cfg.Account)
	fmt.Printf("プロパティ数: %d\n", len(cfg.Properties))

	for i, property := range cfg.Properties {
		fmt.Printf("\nプロパティ %d:\n", i+1)
		fmt.Printf("  ID: %s\n", property.ID)
		fmt.Printf("  ストリーム数: %d\n", len(property.Streams))

		for j, stream := range property.Streams {
			fmt.Printf("  ストリーム %d:\n", j+1)
			fmt.Printf("    ID: %s\n", stream.ID)
			fmt.Printf("    ベースURL: %s\n", stream.BaseURL)
			fmt.Printf("    ディメンション: %v\n", stream.Dimensions)
			fmt.Printf("    メトリクス: %v\n", stream.Metrics)
		}
	}
}