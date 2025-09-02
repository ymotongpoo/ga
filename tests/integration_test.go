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

package tests

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ymotongpoo/ga/internal/analytics"
	"github.com/ymotongpoo/ga/internal/auth"
	"github.com/ymotongpoo/ga/internal/config"
	"github.com/ymotongpoo/ga/internal/errors"
	"github.com/ymotongpoo/ga/internal/logger"
	"github.com/ymotongpoo/ga/internal/output"
)

// TestConfigServiceIntegration は設定サービスの統合テスト
func TestConfigServiceIntegration(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// 有効な設定ファイルを作成
	validConfigPath := filepath.Join(tempDir, "valid_config.yaml")
	validConfigContent := `
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
          - "newUsers"
          - "averageSessionDuration"
`
	err := os.WriteFile(validConfigPath, []byte(validConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create valid config file: %v", err)
	}

	// 設定サービスを作成
	configService := config.NewConfigService()

	// 設定ファイルの読み込みテスト
	cfg, err := configService.LoadConfig(validConfigPath)
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// 設定の検証テスト
	err = configService.ValidateConfig(cfg)
	if err != nil {
		t.Fatalf("ValidateConfig() failed: %v", err)
	}

	// 設定内容の確認
	if cfg.StartDate != "2023-01-01" {
		t.Errorf("StartDate = %s, want 2023-01-01", cfg.StartDate)
	}

	if cfg.EndDate != "2023-01-31" {
		t.Errorf("EndDate = %s, want 2023-01-31", cfg.EndDate)
	}

	if cfg.Account != "123456789" {
		t.Errorf("Account = %s, want 123456789", cfg.Account)
	}

	if len(cfg.Properties) != 1 {
		t.Fatalf("Properties length = %d, want 1", len(cfg.Properties))
	}

	property := cfg.Properties[0]
	if property.ID != "987654321" {
		t.Errorf("Property ID = %s, want 987654321", property.ID)
	}

	if len(property.Streams) != 1 {
		t.Fatalf("Streams length = %d, want 1", len(property.Streams))
	}

	stream := property.Streams[0]
	if stream.ID != "1234567" {
		t.Errorf("Stream ID = %s, want 1234567", stream.ID)
	}

	expectedDimensions := []string{"date", "pagePath"}
	if len(stream.Dimensions) != len(expectedDimensions) {
		t.Errorf("Dimensions length = %d, want %d", len(stream.Dimensions), len(expectedDimensions))
	}

	expectedMetrics := []string{"sessions", "activeUsers", "newUsers", "averageSessionDuration"}
	if len(stream.Metrics) != len(expectedMetrics) {
		t.Errorf("Metrics length = %d, want %d", len(stream.Metrics), len(expectedMetrics))
	}
}

// TestConfigServiceIntegration_InvalidConfig は無効な設定ファイルの統合テスト
func TestConfigServiceIntegration_InvalidConfig(t *testing.T) {
	tempDir := t.TempDir()
	configService := config.NewConfigService()

	tests := []struct {
		name           string
		configContent  string
		expectedError  string
	}{
		{
			name: "Missing start_date",
			configContent: `
end_date: "2023-01-31"
account: "123456789"
properties:
  - property: "987654321"
    streams:
      - stream: "1234567"
        dimensions: ["date"]
        metrics: ["sessions"]
`,
			expectedError: "start_date は必須項目です",
		},
		{
			name: "Invalid date format",
			configContent: `
start_date: "2023/01/01"
end_date: "2023-01-31"
account: "123456789"
properties:
  - property: "987654321"
    streams:
      - stream: "1234567"
        dimensions: ["date"]
        metrics: ["sessions"]
`,
			expectedError: "start_date の形式が不正です",
		},
		{
			name: "Invalid account ID",
			configContent: `
start_date: "2023-01-01"
end_date: "2023-01-31"
account: "abc123"
properties:
  - property: "987654321"
    streams:
      - stream: "1234567"
        dimensions: ["date"]
        metrics: ["sessions"]
`,
			expectedError: "account ID の形式が不正です",
		},
		{
			name: "Invalid metric",
			configContent: `
start_date: "2023-01-01"
end_date: "2023-01-31"
account: "123456789"
properties:
  - property: "987654321"
    streams:
      - stream: "1234567"
        dimensions: ["date"]
        metrics: ["invalidMetric"]
`,
			expectedError: "無効なメトリクスが含まれています",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, tt.name+".yaml")
			err := os.WriteFile(configPath, []byte(tt.configContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create config file: %v", err)
			}

			cfg, err := configService.LoadConfig(configPath)
			if err != nil {
				// YAML解析エラーの場合
				if !strings.Contains(err.Error(), "YAML形式が不正です") {
					t.Fatalf("Unexpected load error: %v", err)
				}
				return
			}

			err = configService.ValidateConfig(cfg)
			if err == nil {
				t.Errorf("ValidateConfig() should fail for %s", tt.name)
				return
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("ValidateConfig() error = %v, should contain %s", err, tt.expectedError)
			}
		})
	}
}

