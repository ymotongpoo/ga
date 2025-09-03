# Requirements Document

## Introduction

Google Analytics 4 (GA4) を使用して、WebサイトのURLごとのトラッキングデータを取得・分析するためのGoコマンドラインツールです。このツールは、セッション数、アクティブユーザー数、新規ユーザー数、セッションあたりの平均エンゲージメント時間を取得し、CSV形式で出力します。

## Requirements

### Requirement 1: Google Analytics API認証

**User Story:** 分析担当者として、Google Analytics 4のデータにアクセスするために、OAuth認証を通じてAPIクレデンシャルを取得したい。

#### Acceptance Criteria

1. WHEN ユーザーが `ga --login` コマンドを実行する THEN システムは OAuth2オフラインアクセスフローを開始する SHALL
2. WHEN OAuth認証フローが開始される THEN システムはローカルHTTPサーバーを起動してリダイレクトを受信する SHALL
3. WHEN OAuth認証が成功する THEN システムは認証トークンをローカルに安全に保存する SHALL
4. IF 認証が失敗する THEN システムは適切なエラーメッセージを表示する SHALL
5. WHEN 認証トークンが期限切れになる THEN システムは自動的にトークンを更新する SHALL
6. WHEN OAuth2設定を行う THEN システムはOOB（Out-of-Band）フローを使用してはならない SHALL
7. WHEN リダイレクトURLを設定する THEN システムは `http://localhost:8080/callback` を使用する SHALL
8. WHEN 認証フローが完了する THEN システムはローカルHTTPサーバーを適切に停止する SHALL

### Requirement 2: YAML設定ファイル処理

**User Story:** データアナリストとして、柔軟にデータ取得条件を設定するために、YAML形式の設定ファイルを使用したい。

#### Acceptance Criteria

1. WHEN システムが起動する THEN `ga.yaml` ファイルを読み込む SHALL
2. IF `ga.yaml` ファイルが存在しない THEN システムは適切なエラーメッセージを表示する SHALL
3. WHEN 設定ファイルに不正な形式が含まれる THEN システムは詳細なエラーメッセージを表示する SHALL
4. WHEN 設定ファイルが読み込まれる THEN 以下の項目を検証する SHALL:
   - start_date（集計開始日）
   - end_date（集計終了日）
   - account（アカウントID）
   - property（プロパティID）
   - stream（ストリームID）
   - dimensions（ディメンションリスト）
   - metrics（メトリクスリスト）

### Requirement 3: Google Analytics 4データ取得

**User Story:** マーケティング担当者として、指定した期間とURLに対するトラッキングデータを取得したい。

#### Acceptance Criteria

1. WHEN 有効な設定ファイルが提供される THEN システムはGoogle Analytics 4 APIに接続する SHALL
2. WHEN APIリクエストが成功する THEN システムは以下のメトリクスを取得する SHALL:
   - セッション数（sessions）
   - アクティブユーザー数（activeUsers）
   - 新規ユーザー数（newUsers）
   - セッションあたりの平均エンゲージメント時間（engagementRateDuration）
3. WHEN APIリクエストが失敗する THEN システムは適切なエラーメッセージとリトライ機能を提供する SHALL
4. WHEN データ取得が完了する THEN システムは取得したレコード数を表示する SHALL
5. WHEN 複数のプロパティが設定されている THEN システムは各プロパティのデータを個別に取得する SHALL
6. WHEN pagePathディメンションが含まれる THEN システムはGoogle Analytics Management APIを使用してストリーム情報を取得する SHALL
7. WHEN ストリーム情報を取得する THEN システムはストリームURLを含むメタデータを取得する SHALL

### Requirement 4: データ出力機能

**User Story:** データアナリストとして、取得したデータを他のツールで分析するために、CSV形式またはJSON形式で出力したい。

#### Acceptance Criteria

