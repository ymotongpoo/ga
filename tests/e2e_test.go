package tests

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCLI_Help はCLIのヘルプ機能のエンドツーエンドテスト
func TestCLI_Help(t *testing.T) {
	// バイナリをビルド
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// ヘルプコマンドを実行
	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Help command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// ヘルプメッセージの内容を確認
	expectedStrings := []string{
		"ga - Google Analytics 4 データ取得ツール",
		"使用方法:",
		"--config",
		"--output",
		"--debug",
		"--login",
		"--help",
		"--version",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Help output should contain '%s'\nActual output: %s", expected, outputStr)
		}
	}

	// 終了コードが0であることを確認
	if cmd.ProcessState.ExitCode() != 0 {
		t.Errorf("Help command should exit with code 0, got %d", cmd.ProcessState.ExitCode())
	}
}

// TestCLI_Version はCLIのバージョン機能のエンドツーエンドテスト
func TestCLI_Version(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// バージョンコマンドを実行
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Version command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// バージョンメッセージの内容を確認
	expectedStrings := []string{
		"ga version",
		"Google Analytics 4 データ取得ツール",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Version output should contain '%s'\nActual output: %s", expected, outputStr)
		}
	}

	// 終了コードが0であることを確認
	if cmd.ProcessState.ExitCode() != 0 {
		t.Errorf("Version command should exit with code 0, got %d", cmd.ProcessState.ExitCode())
	}
}

// TestCLI_InvalidFlag は無効なフラグのエンドツーエンドテスト
func TestCLI_InvalidFlag(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// 無効なフラグでコマンドを実行
	cmd := exec.Command(binaryPath, "--invalid-flag")
	output, err := cmd.CombinedOutput()

	// エラーが発生することを期待
	if err == nil {
		t.Fatal("Invalid flag should cause an error")
	}

	outputStr := string(output)

	// エラーメッセージの内容を確認
	if !strings.Contains(outputStr, "無効なオプション") {
		t.Errorf("Error output should contain '無効なオプション'\nActual output: %s", outputStr)
	}

	// 終了コードが2であることを確認（使用方法エラー）
	if cmd.ProcessState.ExitCode() != 2 {
		t.Errorf("Invalid flag should exit with code 2, got %d", cmd.ProcessState.ExitCode())
	}
}

// TestCLI_Login はログイン機能のエンドツーエンドテスト
func TestCLI_Login(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// ログインコマンドを実行（認証サービスが未実装なのでエラーになる）
	cmd := exec.Command(binaryPath, "--login")
	output, err := cmd.CombinedOutput()

	// エラーが発生することを期待（認証サービス未実装のため）
	if err == nil {
		t.Fatal("Login command should fail when auth service is not implemented")
	}

	outputStr := string(output)

	// エラーメッセージの内容を確認
	if !strings.Contains(outputStr, "認証サービスが初期化されていません") {
		t.Errorf("Login error should contain '認証サービスが初期化されていません'\nActual output: %s", outputStr)
	}

	// 終了コードが1であることを確認
	if cmd.ProcessState.ExitCode() != 1 {
		t.Errorf("Login command should exit with code 1, got %d", cmd.ProcessState.ExitCode())
	}
}

// TestCLI_ConfigNotFound は設定ファイルが見つからない場合のエンドツーエンドテスト
func TestCLI_ConfigNotFound(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// 存在しない設定ファイルを指定してコマンドを実行
	cmd := exec.Command(binaryPath, "--config", "nonexistent.yaml")
	output, err := cmd.CombinedOutput()

	// エラーが発生することを期待
	if err == nil {
		t.Fatal("Command should fail when config file does not exist")
	}

	outputStr := string(output)

	// エラーメッセージの内容を確認
	if !strings.Contains(outputStr, "が見つかりません") {
		t.Errorf("Error output should contain 'が見つかりません'\nActual output: %s", outputStr)
	}

	// 終了コードが1であることを確認
	if cmd.ProcessState.ExitCode() != 1 {
		t.Errorf("Config not found should exit with code 1, got %d", cmd.ProcessState.ExitCode())
	}
}