// TestAuthServiceIntegration は認証サービスの統合テスト
func TestAuthServiceIntegration(t *testing.T) {
	// テスト用の認証設定
	config := &auth.OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes:       []string{"https://www.googleapis.com/auth/analytics.readonly"},
	}

	authService := auth.NewAuthService(config)

	// 認証状態の確認（トークンが存在しない場合）
	ctx := context.Background()
	if authService.IsAuthenticated(ctx) {
		t.Error("IsAuthenticated() should return false when no token exists")
	}

	// 認証情報の取得（トークンが存在しない場合）
	_, err := authService.GetCredentials(ctx)
	if err == nil {
		t.Error("GetCredentials() should fail when no token exists")
	}

	// エラーがGAErrorであることを確認
	if gaErr, ok := err.(*errors.GAError); ok {
		if gaErr.Type != errors.AuthError {
			t.Errorf("GetCredentials() error type = %v, want %v", gaErr.Type, errors.AuthError)
		}
	} else {
		t.Errorf("GetCredentials() should return GAError, got %T", err)
	}

	// トークン情報の取得（トークンが存在しない場合）
	_, err = authService.GetTokenInfo()
	if err == nil {
		t.Error("GetTokenInfo() should fail when no token exists")
	}

	// トークンのクリア（存在しない場合でもエラーにならない）
	err = authService.ClearToken()
	if err != nil {
		t.Errorf("ClearToken() should not fail when no token exists: %v", err)
	}
}

// TestOutputServiceIntegration は出力サービスの統合テスト
func TestOutputServiceIntegration(t *testing.T) {
	outputService := output.NewOutputService()

	// テスト用のレポートデータを作成
	testData := createTestReportData()

	// 一時ディレクトリを作成
	tempDir := t.TempDir()

	// ファイル出力のテスト
	outputFile := filepath.Join(tempDir, "test_output.csv")
	err := outputService.WriteToFile(testData, outputFile)
	if err != nil {
		t.Fatalf("WriteToFile() failed: %v", err)
	}

	// ファイルが作成されたことを確認
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// ファイル内容を確認
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	lines := strings.Split(strings.TrimSpace(contentStr), "\n")

	// ヘッダー行の確認
	expectedHeader := "property_id,date,pagePath,sessions,activeUsers,newUsers,averageSessionDuration"
	if lines[0] != expectedHeader {
		t.Errorf("Header = %s, want %s", lines[0], expectedHeader)
	}

	// データ行数の確認
	expectedLines := len(testData.Rows) + 1 // ヘッダー + データ行
	if len(lines) != expectedLines {
		t.Errorf("Lines count = %d, want %d", len(lines), expectedLines)
	}

	// WriteOutputメソッドのテスト（ファイル出力）
	outputFile2 := filepath.Join(tempDir, "test_output2.csv")
	err = outputService.WriteOutput(testData, outputFile2)
	if err != nil {
		t.Fatalf("WriteOutput() to file failed: %v", err)
	}

	// ファイルが作成されたことを確認
	if _, err := os.Stat(outputFile2); os.IsNotExist(err) {
		t.Fatal("Output file from WriteOutput() was not created")
	}
}

