# 設計書

## 概要

Google Analytics 4 (GA4) トラッキングツールは、GoでCLIアプリケーションとして実装されます。OAuth認証、YAML設定ファイル処理、Google Analytics 4 API統合、CSV・JSON出力機能を提供し、Webサイトのトラッキングデータを効率的に取得・分析できるツールです。

## アーキテクチャ

### 全体アーキテクチャ

```mermaid
graph TB
    CLI[CLI Interface] --> Auth[Authentication Module]
    CLI --> Config[Configuration Module]
    CLI --> Analytics[Analytics Client]
    CLI --> Output[Output Module]

    Auth --> OAuth[OAuth2 Flow]
    Auth --> Token[Token Storage]

    Config --> YAML[YAML Parser]
    Config --> Validator[Config Validator]

    Analytics --> GA4[GA4 API Client]
    Analytics --> Retry[Retry Logic]

    Output --> CSV[CSV Writer]
    Output --> JSON[JSON Writer]
    Output --> Console[Console Output]

    GA4 --> API[Google Analytics 4 API]
```

### レイヤー構造

1. **CLI Layer**: コマンドライン引数の処理とユーザーインターフェース
2. **Service Layer**: ビジネスロジックとワークフロー制御
3. **Client Layer**: 外部API（Google Analytics 4）との通信
4. **Infrastructure Layer**: 認証、設定、出力処理

## コンポーネントと インターフェース

### 1. CLI Interface (`cmd/ga/main.go`)

```go
type CLIApp struct {
    authService      AuthService
    configService    ConfigService
    analyticsService AnalyticsService
    outputService    OutputService
}

type Command struct {
    Name        string
    Description string
    Handler     func(args []string) error
}

type CLIOptions struct {
    ConfigPath   string
    OutputPath   string
    OutputFormat string  // "csv" または "json"
    Debug        bool
    Help         bool
    Version      bool
    Login        bool
}
```

**責任:**
- コマンドライン引数の解析
- サブコマンド（--login, --help, --version, --config）の処理
- エラーハンドリングとユーザーフィードバック
- 適切な終了コードの管理

### 2. Authentication Service (`internal/auth/`)

```go
type AuthService interface {
    Login(ctx context.Context) error
    GetCredentials(ctx context.Context) (*oauth2.Token, error)
    RefreshToken(ctx context.Context) error
    IsAuthenticated(ctx context.Context) bool
    ClearToken() error
}

type OAuth2Config struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
    Scopes       []string
}

type LocalServer struct {
    server   *http.Server
    authCode chan string
    errChan  chan error
}
```

**責任:**
- OAuth2オフラインアクセスフローの管理
- ローカルHTTPサーバーによるリダイレクト処理
- トークンの保存と更新
- 認証状態の検証

**オフラインアクセスフロー実装:**
1. ローカルHTTPサーバー（`http://localhost:8080/callback`）を起動
2. OAuth2認証URLを生成してブラウザで開く
3. ユーザーが認証を完了するとリダイレクトでコールバックを受信
4. 認証コードを取得してトークンに交換
5. リフレッシュトークンを含むトークンを安全に保存
6. ローカルサーバーを停止

### 3. Configuration Service (`internal/config/`)

```go
type ConfigService interface {
    LoadConfig(path string) (*Config, error)
    ValidateConfig(config *Config) error
}

type Config struct {
    StartDate   string     `yaml:"start_date"`
    EndDate     string     `yaml:"end_date"`
    Account     string     `yaml:"account"`
    Properties  []Property `yaml:"properties"`
}

type Property struct {
    ID      string   `yaml:"property"`
    Streams []Stream `yaml:"streams"`
}

type Stream struct {
    ID         string   `yaml:"stream"`
    BaseURL    string   `yaml:"base_url,omitempty"`
    Dimensions []string `yaml:"dimensions"`
    Metrics    []string `yaml:"metrics"`
}
```

**責任:**
- YAML設定ファイルの読み込み
- 設定値の検証
- デフォルト値の適用

### 4. Analytics Service (`internal/analytics/`)

```go
type AnalyticsService interface {
    GetReportData(ctx context.Context, config *Config) (*ReportData, error)
}

type GA4Client struct {
    service *analyticsreporting.Service
    config  *Config
}

type ReportData struct {
    Headers []string
    Rows    [][]string
    Summary ReportSummary
    StreamURLs map[string]string // ストリームID -> ベースURL のマッピング
}

type ReportSummary struct {
    TotalRows    int
    DateRange    string
    Properties   []string
}
```

