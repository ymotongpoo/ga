<!--
 Copyright 2025 Yoshi Yamaguchi
 
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
 
     https://www.apache.org/licenses/LICENSE-2.0
 
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

# Google Analytics 4 (GA4) データ取得ツール

Google Analytics 4 (GA4) を使用して、WebサイトのURLごとのトラッキングデータを取得・分析するためのGoコマンドラインツールです。

## 主な機能

- **セッション数**の取得
- **アクティブユーザー数**の取得
- **新規ユーザー数**の取得
- **セッションあたりの平均エンゲージメント時間**の取得
- **YAML設定ファイル**による柔軟な設定
- **CSV形式**でのデータ出力
- **複数プロパティ**の並行データ取得
- **OAuth2認証**によるセキュアなAPI接続

## 前提条件

- Google Analytics 4 (GA4) のアカウントとプロパティが設定済み
- Google Cloud Console でのOAuth2クライアント設定
- Go 1.23以上

## インストール方法

### ソースからビルド

```bash
git clone https://github.com/ymotongpoo/ga
cd ga
go build -o ga cmd/ga/main.go
```

### Go install（将来対応予定）

```bash
go install github.com/ymotongpoo/ga@latest
```

## セットアップ

### 1. Google Cloud Console での設定

1. [Google Cloud Console](https://console.cloud.google.com/) にアクセス
2. プロジェクトを作成または選択
3. Google Analytics Reporting API を有効化
4. OAuth 2.0 クライアント ID を作成
5. クライアント ID とクライアントシークレットを取得

### 2. 環境変数の設定

OAuth認証に必要な環境変数を設定します：

```bash
export GA_CLIENT_ID="your-client-id.googleusercontent.com"
export GA_CLIENT_SECRET="your-client-secret"
```

### 3. 初回認証

```bash
ga --login
```

ブラウザが開き、Google アカウントでの認証を求められます。認証完了後、認証トークンがローカルに保存されます。

## 使用方法

### 基本的な使用方法

```bash
# デフォルト設定ファイル（ga.yaml）を使用してデータ取得
ga

# カスタム設定ファイルを使用
ga --config custom.yaml

# CSVファイルに出力
ga --output data.csv

# デバッグモードで実行
ga --debug
```

### コマンドラインオプション

| オプション | 短縮形 | 説明 |
|-----------|--------|------|
| `--config PATH` | | 設定ファイルのパス（デフォルト: ga.yaml） |
| `--output PATH` | | 出力ファイルのパス（未指定時は標準出力） |
| `--debug` | | デバッグモードを有効にする |
| `--login` | | OAuth認証を実行する |
| `--help` | `-h` | ヘルプメッセージを表示 |
| `--version` | `-v` | バージョン情報を表示 |

## 設定ファイル

### 基本的な設定ファイル（ga.yaml）

```yaml
start_date: "2024-01-01"  # 集計開始日（YYYY-MM-DD形式）
end_date: "2024-01-31"    # 集計終了日（YYYY-MM-DD形式）
account: "123456789"      # Google Analytics アカウントID

properties:
  - property: "987654321"  # プロパティID
    streams:
      - stream: "1234567"  # データストリームID
        dimensions:        # 取得するディメンション
          - "date"
          - "pagePath"
        metrics:          # 取得するメトリクス
          - "sessions"
          - "activeUsers"
          - "newUsers"
          - "averageSessionDuration"
```

### 複数プロパティの設定例

```yaml
start_date: "2024-01-01"
end_date: "2024-01-31"
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

  - property: "111222333"
    streams:
      - stream: "7654321"
        dimensions:
          - "date"
          - "deviceCategory"
        metrics:
          - "newUsers"
          - "averageSessionDuration"
```

### サポートされるメトリクス

| メトリクス名 | 説明 |
|-------------|------|
| `sessions` | セッション数 |
| `activeUsers` | アクティブユーザー数 |
| `newUsers` | 新規ユーザー数 |
| `averageSessionDuration` | セッションあたりの平均エンゲージメント時間 |

### よく使用されるディメンション

| ディメンション名 | 説明 |
|----------------|------|
| `date` | 日付 |
| `pagePath` | ページパス |
| `deviceCategory` | デバイスカテゴリ |
| `country` | 国 |
| `city` | 都市 |

## 出力形式

### CSV出力例

```csv
property_id,date,pagePath,sessions,activeUsers,newUsers,averageSessionDuration
987654321,20240101,/,1250,1100,850,120.5
987654321,20240101,/about,450,420,380,95.2
987654321,20240102,/,1180,1050,800,115.3
```

### 出力先の指定

```bash
# 標準出力（デフォルト）
ga

# ファイル出力
ga --output data.csv

# パイプでの利用
ga | head -10
```

## エラーハンドリング

### よくあるエラーと対処法

#### 認証エラー

```
認証エラー: 認証トークンが見つかりません
```

**対処法**: `ga --login` で再認証を実行

#### 設定ファイルエラー

```
設定ファイルの検証に失敗しました: start_date は必須項目です
```

**対処法**: 設定ファイルの必須項目を確認・修正

#### API制限エラー

```
API制限エラー: リクエスト数が上限に達しました
```

**対処法**: しばらく待ってから再実行（自動リトライ機能あり）

### 終了コード

| コード | 意味 |
|--------|------|
| 0 | 正常終了 |
| 1 | 一般的なエラー（認証、API、出力エラー） |
| 2 | 使用方法エラー（無効なオプションなど） |

## トラブルシューティング

### デバッグモード

詳細なログ情報を確認するには、`--debug` オプションを使用：

```bash
ga --debug --config my-config.yaml
```

### 認証トークンのリセット

認証に問題がある場合は、保存されたトークンを削除して再認証：

```bash
rm ~/.ga_token.json
ga --login
```

### 設定ファイルの検証

設定ファイルの構文をチェック：

```bash
# 設定ファイルが存在し、正しい形式かテスト
ga --config test.yaml --debug
```

## 開発者向け情報

### プロジェクト構造

```
├── cmd/ga/           # メインアプリケーション
├── internal/         # 内部パッケージ
│   ├── analytics/    # Google Analytics API クライアント
│   ├── auth/         # OAuth2 認証
│   ├── config/       # 設定ファイル処理
│   ├── errors/       # エラーハンドリング
│   ├── logger/       # ログ機能
│   └── output/       # CSV出力
├── tests/            # テストファイル
└── scripts/          # ビルド・テストスクリプト
```

### テスト実行

```bash
# 全テスト実行
go test ./...

# カバレッジ付きテスト
go test -cover ./...

# 統合テスト実行
go test ./tests/... -v
```

### ビルド

```bash
# 開発用ビルド
go build -o ga cmd/ga/main.go

# リリース用ビルド
go build -ldflags="-s -w" -o ga cmd/ga/main.go
```

## ライセンス

このプロジェクトは MIT ライセンスの下で公開されています。

## 貢献

バグ報告や機能要望は、GitHub Issues でお知らせください。プルリクエストも歓迎します。

## サポート

- GitHub Issues: バグ報告・機能要望
- ドキュメント: このREADMEファイル
- Google Analytics 4 API ドキュメント: [公式ドキュメント](https://developers.google.com/analytics/devguides/reporting/data/v1)