// TestLoggerIntegration はログ機能の統合テスト
func TestLoggerIntegration(t *testing.T) {
	// グローバルロガーのリセット
	logger.InitGlobalLogger(true)

	// 各種ログ機能のテスト
	logger.Debug("Debug message: %s", "test")
	logger.Info("Info message: %d", 123)
	logger.Warn("Warning message")
	logger.Error("Error message: %v", "error")

	// API関連のログ機能のテスト
	headers := map[string]string{
		"Authorization": "Bearer token",
		"Content-Type":  "application/json",
	}
	logger.LogAPIRequest("GET", "https://api.example.com/data", headers)
	logger.LogAPIResponse(200, 1024)

	// 設定関連のログ機能のテスト
	logger.LogConfigLoad("/path/to/config.yaml", true)
	logger.LogConfigLoad("/path/to/invalid.yaml", false)

	// データ処理関連のログ機能のテスト
	logger.LogDataProcessing(50, 100)
	logger.LogDataProcessing(100, 100)

	// デバッグモードの確認
	if !logger.IsDebugMode() {
		t.Error("IsDebugMode() should return true")
	}
}

// TestErrorHandlingIntegration はエラーハンドリングの統合テスト
func TestErrorHandlingIntegration(t *testing.T) {
	// 各種エラーの作成と確認
	authErr := errors.NewAuthError("認証に失敗しました", nil)
	configErr := errors.NewConfigError("設定ファイルが不正です", nil)
	apiErr := errors.NewAPIError("API呼び出しに失敗しました", nil)
	outputErr := errors.NewOutputError("ファイル出力に失敗しました", nil)
	networkErr := errors.NewNetworkError("ネットワーク接続に失敗しました", nil)
	validationErr := errors.NewValidationError("バリデーションに失敗しました", nil)

	// エラータイプの確認
	if authErr.Type != errors.AuthError {
		t.Errorf("AuthError type = %v, want %v", authErr.Type, errors.AuthError)
	}
	if configErr.Type != errors.ConfigError {
		t.Errorf("ConfigError type = %v, want %v", configErr.Type, errors.ConfigError)
	}
	if apiErr.Type != errors.APIError {
		t.Errorf("APIError type = %v, want %v", apiErr.Type, errors.APIError)
	}
	if outputErr.Type != errors.OutputError {
		t.Errorf("OutputError type = %v, want %v", outputErr.Type, errors.OutputError)
	}
	if networkErr.Type != errors.NetworkError {
		t.Errorf("NetworkError type = %v, want %v", networkErr.Type, errors.NetworkError)
	}
	if validationErr.Type != errors.ValidationError {
		t.Errorf("ValidationError type = %v, want %v", validationErr.Type, errors.ValidationError)
	}

	// エラーメッセージの確認
	authErrMsg := authErr.Error()
	if !strings.Contains(authErrMsg, "AUTH_ERROR") {
		t.Errorf("AuthError message should contain AUTH_ERROR: %s", authErrMsg)
	}

	// ユーザーフレンドリーメッセージの確認
	userMsg := authErr.GetUserFriendlyMessage()
	if !strings.Contains(userMsg, "認証エラー") {
		t.Errorf("User friendly message should contain 認証エラー: %s", userMsg)
	}
	if !strings.Contains(userMsg, "ga --login") {
		t.Errorf("User friendly message should contain ga --login: %s", userMsg)
	}

	// コンテキスト情報の追加
	authErr.WithContext("user_id", "12345")
	authErr.WithContext("timestamp", "2023-01-01T00:00:00Z")

	if authErr.Context["user_id"] != "12345" {
		t.Errorf("Context user_id = %v, want 12345", authErr.Context["user_id"])
	}
	if authErr.Context["timestamp"] != "2023-01-01T00:00:00Z" {
		t.Errorf("Context timestamp = %v, want 2023-01-01T00:00:00Z", authErr.Context["timestamp"])
	}
}