// TestCLI_ValidConfig は有効な設定ファイルでのエンドツーエンドテスト
func TestCLI_ValidConfig(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")
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
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// 設定ファイルを指定してコマンドを実行
	cmd := exec.Command(binaryPath, "--config", configPath)
	output, err := cmd.CombinedOutput()

	// サービスが未実装なのでエラーになることを期待
	if err == nil {
		t.Fatal("Command should fail when services are not implemented")
	}

	outputStr := string(output)

	// 設定ファイルが読み込まれることを確認
	if !strings.Contains(outputStr, "設定ファイル") {
		t.Errorf("Output should mention config file\nActual output: %s", outputStr)
	}

	// 未実装エラーが発生することを確認
	if !strings.Contains(outputStr, "未実装です") {
		t.Errorf("Error output should contain '未実装です'\nActual output: %s", outputStr)
	}

	// 終了コードが1であることを確認
	if cmd.ProcessState.ExitCode() != 1 {
		t.Errorf("Unimplemented service should exit with code 1, got %d", cmd.ProcessState.ExitCode())
	}
}

// TestCLI_DebugMode はデバッグモードのエンドツーエンドテスト
func TestCLI_DebugMode(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "debug_config.yaml")
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
        metrics:
          - "sessions"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// デバッグモードでコマンドを実行
	cmd := exec.Command(binaryPath, "--config", configPath, "--debug")
	output, err := cmd.CombinedOutput()

	// サービスが未実装なのでエラーになることを期待
	if err == nil {
		t.Fatal("Command should fail when services are not implemented")
	}

	outputStr := string(output)

	// デバッグ情報が出力されることを確認
	if !strings.Contains(outputStr, "[DEBUG]") {
		t.Errorf("Debug mode should output debug information\nActual output: %s", outputStr)
	}

	// 設定ファイルパスがデバッグ出力に含まれることを確認
	if !strings.Contains(outputStr, configPath) {
		t.Errorf("Debug output should contain config file path\nActual output: %s", outputStr)
	}
}

// TestCLI_OutputPath は出力パス指定のエンドツーエンドテスト
func TestCLI_OutputPath(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "output_config.yaml")
	outputPath := filepath.Join(tempDir, "output.csv")
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
        metrics:
          - "sessions"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// 出力パスを指定してコマンドを実行
	cmd := exec.Command(binaryPath, "--config", configPath, "--output", outputPath)
	output, err := cmd.CombinedOutput()

	// サービスが未実装なのでエラーになることを期待
	if err == nil {
		t.Fatal("Command should fail when services are not implemented")
	}

	outputStr := string(output)

	// 出力先が言及されることを確認
	if !strings.Contains(outputStr, outputPath) {
		t.Errorf("Output should mention output path\nActual output: %s", outputStr)
	}
}

// TestCLI_ShortFlags は短縮フラグのエンドツーエンドテスト
func TestCLI_ShortFlags(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	tests := []struct {
		name     string
		flag     string
		expected string
	}{
		{
			name:     "Short help flag",
			flag:     "-h",
			expected: "使用方法:",
		},
		{
			name:     "Short version flag",
			flag:     "-v",
			expected: "ga version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.flag)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s command failed: %v\nOutput: %s", tt.name, err, output)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, tt.expected) {
				t.Errorf("%s output should contain '%s'\nActual output: %s", tt.name, tt.expected, outputStr)
			}

			// 終了コードが0であることを確認
			if cmd.ProcessState.ExitCode() != 0 {
				t.Errorf("%s should exit with code 0, got %d", tt.name, cmd.ProcessState.ExitCode())
			}
		})
	}
}

// TestCLI_MultipleFlags は複数フラグの組み合わせのエンドツーエンドテスト
func TestCLI_MultipleFlags(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "multi_config.yaml")
	outputPath := filepath.Join(tempDir, "multi_output.csv")
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
        metrics:
          - "sessions"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// 複数のフラグを指定してコマンドを実行
	cmd := exec.Command(binaryPath, "--config", configPath, "--output", outputPath, "--debug")
	output, err := cmd.CombinedOutput()

	// サービスが未実装なのでエラーになることを期待
	if err == nil {
		t.Fatal("Command should fail when services are not implemented")
	}

	outputStr := string(output)

	// 各フラグの効果が確認できることをチェック
	if !strings.Contains(outputStr, "[DEBUG]") {
		t.Errorf("Debug flag should enable debug output\nActual output: %s", outputStr)
	}

	if !strings.Contains(outputStr, configPath) {
		t.Errorf("Config path should be mentioned\nActual output: %s", outputStr)
	}

	if !strings.Contains(outputStr, outputPath) {
		t.Errorf("Output path should be mentioned\nActual output: %s", outputStr)
	}
}

