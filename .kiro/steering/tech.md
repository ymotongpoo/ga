---
inclusion: always
---

# 技術スタック・開発規約

## 技術スタック
- **言語**: Go 1.23以上
- **主要ライブラリ**: `google.golang.org/api/analyticsreporting/v4`
- **設定**: YAML形式（`ga.yaml`）
- **出力**: CSV形式
- **認証**: OAuth 2.0（OOBフロー禁止、オフラインアクセスフロー必須）

## アーキテクチャパターン
- CLIアプリケーション（`cmd/ga/main.go`）
- 内部パッケージ構成（`internal/`配下）
- 設定ファイル駆動型
- Google Analytics API クライアントパターン

## コーディング規約
- Go標準フォーマット（`go fmt`）必須
- `go vet`によるコード品質チェック必須
- Goの慣例に従ったエラーハンドリング
- パッケージ名: 小文字、短く、意味のある名前
- ファイル名: スネークケース（例: `analytics_client.go`）
- 関数名: キャメルケース（公開関数は大文字開始）

## 必須コマンド
```bash
# ビルド・テスト
go build
go test ./...
go mod tidy

# 実行
ga --login    # 初回認証
ga           # データ取得（ga.yaml使用）
```

## 重要な制約
- Google Analytics 4プロパティの事前設定必須
- OAuth 2.0は[オフラインアクセスフロー](https://developers.google.com/identity/protocols/oauth2/resources/oob-migration?hl=ja#server-side-offline-access_2)を実装（OOBフロー禁止）
- `internal/`パッケージは外部アクセス不可