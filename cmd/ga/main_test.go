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

package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gaerrors "github.com/ymotongpoo/ga/internal/errors"
)

func TestNewCLIApp(t *testing.T) {
	app := NewCLIApp()
	if app == nil {
		t.Fatal("NewCLIApp() returned nil")
	}

	// 初期状態では各サービスはnilであることを確認
	if app.authService != nil {
		t.Error("authService should be nil initially")
	}
	if app.configService != nil {
		t.Error("configService should be nil initially")
	}
	if app.analyticsService != nil {
		t.Error("analyticsService should be nil initially")
	}
	if app.outputService != nil {
		t.Error("outputService should be nil initially")
	}
}

func TestCLIApp_initializeServices(t *testing.T) {
	app := NewCLIApp()
	err := app.initializeServices()
	if err != nil {
		t.Fatalf("initializeServices() error = %v", err)
	}

	// 出力サービスが初期化されていることを確認
	if app.outputService == nil {
		t.Error("outputService should be initialized")
	}

	// 他のサービスはまだnilであることを確認（未実装のため）
	if app.authService != nil {
		t.Error("authService should still be nil (not implemented)")
	}
	if app.configService != nil {
		t.Error("configService should still be nil (not implemented)")
	}
	if app.analyticsService != nil {
		t.Error("analyticsService should still be nil (not implemented)")
	}
}

func TestCLIApp_getExitCodeFromError(t *testing.T) {
	app := NewCLIApp()

	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{
			name:     "AuthError",
			err:      gaerrors.NewAuthError("認証エラー", nil),
			wantCode: 1,
		},
		{
			name:     "ConfigError",
			err:      gaerrors.NewConfigError("設定エラー", nil),
			wantCode: 2,
		},
		{
			name:     "APIError",
			err:      gaerrors.NewAPIError("APIエラー", nil),
			wantCode: 1,
		},
		{
			name:     "OutputError",
			err:      gaerrors.NewOutputError("出力エラー", nil),
			wantCode: 1,
		},
		{
			name:     "Generic error",
			err:      errors.New("generic error"),
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := app.getExitCodeFromError(tt.err)
			if got != tt.wantCode {
				t.Errorf("getExitCodeFromError() = %v, want %v", got, tt.wantCode)
			}
		})
	}
}

func TestCLIApp_parseArgs(t *testing.T) {
	app := NewCLIApp()

	tests := []struct {
		name    string
		args    []string
		want    *CLIOptions
		wantErr bool
	}{
		{
			name: "Default options",
			args: []string{},
			want: &CLIOptions{
				ConfigPath: "ga.yaml",
				OutputPath: "",
				Debug:      false,
				Help:       false,
				Version:    false,
				Login:      false,
			},
			wantErr: false,
		},
		{
			name: "Custom config path",
			args: []string{"--config", "custom.yaml"},
			want: &CLIOptions{
				ConfigPath: "custom.yaml",
				OutputPath: "",
				Debug:      false,
				Help:       false,
				Version:    false,
				Login:      false,
			},
			wantErr: false,
		},
		{
			name: "Output path specified",
			args: []string{"--output", "data.csv"},
			want: &CLIOptions{
				ConfigPath: "ga.yaml",
				OutputPath: "data.csv",
				Debug:      false,
				Help:       false,
				Version:    false,
				Login:      false,
			},
			wantErr: false,
		},
		{
			name: "Debug mode enabled",
			args: []string{"--debug"},
			want: &CLIOptions{
				ConfigPath: "ga.yaml",
				OutputPath: "",
				Debug:      true,
				Help:       false,
				Version:    false,
				Login:      false,
			},
			wantErr: false,
		},
		{
			name: "Help flag",
			args: []string{"--help"},
			want: &CLIOptions{
				ConfigPath: "ga.yaml",
				OutputPath: "",
				Debug:      false,
				Help:       true,
				Version:    false,
				Login:      false,
			},
			wantErr: false,
		},
		{
			name: "Help flag short form",
			args: []string{"-h"},
			want: &CLIOptions{
				ConfigPath: "ga.yaml",
				OutputPath: "",
				Debug:      false,
				Help:       true,
				Version:    false,
				Login:      false,
			},
			wantErr: false,
		},
		{
			name: "Version flag",
			args: []string{"--version"},
			want: &CLIOptions{
				ConfigPath: "ga.yaml",
				OutputPath: "",
				Debug:      false,
				Help:       false,
				Version:    true,
				Login:      false,
			},
			wantErr: false,
		},
		{
			name: "Version flag short form",
			args: []string{"-v"},
			want: &CLIOptions{
				ConfigPath: "ga.yaml",
				OutputPath: "",
				Debug:      false,
				Help:       false,
				Version:    true,
				Login:      false,
			},
			wantErr: false,
		},
		{
			name: "Login flag",
			args: []string{"--login"},
			want: &CLIOptions{
				ConfigPath: "ga.yaml",
				OutputPath: "",
				Debug:      false,
				Help:       false,
				Version:    false,
				Login:      true,
			},
			wantErr: false,
		},
		{
			name: "Multiple flags",
			args: []string{"--config", "test.yaml", "--output", "output.csv", "--debug"},
			want: &CLIOptions{
				ConfigPath: "test.yaml",
				OutputPath: "output.csv",
				Debug:      true,
				Help:       false,
				Version:    false,
				Login:      false,
			},
			wantErr: false,
		},
		{
			name:    "Invalid flag",
			args:    []string{"--invalid"},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.parseArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != nil && tt.want != nil {
				if got.ConfigPath != tt.want.ConfigPath {
					t.Errorf("parseArgs() ConfigPath = %v, want %v", got.ConfigPath, tt.want.ConfigPath)
				}
				if got.OutputPath != tt.want.OutputPath {
					t.Errorf("parseArgs() OutputPath = %v, want %v", got.OutputPath, tt.want.OutputPath)
				}
				if got.Debug != tt.want.Debug {
					t.Errorf("parseArgs() Debug = %v, want %v", got.Debug, tt.want.Debug)
				}
				if got.Help != tt.want.Help {
					t.Errorf("parseArgs() Help = %v, want %v", got.Help, tt.want.Help)
				}
				if got.Version != tt.want.Version {
					t.Errorf("parseArgs() Version = %v, want %v", got.Version, tt.want.Version)
				}
				if got.Login != tt.want.Login {
					t.Errorf("parseArgs() Login = %v, want %v", got.Login, tt.want.Login)
				}
			}
		})
	}
}