**責任:**
- Google Analytics 4 APIとの通信
- データ取得とフォーマット
- エラーハンドリングとリトライ

### 5. Output Service (`internal/output/`)

```go
type OutputService interface {
    WriteCSV(data *ReportData, writer io.Writer) error
    WriteJSON(data *ReportData, writer io.Writer) error
    WriteToFile(data *ReportData, filename string, format OutputFormat) error
    WriteToConsole(data *ReportData, format OutputFormat) error
}

type OutputFormat int

const (
    FormatCSV OutputFormat = iota
    FormatJSON
)

type CSVWriter struct {
    encoding string
    delimiter rune
}

type JSONWriter struct {
    encoding string
    indent   string
}

type URLProcessor struct {
    streamURLs map[string]string
}

func (up *URLProcessor) ProcessPagePath(streamID, pagePath string) string

type JSONRecord struct {
    Dimensions map[string]string `json:"dimensions"`
    Metrics    map[string]string `json:"metrics"`
    Metadata   JSONMetadata      `json:"metadata"`
}

type JSONMetadata struct {
    RetrievedAt  string `json:"retrieved_at"`
    PropertyID   string `json:"property_id"`
    StreamID     string `json:"stream_id,omitempty"`
    DateRange    string `json:"date_range"`
}
```

**責任:**
- CSV形式とJSON形式でのデータ出力
- ファイル出力と標準出力の管理
- UTF-8エンコーディング処理
- ストリームURLとpagePathの結合処理
- JSON構造化データの生成

## データモデル

### 1. 設定データモデル

```yaml
# ga.yaml の構造
start_date: "2023-01-01"
end_date: "2023-01-31"
account: "123456789"
properties:
  - property: "987654321"
    streams:
      - stream: "1234567"
        base_url: "https://example.com"  # オプション: ストリームのベースURL
        dimensions:
          - "date"
          - "pagePath"
        metrics:
          - "sessions"
          - "activeUsers"
          - "newUsers"
          - "averageSessionDuration"
```

### 2. API レスポンスデータモデル

```go
type GAReportResponse struct {
    DimensionHeaders []DimensionHeader
    MetricHeaders    []MetricHeader
    Rows             []ReportRow
}

type ReportRow struct {
    Dimensions []string
    Metrics    []MetricValue
}

type MetricValue struct {
    Value string
}
```

### 3. 出力データモデル

#### CSV出力形式

```csv
date,fullURL,sessions,activeUsers,newUsers,averageSessionDuration
2023-01-01,https://example.com/home,1250,1100,850,120.5
2023-01-01,https://example.com/about,450,420,380,95.2
```

#### JSON出力形式

```json
[
  {
    "dimensions": {
      "date": "2023-01-01",
      "fullURL": "https://example.com/home"
    },
    "metrics": {
      "sessions": "1250",
      "activeUsers": "1100",
      "newUsers": "850",
      "averageSessionDuration": "120.5"
    },
    "metadata": {
      "retrieved_at": "2023-02-01T10:30:00Z",
      "property_id": "987654321",
      "stream_id": "1234567",
      "date_range": "2023-01-01 to 2023-01-31"
    }
  },
  {
    "dimensions": {
      "date": "2023-01-01",
      "fullURL": "https://example.com/about"
    },
    "metrics": {
      "sessions": "450",
      "activeUsers": "420",
      "newUsers": "380",
      "averageSessionDuration": "95.2"
    },
    "metadata": {
      "retrieved_at": "2023-02-01T10:30:00Z",
      "property_id": "987654321",
      "stream_id": "1234567",
      "date_range": "2023-01-01 to 2023-01-31"
    }
  }
]
```

### 4. URL結合処理ロジック

```go
// URL結合処理の例
func ProcessPagePath(baseURL, pagePath string) string {
    // 絶対URLの場合はそのまま返す
    if strings.HasPrefix(pagePath, "http://") || strings.HasPrefix(pagePath, "https://") {
        return pagePath
    }

    // ベースURLが設定されていない場合はpagePathをそのまま返す
    if baseURL == "" {
        return pagePath
    }

    // pagePathが空の場合はベースURLのみ返す
    if pagePath == "" {
        return baseURL
    }

    // スラッシュの重複を処理してURL結合
    baseURL = strings.TrimSuffix(baseURL, "/")
    pagePath = strings.TrimPrefix(pagePath, "/")

    return fmt.Sprintf("%s/%s", baseURL, pagePath)
}
```

## エラーハンドリング

### エラー分類

