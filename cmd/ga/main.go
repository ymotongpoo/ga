package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ymotongpoo/ga/internal/analytics"
	"github.com/ymotongpoo/ga/internal/auth"
	"github.com/ymotongpoo/ga/internal/config"
	"github.com/ymotongpoo/ga/internal/output"
)

func main() {
	ctx := context.Background()

	app := &CLIApp{}

	if err := app.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// CLIApp はメインアプリケーション構造体
type CLIApp struct {
	authService      auth.AuthService
	configService    config.ConfigService
	analyticsService analytics.AnalyticsService
	outputService    output.OutputService
}

// Run はCLIアプリケーションのメインエントリーポイント
func (app *CLIApp) Run(ctx context.Context, args []string) error {
	// TODO: 実装予定
	return fmt.Errorf("not implemented yet")
}

// CLIOptions はコマンドライン引数を表す構造体
type CLIOptions struct {
	ConfigPath string
	OutputPath string
	Debug      bool
	Help       bool
	Version    bool
	Login      bool
}

// Command はサブコマンドを表す構造体
type Command struct {
	Name        string
	Description string
	Handler     func(args []string) error
}