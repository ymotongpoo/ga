# クイックスタートガイド

Google Analytics 4 データ取得ツールを素早く開始するためのガイドです。

## 1. 事前準備（5分）

### Google Cloud Console での設定

1. [Google Cloud Console](https://console.cloud.google.com/) にアクセス
2. 新しいプロジェクトを作成または既存のプロジェクトを選択
3. 「APIとサービス」→「ライブラリ」から「Google Analytics Reporting API」を検索して有効化
4. 「APIとサービス」→「認証情報」→「認証情報を作成」→「OAuth クライアント ID」を選択
5. アプリケーションの種類で「デスクトップアプリケーション」を選択
6. 名前を入力（例：「GA4 Data Tool」）
7. 作成後、クライアントIDとクライアントシークレットをメモ

### Google Analytics での確認

1. [Google Analytics](https://analytics.google.com/) にアクセス
2. 管理画面で以下の情報を確認・メモ：
   - アカウントID
   - プロパティID
   - データストリームID

## 2. ツールのセットアップ（2分）

### インストール

```bash
# リポジトリをクローン
git clone https://github.com/ymotongpoo/ga
cd ga

# ビルド
go build -o ga cmd/ga/main.go
```

### 環境変数の設定

```bash
# OAuth認証情報を環境変数に設定
export GA_CLIENT_ID="your-client-id.googleusercontent.com"
export GA_CLIENT_SECRET="your-client-secret"
```

## 3. 初回認証（1分）

```bash
# OAuth認証を実行
./ga --login
```

ブラウザが開くので、Googleアカウントでログインし、アクセスを許可してください。

## 4. 設定ファイルの作成（2分）

```bash
# 設定ファイル例をコピー
cp ga.yaml.example ga.yaml
```

`ga.yaml` を編集して、実際の値に変更：

```yaml
start_date: "2024-01-01"  # 取得したい期間の開始日
end_date: "2024-01-31"    # 取得したい期間の終了日
account: "123456789"      # あなたのアカウントID
properties:
  - property: "987654321"  # あなたのプロパティID
    streams:
      - stream: "1234567"  # あなたのストリームID
        dimensions:
          - "date"
          - "pagePath"
        metrics:
          - "sessions"
          - "activeUsers"
          - "newUsers"
          - "averageSessionDuration"
```

## 5. データ取得の実行（1分）

```bash
# データを取得して標準出力に表示
./ga

# CSVファイルに出力
./ga --output data.csv

# デバッグ情報付きで実行
./ga --debug
```

## 完了！

これで Google Analytics 4 からデータを取得できるようになりました。

## よくある問題と解決方法

### 認証エラー

```
認証エラー: OAuth認証に必要な環境変数が設定されていません
```

**解決方法**: 環境変数 `GA_CLIENT_ID` と `GA_CLIENT_SECRET` が正しく設定されているか確認

### 設定ファイルエラー

```
設定ファイルの検証に失敗しました: account ID の形式が不正です
```

**解決方法**: アカウント、プロパティ、ストリームIDが数字のみで構成されているか確認

### API エラー

```
APIエラー: プロパティにアクセスする権限がありません
```

**解決方法**: Google Analytics でプロパティへのアクセス権限があることを確認

## 次のステップ

- [README.md](README.md) で詳細な使用方法を確認
- 複数プロパティの設定方法を学習
- 異なるディメンションとメトリクスを試す
- 定期実行のためのスクリプト作成

## サポート

問題が発生した場合は、GitHub Issues でお知らせください。