1. **認証エラー**: OAuth認証失敗、トークン期限切れ
2. **設定エラー**: YAML形式エラー、必須項目不足
3. **APIエラー**: Google Analytics API制限、ネットワークエラー
4. **出力エラー**: ファイル書き込み権限、ディスク容量不足

### エラーハンドリング戦略

```go
type GAError struct {
    Type    ErrorType
    Message string
    Cause   error
}

type ErrorType int

const (
    AuthError ErrorType = iota
    ConfigError
    APIError
    OutputError
)

func (e *GAError) Error() string {
    return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
}
```

### リトライ戦略

- **API制限エラー**: 指数バックオフで最大3回リトライ
- **ネットワークエラー**: 線形バックオフで最大5回リトライ
- **認証エラー**: 自動トークン更新を1回試行

## テスト戦略

### 1. 単体テスト

- **対象**: 各サービスとユーティリティ関数
- **モック**: Google Analytics API、ファイルシステム
- **カバレッジ**: 80%以上を目標

### 2. 統合テスト

- **対象**: サービス間の連携
- **テストデータ**: サンプルYAML設定ファイル
- **モック**: 外部API呼び出し

### 3. エンドツーエンドテスト

- **対象**: CLI全体のワークフロー
- **環境**: テスト用Google Analytics プロパティ
- **検証**: CSV出力の正確性

### テスト構造

```
tests/
├── unit/
│   ├── auth_test.go
│   ├── config_test.go
│   ├── analytics_test.go
│   └── output_test.go
├── integration/
│   ├── service_integration_test.go
│   └── api_integration_test.go
├── e2e/
│   └── cli_e2e_test.go
└── testdata/
    ├── valid_config.yaml
    ├── invalid_config.yaml
    └── sample_response.json
```

## OAuth2オフラインアクセスフロー詳細

### フロー概要

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant CLI as CLIアプリ
    participant Server as ローカルサーバー
    participant Browser as ブラウザ
    participant Google as Google OAuth2

    User->>CLI: ga --login
    CLI->>Server: HTTPサーバー起動 (localhost:8080)
    CLI->>Browser: 認証URL開く
    Browser->>Google: 認証リクエスト
    Google->>User: 認証画面表示
    User->>Google: 認証情報入力
    Google->>Server: リダイレクト (/callback?code=...)
    Server->>CLI: 認証コード送信
    CLI->>Google: 認証コード→トークン交換
    Google->>CLI: アクセストークン + リフレッシュトークン
    CLI->>CLI: トークン保存
    CLI->>Server: サーバー停止
    CLI->>User: 認証完了通知
```

### 実装詳細

```go
type LocalServer struct {
    server   *http.Server
    authCode chan string
    errChan  chan error
    port     int
}

func (ls *LocalServer) Start(ctx context.Context) error {
    mux := http.NewServeMux()
    mux.HandleFunc("/callback", ls.handleCallback)

    ls.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", ls.port),
        Handler: mux,
    }

    go func() {
        if err := ls.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            ls.errChan <- err
        }
    }()

    return nil
}

func (ls *LocalServer) handleCallback(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    if code == "" {
        ls.errChan <- errors.New("認証コードが取得できませんでした")
        return
    }

    ls.authCode <- code
    fmt.Fprintf(w, "認証が完了しました。このタブを閉じてください。")
}
```

## セキュリティ考慮事項

### 1. 認証情報の保護

- OAuth2トークンの暗号化保存
- 設定ファイルでの平文パスワード禁止
- 環境変数での機密情報管理
- ローカルサーバーのポート制限（localhost のみ）

### 2. API通信のセキュリティ

- HTTPS通信の強制
- TLS証明書の検証
- リクエスト/レスポンスのログ制限
- OAuth2 state パラメータによるCSRF攻撃防止

### 3. ファイルアクセス制御

- 設定ファイルの適切な権限設定
- 出力ファイルの安全な作成
- 一時ファイルの適切な削除
- トークンファイルの権限制限（0600）

### 4. ローカルサーバーのセキュリティ

- localhost のみでのバインド
- 認証完了後の即座なサーバー停止
- タイムアウト設定による自動停止
- 不正なリクエストの適切な処理

## パフォーマンス考慮事項

### 1. API呼び出し最適化

- バッチリクエストの活用
- 並行処理での複数プロパティ取得
- キャッシュ機能（オプション）

### 2. メモリ使用量最適化

- ストリーミング処理でのCSV出力
- 大量データの分割処理
- ガベージコレクション最適化

### 3. 実行時間最適化

- 設定ファイルの事前検証
- 早期エラー検出
- プログレス表示機能