package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ymotongpoo/ga/internal/analytics"
	"github.com/ymotongpoo/ga/internal/auth"
	"github.com/ymotongpoo/ga/internal/config"
	"github.com/ymotongpoo/ga/internal/errors"
	"github.com/ymotongpoo/ga/internal/output"
)

func main() {
	ctx := context.Background()

	app := NewCLIApp()

	// サービスの初期化
	if err := app.initializeServices(); err != nil {
		fmt.Fprintf(os.Stderr, "サービス初期化エラー: %v\n", err)
		os.Exit(1)
	}

	exitCode := app.Run(ctx, os.Args[1:])
	os.Exit(exitCode)
}

// CLIApp はメインアプリケーション構造体
type CLIApp struct {
	authService      auth.AuthService
	configService    config.ConfigService
	analyticsService analytics.AnalyticsService
	outputService    output.OutputService
}

// NewCLIApp は新しいCLIAppインスタンスを作成する
func NewCLIApp() *CLIApp {
	return &CLIApp{
		// サービスは後で初期化される
		// 現在は未実装のため nil のまま
	}
}

// initializeServices はサービスを初期化する
func (app *CLIApp) initializeServices() error {
	// TODO: 各サービスの実装後に初期化処理を追加
	// app.authService = auth.NewAuthService()
	// app.configService = config.NewConfigService()
	// app.analyticsService = analytics.NewAnalyticsService()
	// app.outputService = output.NewOutputService()
	return nil
}

// getExitCodeFromError はエラーから適切な終了コードを取得する
func (app *CLIApp) getExitCodeFromError(err error) int {
	if gaErr, ok := err.(*errors.GAError); ok {
		switch gaErr.Type {
		case errors.AuthError:
			return 1 // 認証エラー
		case errors.ConfigError:
			return 2 // 設定エラー
		case errors.APIError:
			return 1 // APIエラー
		case errors.OutputError:
			return 1 // 出力エラー
		default:
			return 1 // その他のエラー
		}
	}
	return 1 // 一般的なエラー
}

// Run はCLIアプリケーションのメインエントリーポイント
// 適切な終了コードを返す（0: 成功, 1: 一般的なエラー, 2: 使用方法エラー）
func (app *CLIApp) Run(ctx context.Context, args []string) int {
	options, err := app.parseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2 // 使用方法エラー
	}

	// parseArgsがnilを返した場合（ヘルプ表示など）
	if options == nil {
		return 0
	}

	// ヘルプオプションの処理
	if options.Help {
		app.showHelp()
		return 0
	}

	// バージョンオプションの処理
	if options.Version {
		app.showVersion()
		return 0
	}

	// ログインオプションの処理
	if options.Login {
		if err := app.handleLogin(ctx); err != nil {
			exitCode := app.getExitCodeFromError(err)
			fmt.Fprintf(os.Stderr, "認証エラー: %v\n", err)
			return exitCode
		}
		return 0
	}

	// デフォルト動作: データ取得
	if err := app.handleDataRetrieval(ctx, options); err != nil {
		exitCode := app.getExitCodeFromError(err)
		fmt.Fprintf(os.Stderr, "データ取得エラー: %v\n", err)
		return exitCode
	}

	return 0
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

	// 設定ファイルパスの検証
	if options.ConfigPath == "" {
		return nil, fmt.Errorf("設定ファイルパスが指定されていません")
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
	fmt.Println("OAuth認証を開始します...")

	// 認証サービスが未実装の場合の処理
	if app.authService == nil {
		return fmt.Errorf("認証サービスが初期化されていません。認証機能は未実装です")
	}

	// 認証処理を実行
	if err := app.authService.Login(ctx); err != nil {
		return fmt.Errorf("OAuth認証に失敗しました: %w", err)
	}

	fmt.Println("OAuth認証が完了しました")
	return nil
}

// handleDataRetrieval はデータ取得処理を実行する
func (app *CLIApp) handleDataRetrieval(ctx context.Context, options *CLIOptions) error {
	if options.Debug {
		fmt.Printf("[DEBUG] 設定ファイル: %s\n", options.ConfigPath)
		fmt.Printf("[DEBUG] 出力先: %s\n", options.OutputPath)
	}

	fmt.Printf("設定ファイル '%s' を使用してデータを取得します...\n", options.ConfigPath)

	// 設定ファイルの存在確認
	if _, err := os.Stat(options.ConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("設定ファイル '%s' が見つかりません", options.ConfigPath)
	}

	// サービスが未実装の場合の処理
	if app.configService == nil || app.analyticsService == nil || app.outputService == nil {
		if options.OutputPath != "" {
			fmt.Printf("結果を '%s' に出力します\n", options.OutputPath)
		} else {
			fmt.Println("結果を標準出力に出力します")
		}
		return fmt.Errorf("データ取得機能は未実装です。必要なサービスが初期化されていません")
	}

	// 設定ファイルの読み込み
	config, err := app.configService.LoadConfig(options.ConfigPath)
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	// 設定の検証
	if err := app.configService.ValidateConfig(config); err != nil {
		return fmt.Errorf("設定ファイルの検証に失敗しました: %w", err)
	}

	// データ取得
	reportData, err := app.analyticsService.GetReportData(ctx, config)
	if err != nil {
		return fmt.Errorf("データ取得に失敗しました: %w", err)
	}

	// データ出力
	if options.OutputPath != "" {
		if err := app.outputService.WriteToFile(reportData, options.OutputPath); err != nil {
			return fmt.Errorf("ファイル出力に失敗しました: %w", err)
		}
		fmt.Printf("結果を '%s' に出力しました\n", options.OutputPath)
	} else {
		if err := app.outputService.WriteToConsole(reportData); err != nil {
			return fmt.Errorf("標準出力に失敗しました: %w", err)
		}
	}

	fmt.Printf("データ取得が完了しました。取得レコード数: %d\n", reportData.Summary.TotalRows)
	return nil
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