// TestServiceInteraction はサービス間の相互作用テスト
func TestServiceInteraction(t *testing.T) {
	// 設定サービスと出力サービスの連携テスト
	tempDir := t.TempDir()

	// 設定ファイルを作成
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `
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
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// 設定サービスで設定を読み込み
	configService := config.NewConfigService()
	cfg, err := configService.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	err = configService.ValidateConfig(cfg)
	if err != nil {
		t.Fatalf("ValidateConfig() failed: %v", err)
	}

	// 設定情報を使用してテストデータを作成
	testData := createTestReportDataFromConfig(cfg)

	// 出力サービスでデータを出力
	outputService := output.NewOutputService()
	outputPath := filepath.Join(tempDir, "output.csv")
	err = outputService.WriteOutput(testData, outputPath)
	if err != nil {
		t.Fatalf("WriteOutput() failed: %v", err)
	}

	// 出力ファイルの確認
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// ファイル内容の確認
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, cfg.Properties[0].ID) {
		t.Error("Output should contain property ID from config")
	}
}

// ヘルパー関数: テスト用のレポートデータを作成
func createTestReportData() *analytics.ReportData {
	return &analytics.ReportData{
		Headers: []string{"property_id", "date", "pagePath", "sessions", "activeUsers", "newUsers", "averageSessionDuration"},
		Rows: [][]string{
			{"123456789", "2023-01-01", "/home", "1250", "1100", "850", "120.5"},
			{"123456789", "2023-01-01", "/about", "450", "420", "380", "95.2"},
			{"123456789", "2023-01-02", "/home", "1180", "1050", "780", "115.8"},
		},
		Summary: analytics.ReportSummary{
			TotalRows:  3,
			DateRange:  "2023-01-01 - 2023-01-02",
			Properties: []string{"123456789"},
		},
	}
}

// ヘルパー関数: 設定から テスト用のレポートデータを作成
func createTestReportDataFromConfig(cfg *config.Config) *analytics.ReportData {
	propertyID := cfg.Properties[0].ID

	return &analytics.ReportData{
		Headers: []string{"property_id", "date", "pagePath", "sessions", "activeUsers"},
		Rows: [][]string{
			{propertyID, cfg.StartDate, "/home", "1250", "1100"},
			{propertyID, cfg.StartDate, "/about", "450", "420"},
		},
		Summary: analytics.ReportSummary{
			TotalRows:  2,
			DateRange:  cfg.StartDate + " - " + cfg.EndDate,
			Properties: []string{propertyID},
		},
	}
}

// TestCompleteWorkflow は完全なワークフローの統合テスト
func TestCompleteWorkflow(t *testing.T) {
	tempDir := t.TempDir()

	// 1. 設定ファイルの作成
	configPath := filepath.Join(tempDir, "workflow_config.yaml")
	configContent := `
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
          - "newUsers"
          - "averageSessionDuration"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// 2. 設定の読み込みと検証
	configService := config.NewConfigService()
	cfg, err := configService.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	err = configService.ValidateConfig(cfg)
	if err != nil {
		t.Fatalf("ValidateConfig() failed: %v", err)
	}

	// 3. ログ機能の初期化
	logger.InitGlobalLogger(true)
	logger.LogConfigLoad(configPath, true)

	// 4. テストデータの作成（実際のAPIの代わり）
	testData := createTestReportDataFromConfig(cfg)

	// 5. データの出力
	outputService := output.NewOutputService()
	outputPath := filepath.Join(tempDir, "workflow_output.csv")
	err = outputService.WriteOutput(testData, outputPath)
	if err != nil {
		t.Fatalf("WriteOutput() failed: %v", err)
	}

	// 6. 結果の検証
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	lines := strings.Split(strings.TrimSpace(contentStr), "\n")

	// ヘッダーの確認
	if !strings.Contains(lines[0], "property_id") {
		t.Error("Output should contain property_id header")
	}

	// データ行の確認
	if len(lines) < 2 {
		t.Error("Output should contain at least one data row")
	}

	// プロパティIDの確認
	if !strings.Contains(contentStr, cfg.Properties[0].ID) {
		t.Error("Output should contain the correct property ID")
	}

	// 7. ログ出力の確認
	logger.Info("Workflow test completed successfully")
	logger.LogDataProcessing(len(testData.Rows), len(testData.Rows))
}