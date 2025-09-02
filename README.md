# Google Analytics 4 (GA4) トラッキングツール

このツールは、Google Analytics 4 (GA4) を使用して、各URLに関して次のデータのトラッキングを行うためのツールです。

- セッション数
- アクティブユーザー
- 新規ユーザー数
- セッションあたりの平均エンゲージメント時間

## 前提条件

- プロパティが設定済みの Google Analytics 4 (GA4) のアカウント

## 技術要件

- Go 1.23以上
- Google Analytics 4用SDK (`google.golang.org/api/analyticsreporting/v4`)

## インストール方法

```bash
go install github.com/ymotongpoo/ga
```

## 使用方法

まずGoogle Analytics APIのためのクレデンシャルを取得するために `--login` オプションでログインをします。

```bash
ga --login
```

ログインした後、設定ファイルに基づいてGoogle Analytics 4からデータを取得します。
設定ファイルの書式は以下のとおりです。

```yaml
# ga.yaml
start_date: 2023-01-01 # 集計開始日
end_date: 2023-01-31   # 集計終了日
account: 123456789     # アカウントID
  - property: 987654321  # プロパティID
      - stream: 1234567  # ストリームID
          # ディメンションのリスト
          # date: 日付
          # pagePath: URLパス
          dimensions:
            - date
            - pagePath
          # 取得するメトリクス
          # sessions: セッション数
          # activeUsers: アクティブユーザー数
          # newUsers: 新規ユーザー数
          # engagementRate: エンゲージメント平均時間
          metrics:
            - sessions
            - activeUsers
            - newUsers
            - engagementRate
```

結果はCSVで出力します。出力する際には、 `pagePath` はストリームIDの設定に応じたフルパスとして出力します。
