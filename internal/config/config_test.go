package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigServiceImpl_LoadConfig(t *testing.T) {
	service := NewConfigService()

	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	t.Run("正常な設定ファイルの読み込み", func(t *testing.T) {
		// 有効な設定ファイルを作成
		validConfig := `
start_date: "2023-01-01"
end_date: "2023-01-31"
account: "123456789"
properties:
  - property: "987654321"
    streams:
      - stream: "1234567"
        dimensions:
          - "date"
          - "pagePath"
        metrics:
          - "sessions"
          - "activeUsers"
`
		configPath := filepath.Join(tempDir, "valid_config.yaml")
		err := os.WriteFile(configPath, []byte(validConfig), 0644)
		if err != nil {
			t.Fatalf("テストファイルの作成に失敗: %v", err)
		}

		config, err := service.LoadConfig(configPath)
		if err != nil {
			t.Errorf("LoadConfig() エラー = %v", err)
			return
		}

		if config.StartDate != "2023-01-01" {
			t.Errorf("StartDate = %v, want %v", config.StartDate, "2023-01-01")
		}
		if config.Account != "123456789" {
			t.Errorf("Account = %v, want %v", config.Account, "123456789")
		}
		if len(config.Properties) != 1 {
			t.Errorf("Properties length = %v, want %v", len(config.Properties), 1)
		}
	})

	t.Run("存在しないファイル", func(t *testing.T) {
		_, err := service.LoadConfig("nonexistent.yaml")
		if err == nil {
			t.Error("存在しないファイルでエラーが発生しませんでした")
		}
	})

	t.Run("不正なYAML形式", func(t *testing.T) {
		invalidConfig := `
start_date: "2023-01-01"
end_date: "2023-01-31"
account: 123456789
properties:
  - property: "987654321"
    streams:
      - stream: "1234567"
        dimensions: [
          - "date"
          - "pagePath"
        metrics:
          - "sessions"
`
		configPath := filepath.Join(tempDir, "invalid_config.yaml")
		err := os.WriteFile(configPath, []byte(invalidConfig), 0644)
		if err != nil {
			t.Fatalf("テストファイルの作成に失敗: %v", err)
		}

		_, err = service.LoadConfig(configPath)
		if err == nil {
			t.Error("不正なYAML形式でエラーが発生しませんでした")
		}
	})
}

func TestConfigServiceImpl_ValidateConfig(t *testing.T) {
	service := &ConfigServiceImpl{}

	t.Run("有効な設定", func(t *testing.T) {
		config := &Config{
			StartDate: "2023-01-01",
			EndDate:   "2023-01-31",
			Account:   "123456789",
			Properties: []Property{
				{
					ID: "987654321",
					Streams: []Stream{
						{
							ID:         "1234567",
							Dimensions: []string{"date", "pagePath"},
							Metrics:    []string{"sessions", "activeUsers"},
						},
					},
				},
			},
		}

		err := service.ValidateConfig(config)
		if err != nil {
			t.Errorf("ValidateConfig() エラー = %v", err)
		}
	})

	t.Run("必須項目不足 - start_date", func(t *testing.T) {
		config := &Config{
			EndDate: "2023-01-31",
			Account: "123456789",
			Properties: []Property{
				{
					ID: "987654321",
					Streams: []Stream{
						{
							ID:         "1234567",
							Dimensions: []string{"date"},
							Metrics:    []string{"sessions"},
						},
					},
				},
			},
		}

		err := service.ValidateConfig(config)
		if err == nil {
			t.Error("start_date不足でエラーが発生しませんでした")
		}
	})

	t.Run("不正な日付形式", func(t *testing.T) {
		config := &Config{
			StartDate: "2023/01/01",
			EndDate:   "2023-01-31",
			Account:   "123456789",
			Properties: []Property{
				{
					ID: "987654321",
					Streams: []Stream{
						{
							ID:         "1234567",
							Dimensions: []string{"date"},
							Metrics:    []string{"sessions"},
						},
					},
				},
			},
		}

		err := service.ValidateConfig(config)
		if err == nil {
			t.Error("不正な日付形式でエラーが発生しませんでした")
		}
	})

	t.Run("不正なアカウントID形式", func(t *testing.T) {
		config := &Config{
			StartDate: "2023-01-01",
			EndDate:   "2023-01-31",
			Account:   "abc123",
			Properties: []Property{
				{
					ID: "987654321",
					Streams: []Stream{
						{
							ID:         "1234567",
							Dimensions: []string{"date"},
							Metrics:    []string{"sessions"},
						},
					},
				},
			},
		}

		err := service.ValidateConfig(config)
		if err == nil {
			t.Error("不正なアカウントID形式でエラーが発生しませんでした")
		}
	})

	t.Run("無効なメトリクス", func(t *testing.T) {
		config := &Config{
			StartDate: "2023-01-01",
			EndDate:   "2023-01-31",
			Account:   "123456789",
			Properties: []Property{
				{
					ID: "987654321",
					Streams: []Stream{
						{
							ID:         "1234567",
							Dimensions: []string{"date"},
							Metrics:    []string{"invalidMetric"},
						},
					},
				},
			},
		}

		err := service.ValidateConfig(config)
		if err == nil {
			t.Error("無効なメトリクスでエラーが発生しませんでした")
		}
	})

	t.Run("開始日が終了日より後", func(t *testing.T) {
		config := &Config{
			StartDate: "2023-02-01",
			EndDate:   "2023-01-31",
			Account:   "123456789",
			Properties: []Property{
				{
					ID: "987654321",
					Streams: []Stream{
						{
							ID:         "1234567",
							Dimensions: []string{"date"},
							Metrics:    []string{"sessions"},
						},
					},
				},
			},
		}

		err := service.ValidateConfig(config)
		if err == nil {
			t.Error("開始日が終了日より後でエラーが発生しませんでした")
		}
	})
}