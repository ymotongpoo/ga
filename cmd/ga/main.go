package main

import (
	"context"
	"flag"
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
	options, err := app.parseArgs(args)
	if err != nil {
		return err
	}

	// ヘルプオプションの処理
	if options.Help {
		app.showHelp()
		return nil
	}

	// バージョンオプションの処理
	if options.Version {
		app.showVersion()
		return nil
	}

	// ログインオプションの処理
	if options.Login {
		return app.handleLogin(ctx)
	}

	// デフォルト動作: データ取得
	return app.handleDataRetrieval(ctx, options)
}

// parseArgs はコマンドライン引数を解析してCLIOptionsを返す
func (app *CLIApp) parseArgs(args []string) (*CLIOptions, error) {
	options := &CLIOptions{}

	// カスタムFlagSetを作成してエラーハンドリングを制御
	fs := flag.NewFlagSet("ga", flag.ContinueOnError)
	fs.Usage = func() {
		// カスタムUsage関数で標準エラー出力を抑制
	}

	// フラグの定義
	fs.StringVar(&options.ConfigPath, "config", "ga.yaml", "設定ファイルのパス")
	fs.StringVar(&options.OutputPath, "output", "", "出力ファイルのパス（指定しない場合は標準出力）")
	fs.BoolVar(&options.Debug, "debug", false, "デバッグモードを有効にする")
	fs.BoolVar(&options.Help, "help", false, "ヘルプを表示する")
	fs.BoolVar(&options.Version, "version", false, "バージョン情報を表示する")
	fs.BoolVar(&options.Login, "login", false, "OAuth認証を実行する")

	// 短縮形のフラグも追加
	fs.BoolVar(&options.Help, "h", false, "ヘルプを表示する")
	fs.BoolVar(&options.Version, "v", false, "バージョン情報を表示する")

	// 引数を解析
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			app.showHelp()
			return nil, nil
		}
		return nil, fmt.Errorf("無効なオプションが指定されました: %v\n\n使用方法については 'ga --help' を実行してください", err)
	}

	return options, nil
}

// showHelp はヘルプメッセージを表示する
func (app *CLIApp) showHelp() {
	fmt.Println("ga - Google Analytics 4 データ取得ツール")
	fmt.Println()
	fmt.Println("使用方法:")
	fmt.Println("  ga [オプション]")
	fmt.Println()
	fmt.Println("オプション:")
	fmt.Println("  --config PATH    設定ファイルのパス (デフォルト: ga.yaml)")
	fmt.Println("  --output PATH    出力ファイルのパス (指定しない場合は標準出力)")
	fmt.Println("  --debug          デバッグモードを有効にする")
	fmt.Println("  --login          OAuth認証を実行する")
	fmt.Println("  --help, -h       このヘルプメッセージを表示する")
	fmt.Println("  --version, -v    バージョン情報を表示する")
	fmt.Println()
	fmt.Println("例:")
	fmt.Println("  ga                           # デフォルト設定でデータを取得")
	fmt.Println("  ga --config custom.yaml      # カスタム設定ファイルを使用")
	fmt.Println("  ga --output data.csv         # CSVファイルに出力")
	fmt.Println("  ga --login                   # OAuth認証を実行")
}

// showVersion はバージョン情報を表示する
func (app *CLIApp) showVersion() {
	fmt.Println("ga version 1.0.0")
	fmt.Println("Google Analytics 4 データ取得ツール")
}

// handleLogin はログイン処理を実行する
func (app *CLIApp) handleLogin(ctx context.Context) error {
	// TODO: 認証サービスの実装後に実装
	fmt.Println("OAuth認証を開始します...")
	return fmt.Errorf("認証機能は未実装です")
}

// handleDataRetrieval はデータ取得処理を実行する
func (app *CLIApp) handleDataRetrieval(ctx context.Context, options *CLIOptions) error {
	// TODO: データ取得機能の実装後に実装
	fmt.Printf("設定ファイル '%s' を使用してデータを取得します...\n", options.ConfigPath)
	if options.OutputPath != "" {
		fmt.Printf("結果を '%s' に出力します\n", options.OutputPath)
	} else {
		fmt.Println("結果を標準出力に出力します")
	}
	return fmt.Errorf("データ取得機能は未実装です")
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