func TestCLIApp_Run_Help(t *testing.T) {
	app := NewCLIApp()
	app.initializeServices()

	// ヘルプフラグのテスト
	exitCode := app.Run(context.Background(), []string{"--help"})
	if exitCode != 0 {
		t.Errorf("Run() with --help = %v, want 0", exitCode)
	}

	// ヘルプフラグ短縮形のテスト
	exitCode = app.Run(context.Background(), []string{"-h"})
	if exitCode != 0 {
		t.Errorf("Run() with -h = %v, want 0", exitCode)
	}
}

func TestCLIApp_Run_Version(t *testing.T) {
	app := NewCLIApp()
	app.initializeServices()

	// バージョンフラグのテスト
	exitCode := app.Run(context.Background(), []string{"--version"})
	if exitCode != 0 {
		t.Errorf("Run() with --version = %v, want 0", exitCode)
	}

	// バージョンフラグ短縮形のテスト
	exitCode = app.Run(context.Background(), []string{"-v"})
	if exitCode != 0 {
		t.Errorf("Run() with -v = %v, want 0", exitCode)
	}
}

func TestCLIApp_Run_Login(t *testing.T) {
	app := NewCLIApp()
	app.initializeServices()

	// ログインフラグのテスト（認証サービスが未実装なのでエラーになる）
	exitCode := app.Run(context.Background(), []string{"--login"})
	if exitCode == 0 {
		t.Error("Run() with --login should return non-zero exit code when auth service is not implemented")
	}
}

func TestCLIApp_Run_InvalidArgs(t *testing.T) {
	app := NewCLIApp()
	app.initializeServices()

	// 無効な引数のテスト
	exitCode := app.Run(context.Background(), []string{"--invalid"})
	if exitCode != 2 {
		t.Errorf("Run() with invalid args = %v, want 2", exitCode)
	}
}

func TestCLIApp_Run_DataRetrieval_ConfigNotFound(t *testing.T) {
	app := NewCLIApp()
	app.initializeServices()

	// 存在しない設定ファイルのテスト
	exitCode := app.Run(context.Background(), []string{"--config", "nonexistent.yaml"})
	if exitCode == 0 {
		t.Error("Run() with nonexistent config file should return non-zero exit code")
	}
}

func TestCLIApp_Run_DataRetrieval_ServicesNotImplemented(t *testing.T) {
	app := NewCLIApp()
	app.initializeServices()

	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.yaml")
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
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// サービスが未実装なのでエラーになる
	exitCode := app.Run(context.Background(), []string{"--config", configFile})
	if exitCode == 0 {
		t.Error("Run() should return non-zero exit code when services are not implemented")
	}
}

func TestCLIApp_handleLogin(t *testing.T) {
	app := NewCLIApp()
	app.initializeServices()

	ctx := context.Background()

	// 認証サービスが未実装の場合のテスト
	err := app.handleLogin(ctx)
	if err == nil {
		t.Error("handleLogin() should return error when auth service is not implemented")
	}

	expectedError := "認証サービスが初期化されていません"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("handleLogin() error should contain '%s', got: %s", expectedError, err.Error())
	}
}

