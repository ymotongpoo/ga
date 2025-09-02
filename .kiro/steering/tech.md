# 技術スタック

## ビルドシステム
- Go標準のビルドシステム（`go build`、`go install`）

## 技術スタック
- プログラミング言語: Go 1.23以上
- 主要ライブラリ: `google.golang.org/api/analyticsreporting/v4`
- 設定ファイル形式: YAML
- 出力形式: CSV

## 開発環境
- Go バージョン: 1.23以上
- Google Analytics 4 API アクセス権限が必要

## よく使用するコマンド

### ビルド
```bash
go build
```

### インストール
```bash
go install github.com/ymotongpoo/ga
```

### テスト
```bash
go test ./...
```

### 実行（ログイン）
```bash
ga --login
```

### 実行（データ取得）
```bash
ga  # ga.yamlファイルを使用
```

### モジュール管理
```bash
go mod tidy
go mod download
```

## コーディング規約
- Go標準のフォーマット（`go fmt`）を使用
- `go vet`でコード品質をチェック
- エラーハンドリングはGoの慣例に従う

## 設定ファイル
- `ga.yaml`: メインの設定ファイル
- Google Analytics APIクレデンシャルファイル（OAuth認証後に生成）

## 注意事項
- Google Analytics 4のプロパティが事前に設定されている必要があります
- APIクレデンシャルの取得と認証が必要です