1. WHEN データ取得が完了する THEN システムはデフォルトでCSV形式でデータを出力する SHALL
2. WHEN `--format json` オプションが指定される THEN システムはJSON形式でデータを出力する SHALL
3. WHEN `--format csv` オプションが指定される THEN システムはCSV形式でデータを出力する SHALL
4. WHEN 無効な出力形式が指定される THEN システムは適切なエラーメッセージを表示する SHALL
5. WHEN CSV出力が実行される THEN ヘッダー行を含む適切な形式で出力する SHALL
6. WHEN JSON出力が実行される THEN 構造化されたJSON配列形式で出力する SHALL
7. WHEN 出力先が指定されない THEN システムは標準出力にデータを出力する SHALL
8. IF 出力ファイルが指定される THEN システムは指定されたファイルにデータを保存する SHALL
9. WHEN 日本語データが含まれる THEN システムはUTF-8エンコーディングで出力する SHALL
10. WHEN pagePathディメンションが含まれる THEN システムはストリームURLとpagePathを結合してフルURLを出力する SHALL
11. WHEN ストリームURLの取得が必要な場合 THEN システムはGoogle Analytics Management APIを使用してストリームURLを取得する SHALL
12. WHEN JSON出力が実行される THEN 各レコードは以下の構造を持つ SHALL:
    - dimensions: ディメンション名と値のキー・バリューペア
    - metrics: メトリクス名と値のキー・バリューペア
    - metadata: 取得日時やプロパティ情報などのメタデータ

### Requirement 5: エラーハンドリングとログ

**User Story:** 開発者として、問題が発生した際に適切な診断情報を得るために、詳細なエラーメッセージとログ機能が欲しい。

#### Acceptance Criteria

1. WHEN エラーが発生する THEN システムは分かりやすいエラーメッセージを表示する SHALL
2. WHEN API制限に達する THEN システムは適切な待機時間を設けてリトライする SHALL
3. WHEN ネットワークエラーが発生する THEN システムは接続の再試行を行う SHALL
4. IF デバッグモードが有効になっている THEN システムは詳細なログ情報を出力する SHALL
5. WHEN 致命的なエラーが発生する THEN システムは適切な終了コード（非ゼロ）で終了する SHALL

### Requirement 6: URL結合機能

**User Story:** データアナリストとして、CSV出力でページの完全なURLを確認するために、ストリームのベースURLとpagePathを結合した形で表示したい。

#### Acceptance Criteria

1. WHEN ストリーム設定にbase_urlが指定されている THEN システムはpagePathとbase_urlを結合する SHALL
2. WHEN pagePathが相対パス（/で始まる）である THEN システムはbase_urlとpagePathを適切に結合する SHALL
3. WHEN pagePathが絶対URL（http://またはhttps://で始まる）である THEN システムはpagePathをそのまま使用する SHALL
4. WHEN base_urlが設定されていない THEN システムはpagePathをそのまま出力する SHALL
5. WHEN CSV出力が実行される THEN 結合されたフルURLが出力される SHALL
6. WHEN base_urlの末尾にスラッシュがある場合 THEN システムは重複スラッシュを適切に処理する SHALL
7. WHEN pagePathが空文字列またはnullの場合 THEN システムはbase_urlのみを出力する SHALL

### Requirement 7: コマンドラインインターフェース

**User Story:** ユーザーとして、直感的で使いやすいコマンドラインインターフェースを通じてツールを操作したい。

#### Acceptance Criteria

1. WHEN ユーザーが `ga --help` を実行する THEN システムは使用方法を表示する SHALL
2. WHEN ユーザーが `ga --version` を実行する THEN システムはバージョン情報を表示する SHALL
3. WHEN ユーザーが `ga` を実行する THEN システムは `ga.yaml` を使用してデータ取得を開始する SHALL
4. WHEN 無効なオプションが指定される THEN システムは適切なエラーメッセージとヘルプ情報を表示する SHALL
5. WHEN ユーザーが設定ファイルパスを指定する THEN システムは `--config` オプションでカスタム設定ファイルを受け入れる SHALL
6. WHEN ユーザーが出力形式を指定する THEN システムは `--format` オプションで `csv` または `json` を受け入れる SHALL
7. WHEN ユーザーが出力ファイルを指定する THEN システムは `--output` オプションでファイルパスを受け入れる SHALL
8. WHEN `--format` オプションが省略される THEN システムはデフォルトでCSV形式を使用する SHALL