func TestCLIApp_handleDataRetrieval_ConfigNotFound(t *testing.T) {
	app := NewCLIApp()
	app.initializeServices()

	ctx := context.Background()
	options := &CLIOptions{
		ConfigPath: "nonexistent.yaml",
		OutputPath: "",
		Debug:      false,
	}

	err := app.handleDataRetrieval(ctx, options)
	if err == nil {
		t.Error("handleDataRetrieval() should return error when config file does not exist")
	}

	expectedError := "が見つかりません"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("handleDataRetrieval() error should contain '%s', got: %s", expectedError, err.Error())
	}
}

func TestCLIApp_handleDataRetrieval_ServicesNotImplemented(t *testing.T) {
	app := NewCLIApp()
	app.initializeServices()

	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.yaml")
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
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	ctx := context.Background()
	options := &CLIOptions{
		ConfigPath: configFile,
		OutputPath: "",
		Debug:      false,
	}

	err = app.handleDataRetrieval(ctx, options)
	if err == nil {
		t.Error("handleDataRetrieval() should return error when services are not implemented")
	}

	expectedError := "未実装です"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("handleDataRetrieval() error should contain '%s', got: %s", expectedError, err.Error())
	}
}

func TestCLIApp_handleDataRetrieval_DebugMode(t *testing.T) {
	app := NewCLIApp()
	app.initializeServices()

	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.yaml")
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
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	ctx := context.Background()
	options := &CLIOptions{
		ConfigPath: configFile,
		OutputPath: "output.csv",
		Debug:      true,
	}

	// デバッグモードでも最終的にはサービス未実装エラーになる
	err = app.handleDataRetrieval(ctx, options)
	if err == nil {
		t.Error("handleDataRetrieval() should return error when services are not implemented")
	}
}

func TestCLIOptions_Structure(t *testing.T) {
	options := &CLIOptions{
		ConfigPath: "test.yaml",
		OutputPath: "output.csv",
		Debug:      true,
		Help:       false,
		Version:    false,
		Login:      true,
	}

	if options.ConfigPath != "test.yaml" {
		t.Errorf("ConfigPath = %s, want test.yaml", options.ConfigPath)
	}

	if options.OutputPath != "output.csv" {
		t.Errorf("OutputPath = %s, want output.csv", options.OutputPath)
	}

	if !options.Debug {
		t.Error("Debug should be true")
	}

	if options.Help {
		t.Error("Help should be false")
	}

	if options.Version {
		t.Error("Version should be false")
	}

	if !options.Login {
		t.Error("Login should be true")
	}
}

func TestCommand_Structure(t *testing.T) {
	handler := func(args []string) error {
		return nil
	}

	command := &Command{
		Name:        "test",
		Description: "Test command",
		Handler:     handler,
	}

	if command.Name != "test" {
		t.Errorf("Name = %s, want test", command.Name)
	}

	if command.Description != "Test command" {
		t.Errorf("Description = %s, want Test command", command.Description)
	}

	if command.Handler == nil {
		t.Error("Handler should not be nil")
	}

	// ハンドラーの実行テスト
	err := command.Handler([]string{})
	if err != nil {
		t.Errorf("Handler() error = %v", err)
	}
}

// テスト用のヘルパー関数
func createTestConfigFile(t *testing.T, content string) string {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.yaml")
	err := os.WriteFile(configFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	return configFile
}

func TestCLIApp_Integration_HelpAndVersion(t *testing.T) {
	app := NewCLIApp()
	app.initializeServices()

	// 統合テスト: ヘルプとバージョンの組み合わせ
	tests := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{
			name:     "Help only",
			args:     []string{"--help"},
			wantCode: 0,
		},
		{
			name:     "Version only",
			args:     []string{"--version"},
			wantCode: 0,
		},
		{
			name:     "Help with other flags (help takes precedence)",
			args:     []string{"--help", "--debug"},
			wantCode: 0,
		},
		{
			name:     "Version with other flags (version takes precedence)",
			args:     []string{"--version", "--debug"},
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode := app.Run(context.Background(), tt.args)
			if exitCode != tt.wantCode {
				t.Errorf("Run() = %v, want %v", exitCode, tt.wantCode)
			}
		})
	}
}

func TestCLIApp_Integration_ErrorHandling(t *testing.T) {
	app := NewCLIApp()
	app.initializeServices()

	// 統合テスト: エラーハンドリング
	tests := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{
			name:     "Invalid flag",
			args:     []string{"--invalid-flag"},
			wantCode: 2,
		},
		{
			name:     "Login with unimplemented service",
			args:     []string{"--login"},
			wantCode: 1,
		},
		{
			name:     "Nonexistent config file",
			args:     []string{"--config", "nonexistent.yaml"},
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode := app.Run(context.Background(), tt.args)
			if exitCode != tt.wantCode {
				t.Errorf("Run() = %v, want %v", exitCode, tt.wantCode)
			}
		})
	}
}