// TestCLI_Timeout はコマンドタイムアウトのテスト
func TestCLI_Timeout(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// タイムアウト付きでヘルプコマンドを実行
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Help command with timeout failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "使用方法:") {
		t.Errorf("Help output should contain usage information\nActual output: %s", outputStr)
	}
}

// buildTestBinary はテスト用のバイナリをビルドする
func buildTestBinary(t *testing.T) string {
	// 一時ディレクトリにバイナリを作成
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "ga_test")

	// プロジェクトルートディレクトリを取得
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// testsディレクトリからプロジェクトルートに移動
	projectRoot := filepath.Dir(wd)

	// Goバイナリをビルド
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/ga")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build test binary: %v\nOutput: %s", err, output)
	}

	return binaryPath
}

// TestCLI_Integration_FullWorkflow は完全なワークフローの統合テスト
func TestCLI_Integration_FullWorkflow(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	tempDir := t.TempDir()

	// 1. ヘルプの確認
	helpCmd := exec.Command(binaryPath, "--help")
	helpOutput, err := helpCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Help command failed: %v", err)
	}
	if !strings.Contains(string(helpOutput), "使用方法:") {
		t.Error("Help should display usage information")
	}

	// 2. バージョンの確認
	versionCmd := exec.Command(binaryPath, "--version")
	versionOutput, err := versionCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Version command failed: %v", err)
	}
	if !strings.Contains(string(versionOutput), "ga version") {
		t.Error("Version should display version information")
	}

	// 3. 無効なフラグのテスト
	invalidCmd := exec.Command(binaryPath, "--invalid")
	invalidOutput, err := invalidCmd.CombinedOutput()
	if err == nil {
		t.Error("Invalid flag should cause an error")
	}
	if !strings.Contains(string(invalidOutput), "無効なオプション") {
		t.Error("Invalid flag should show error message")
	}

	// 4. 設定ファイルが見つからない場合のテスト
	notFoundCmd := exec.Command(binaryPath, "--config", "nonexistent.yaml")
	notFoundOutput, err := notFoundCmd.CombinedOutput()
	if err == nil {
		t.Error("Nonexistent config should cause an error")
	}
	if !strings.Contains(string(notFoundOutput), "見つかりません") {
		t.Error("Nonexistent config should show not found error")
	}

	// 5. 有効な設定ファイルでのテスト
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
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	validConfigCmd := exec.Command(binaryPath, "--config", configPath, "--debug")
	validConfigOutput, err := validConfigCmd.CombinedOutput()
	// サービス未実装によりエラーになることを期待
	if err == nil {
		t.Error("Valid config should fail due to unimplemented services")
	}
	outputStr := string(validConfigOutput)
	if !strings.Contains(outputStr, "設定ファイル") {
		t.Error("Valid config should mention config file")
	}
	if !strings.Contains(outputStr, "[DEBUG]") {
		t.Error("Debug flag should enable debug output")
	}

	t.Log("Full workflow integration test completed successfully")
}

// TestCLI_ErrorCodes は各種エラーコードのテスト
func TestCLI_ErrorCodes(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	tests := []struct {
		name         string
		args         []string
		expectedCode int
		description  string
	}{
		{
			name:         "Help command",
			args:         []string{"--help"},
			expectedCode: 0,
			description:  "Help should exit with code 0",
		},
		{
			name:         "Version command",
			args:         []string{"--version"},
			expectedCode: 0,
			description:  "Version should exit with code 0",
		},
		{
			name:         "Invalid flag",
			args:         []string{"--invalid"},
			expectedCode: 2,
			description:  "Invalid flag should exit with code 2",
		},
		{
			name:         "Login command",
			args:         []string{"--login"},
			expectedCode: 1,
			description:  "Login should exit with code 1 (unimplemented)",
		},
		{
			name:         "Nonexistent config",
			args:         []string{"--config", "nonexistent.yaml"},
			expectedCode: 1,
			description:  "Nonexistent config should exit with code 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			_, err := cmd.CombinedOutput()

			actualCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					actualCode = exitError.ExitCode()
				}
			}

			if actualCode != tt.expectedCode {
				t.Errorf("%s: expected exit code %d, got %d", tt.description, tt.expectedCode, actualCode)
			}
		})
	}
}