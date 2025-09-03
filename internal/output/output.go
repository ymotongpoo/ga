// Copyright 2025 Yoshi Yamaguchi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package output

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ymotongpoo/ga/internal/analytics"
	"github.com/ymotongpoo/ga/internal/url"
)

// OutputFormat は出力形式を表す列挙型
type OutputFormat int

const (
	FormatCSV OutputFormat = iota
	FormatJSON
)

// String はOutputFormatの文字列表現を返す
func (f OutputFormat) String() string {
	switch f {
	case FormatCSV:
		return "csv"
	case FormatJSON:
		return "json"
	default:
		return "unknown"
	}
}

// ParseOutputFormat は文字列から OutputFormat を解析する
// 要件4.2, 4.3, 4.4: デフォルト形式（CSV）の設定、無効な形式指定時のエラーハンドリング
func ParseOutputFormat(format string) (OutputFormat, error) {
	// 空文字列の場合はデフォルト（CSV）を返す
	if strings.TrimSpace(format) == "" {
		return FormatCSV, nil
	}

	// 大文字小文字を無視して比較
	normalizedFormat := strings.ToLower(strings.TrimSpace(format))

	switch normalizedFormat {
	case "csv":
		return FormatCSV, nil
	case "json":
		return FormatJSON, nil
	default:
		// サポートされている形式の一覧を含む詳細なエラーメッセージ
		return FormatCSV, fmt.Errorf("無効な出力形式: '%s'\n\nサポートされている形式:\n  - csv  (デフォルト)\n  - json\n\n例: --format csv または --format json", format)
	}
}

// GetDefaultOutputFormat はデフォルトの出力形式を返す
// 要件4.2: デフォルト形式（CSV）の設定
func GetDefaultOutputFormat() OutputFormat {
	return FormatCSV
}

// IsValidOutputFormat は指定された形式が有効かどうかを判定する
func IsValidOutputFormat(format string) bool {
	_, err := ParseOutputFormat(format)
	return err == nil
}

// GetSupportedFormats はサポートされている出力形式の一覧を返す
func GetSupportedFormats() []string {
	return []string{"csv", "json"}
}

// OutputService はデータ出力を提供するインターフェース
// 要件4.7, 4.8: 統合出力サービスの更新、形式対応の強化
type OutputService interface {
	// WriteCSV はReportDataをCSV形式でWriterに出力する
	WriteCSV(data *analytics.ReportData, writer io.Writer) error
	// WriteJSON はReportDataをJSON形式でWriterに出力する
	WriteJSON(data *analytics.ReportData, writer io.Writer) error
	// WriteToFile はReportDataを指定された形式でファイルに出力する
	WriteToFile(data *analytics.ReportData, filename string, format OutputFormat) error
	// WriteToConsole はReportDataを指定された形式で標準出力に出力する
	WriteToConsole(data *analytics.ReportData, format OutputFormat) error
	// WriteOutput は出力先と形式に応じて適切な出力方法を選択する
	WriteOutput(data *analytics.ReportData, outputPath string, format OutputFormat) error
	// WriteWithOptions は詳細なオプション付きで出力する
	WriteWithOptions(data *analytics.ReportData, options OutputOptions) error
	// ValidateOutputOptions は出力オプションの妥当性を検証する
	ValidateOutputOptions(options OutputOptions) error
	// GetOutputSummary は出力データのサマリー情報を取得する
	GetOutputSummary(data *analytics.ReportData, format OutputFormat) string
}

// CSVWriter はCSV出力を行う構造体
type CSVWriter struct {
	encoding  string
	delimiter rune
}

// JSONWriter はJSON出力を行う構造体
// 要件4.6, 4.9: 構造化されたJSON配列の生成、UTF-8エンコーディング対応
type JSONWriter struct {
	encoding      string
	indent        string
	escapeHTML    bool
	sortKeys      bool
	compactOutput bool
}

// JSONRecord はJSON出力用のレコード構造体
// 要件4.6, 4.12: ディメンションとメトリクスのキー・バリューペア、メタデータを含む
type JSONRecord struct {
	Dimensions map[string]string `json:"dimensions"`
	Metrics    map[string]string `json:"metrics"`
	Metadata   JSONMetadata      `json:"metadata"`
}

// JSONMetadata はJSON出力用のメタデータ構造体
// 要件4.12: 取得日時、プロパティ情報などのメタデータを含む
type JSONMetadata struct {
	RetrievedAt  string `json:"retrieved_at"`
	PropertyID   string `json:"property_id,omitempty"`
	StreamID     string `json:"stream_id,omitempty"`
	DateRange    string `json:"date_range"`
	RecordIndex  int    `json:"record_index"`
	TotalRecords int    `json:"total_records"`
	OutputFormat string `json:"output_format"`
	ToolVersion  string `json:"tool_version,omitempty"`
}

// OutputServiceImpl はOutputServiceの実装
type OutputServiceImpl struct {
	csvWriter  *CSVWriter
	jsonWriter *JSONWriter
}

// NewOutputService は新しいOutputServiceを作成する
func NewOutputService() OutputService {
	return &OutputServiceImpl{
		csvWriter: &CSVWriter{
			encoding:  "UTF-8",
			delimiter: ',',
		},
		jsonWriter: &JSONWriter{
			encoding:      "UTF-8",
			indent:        "  ",
			escapeHTML:    false,
			sortKeys:      false,
			compactOutput: false,
		},
	}
}

// WriteCSV はReportDataをCSV形式でWriterに出力する
func (o *OutputServiceImpl) WriteCSV(data *analytics.ReportData, writer io.Writer) error {
	if data == nil {
		return fmt.Errorf("出力データがnilです")
	}

	// CSVライターを作成
	csvWriter := csv.NewWriter(writer)
	csvWriter.Comma = o.csvWriter.delimiter
	defer csvWriter.Flush()

	// URL結合処理の準備
	urlProcessor := url.NewURLProcessor(data.StreamURLs)
	processedHeaders, pagePathIndex := o.processHeaders(data.Headers)

	// ヘッダー行を書き込み
	if len(processedHeaders) > 0 {
		if err := csvWriter.Write(processedHeaders); err != nil {
			return fmt.Errorf("ヘッダー行の書き込みに失敗しました: %w", err)
		}
	}

	// データ行を書き込み（URL結合処理付き）
	for i, row := range data.Rows {
		processedRow := o.processRow(row, pagePathIndex, urlProcessor, data.Headers)
		if err := csvWriter.Write(processedRow); err != nil {
			return fmt.Errorf("データ行 %d の書き込みに失敗しました: %w", i+1, err)
		}
	}

	// バッファをフラッシュしてエラーをチェック
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("CSV書き込み中にエラーが発生しました: %w", err)
	}

	return nil
}

// WriteJSON はReportDataをJSON形式でWriterに出力する
// 要件4.6, 4.9: 構造化されたJSON配列形式で出力、UTF-8エンコーディング対応
func (o *OutputServiceImpl) WriteJSON(data *analytics.ReportData, writer io.Writer) error {
	if data == nil {
		return fmt.Errorf("出力データがnilです")
	}

	// JSON レコードの配列を作成（空の場合でも配列として初期化）
	records := make([]JSONRecord, 0, len(data.Rows))

	// 現在時刻を取得
	retrievedAt := time.Now().UTC().Format(time.RFC3339)
	totalRecords := len(data.Rows)

	// URL結合処理の準備
	urlProcessor := url.NewURLProcessor(data.StreamURLs)

	// 各データ行をJSONレコードに変換
	for recordIndex, row := range data.Rows {
		if len(row) != len(data.Headers) {
			continue // 不正な行はスキップ
		}

		// URL結合処理を行った行データを作成
		processedRow := o.processRowForJSON(row, data.Headers, urlProcessor)

		// ディメンションとメトリクスのキー・バリューペアを作成
		dimensions, metrics := o.createKeyValuePairs(data.Headers, processedRow)

		// プロパティIDとストリームIDを抽出
		propertyID := o.extractPropertyID(processedRow, data.Headers)
		streamID := o.extractStreamID(processedRow, data.Headers)

		record := JSONRecord{
			Dimensions: dimensions,
			Metrics:    metrics,
			Metadata: JSONMetadata{
				RetrievedAt:  retrievedAt,
				PropertyID:   propertyID,
				StreamID:     streamID,
				DateRange:    data.Summary.DateRange,
				RecordIndex:  recordIndex + 1, // 1ベースのインデックス
				TotalRecords: totalRecords,
				OutputFormat: "json",
				ToolVersion:  "ga-tool-v1.0", // バージョン情報
			},
		}

		records = append(records, record)
	}

	// JSONライターを使用して出力
	return o.jsonWriter.writeRecords(records, writer)
}

// createKeyValuePairs はヘッダーと行データからディメンションとメトリクスのキー・バリューペアを作成する
// 要件4.6: ディメンションとメトリクスのキー・バリューペア変換
func (o *OutputServiceImpl) createKeyValuePairs(headers []string, row []string) (map[string]string, map[string]string) {
	dimensions := make(map[string]string)
	metrics := make(map[string]string)

	// ヘッダーと値をマッピング
	for i, header := range headers {
		value := ""
		if i < len(row) {
			value = row[i]
		}

		// pagePathの場合はfullURLとして扱う
		displayHeader := header
		if strings.ToLower(header) == "pagepath" {
			displayHeader = "fullURL"
		}

		// ディメンションとメトリクスを分類
		if isDimension(header) {
			dimensions[displayHeader] = value
		} else {
			metrics[displayHeader] = value
		}
	}

	return dimensions, metrics
}

// extractPropertyID は行データからプロパティIDを抽出する
func (o *OutputServiceImpl) extractPropertyID(row []string, headers []string) string {
	for i, header := range headers {
		if strings.ToLower(header) == "property_id" && i < len(row) {
			return row[i]
		}
	}
	return ""
}

// extractStreamID は行データからストリームIDを抽出する
func (o *OutputServiceImpl) extractStreamID(row []string, headers []string) string {
	for i, header := range headers {
		if strings.ToLower(header) == "stream_id" && i < len(row) {
			return row[i]
		}
	}
	return ""
}

// isDimension はヘッダー名がディメンションかどうかを判定する
// 要件4.6: ディメンションとメトリクスの正確な分類
func isDimension(header string) bool {
	headerLower := strings.ToLower(header)

	// 明確にメトリクスと判定できるもの
	knownMetrics := map[string]bool{
		"sessions":               true,
		"activeusers":            true,
		"newusers":               true,
		"averagesessionduration": true,
		"engagementrateduration": true,
		"bouncerate":             true,
		"pageviews":              true,
		"screenpageviews":        true,
		"eventcount":             true,
		"conversions":            true,
		"totalrevenue":           true,
		"engagementrate":         true,
		"engagedsessions":        true,
		"averageengagementtime":  true,
		"sessionsperpuser":       true,
		"eventsperuser":          true,
		"screenviewsperuser":     true,
		"totalusers":             true,
		"userengagementduration": true,
	}

	// スペースやアンダースコアを除去して正規化
	normalizedHeader := strings.ReplaceAll(strings.ReplaceAll(headerLower, "_", ""), " ", "")

	if knownMetrics[normalizedHeader] {
		return false
	}

	// 明確にディメンションと判定できるもの
	knownDimensions := map[string]bool{
		"date":            true,
		"pagepath":        true,
		"fullurl":         true,
		"country":         true,
		"city":            true,
		"browser":         true,
		"operatingsystem": true,
		"devicecategory":  true,
		"channelgrouping": true,
		"source":          true,
		"medium":          true,
		"campaign":        true,
		"landingpage":     true,
		"exitpage":        true,
		"eventname":       true,
		"propertyid":      true,
		"streamid":        true,
		"hostname":        true,
		"pagetitle":       true,
		"referrer":        true,
		"userid":          true,
		"sessionid":       true,
		"transactionid":   true,
		"itemid":          true,
		"itemname":        true,
		"itemcategory":    true,
		"continent":       true,
		"region":          true,
		"metro":           true,
		"language":        true,
		"age":             true,
		"gender":          true,
	}

	normalizedHeader = strings.ReplaceAll(strings.ReplaceAll(headerLower, "_", ""), " ", "")

	if knownDimensions[normalizedHeader] {
		return true
	}

	// 数値的な名前パターンをチェック（メトリクスの可能性が高い）
	if strings.Contains(headerLower, "count") ||
		strings.Contains(headerLower, "rate") ||
		strings.Contains(headerLower, "duration") ||
		strings.Contains(headerLower, "time") ||
		strings.Contains(headerLower, "revenue") ||
		strings.Contains(headerLower, "value") {
		return false
	}

	// デフォルトはディメンションとして扱う
	return true
}

// WriteToFile はReportDataを指定された形式でファイルに出力する
func (o *OutputServiceImpl) WriteToFile(data *analytics.ReportData, filename string, format OutputFormat) error {
	return o.WriteToFileWithErrorHandling(data, filename, format)
}

// WriteToConsole はReportDataを指定された形式で標準出力に出力する
func (o *OutputServiceImpl) WriteToConsole(data *analytics.ReportData, format OutputFormat) error {
	// 標準出力への書き込み前にサマリー情報を標準エラー出力に表示
	formatName := format.String()
	fmt.Fprintf(os.Stderr, "📊 %s出力を標準出力に書き込みます...\n", strings.ToUpper(formatName))
	fmt.Fprintf(os.Stderr, "   - 総行数: %d行", len(data.Rows))
	if format == FormatCSV {
		fmt.Fprintf(os.Stderr, " (ヘッダー含む)")
	}
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "   - 列数: %d列\n", len(data.Headers))
	fmt.Fprintf(os.Stderr, "   - 期間: %s\n", data.Summary.DateRange)
	fmt.Fprintf(os.Stderr, "\n")

	// 指定された形式で標準出力に書き込み
	switch format {
	case FormatCSV:
		if err := o.WriteCSV(data, os.Stdout); err != nil {
			return fmt.Errorf("CSV標準出力への書き込みに失敗しました: %w", err)
		}
	case FormatJSON:
		if err := o.WriteJSON(data, os.Stdout); err != nil {
			return fmt.Errorf("JSON標準出力への書き込みに失敗しました: %w", err)
		}
	default:
		return fmt.Errorf("サポートされていない出力形式です: %s", format)
	}

	return nil
}

// ValidateData はReportDataの妥当性を検証する
func (o *OutputServiceImpl) ValidateData(data *analytics.ReportData) error {
	if data == nil {
		return fmt.Errorf("データがnilです")
	}

	if len(data.Headers) == 0 {
		return fmt.Errorf("ヘッダーが空です")
	}

	// 各行の列数がヘッダーと一致するかチェック
	expectedColumns := len(data.Headers)
	for i, row := range data.Rows {
		if len(row) != expectedColumns {
			return fmt.Errorf("行 %d の列数が不正です: 期待値=%d, 実際=%d", i+1, expectedColumns, len(row))
		}
	}

	return nil
}

// WriteOutput は出力先と形式に応じて適切な出力方法を選択する
func (o *OutputServiceImpl) WriteOutput(data *analytics.ReportData, outputPath string, format OutputFormat) error {
	// データの妥当性を検証
	if err := o.ValidateData(data); err != nil {
		return fmt.Errorf("出力データの検証に失敗しました: %w", err)
	}

	// 出力先が指定されていない場合は標準出力
	if outputPath == "" || outputPath == "-" {
		return o.WriteToConsole(data, format)
	}

	// ファイル出力の場合
	return o.WriteToFileWithErrorHandling(data, outputPath, format)
}

// WriteToFileWithErrorHandling はエラーハンドリングを強化したファイル出力
func (o *OutputServiceImpl) WriteToFileWithErrorHandling(data *analytics.ReportData, filename string, format OutputFormat) error {
	if filename == "" {
		return fmt.Errorf("ファイル名が指定されていません")
	}

	// ファイルパスの妥当性をチェック
	if err := o.validateFilePath(filename); err != nil {
		return fmt.Errorf("ファイルパス '%s' が無効です: %w", filename, err)
	}

	// ファイルが既に存在する場合の確認（上書き警告）
	if _, err := os.Stat(filename); err == nil {
		fmt.Fprintf(os.Stderr, "⚠️  ファイル '%s' は既に存在します。上書きします。\n", filename)
	}

	// ディレクトリが存在しない場合は作成を試行
	if err := o.ensureDirectoryExists(filename); err != nil {
		return fmt.Errorf("出力ディレクトリの作成に失敗しました: %w", err)
	}

	// ファイルを作成（存在する場合は上書き）
	file, err := os.Create(filename)
	if err != nil {
		return o.handleFileCreationError(filename, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "警告: ファイル '%s' のクローズに失敗しました: %v\n", filename, closeErr)
		}
	}()

	// 指定された形式で書き込み
	var writeErr error
	switch format {
	case FormatCSV:
		writeErr = o.WriteCSV(data, file)
	case FormatJSON:
		writeErr = o.WriteJSON(data, file)
	default:
		writeErr = fmt.Errorf("サポートされていない出力形式です: %s", format)
	}

	if writeErr != nil {
		// 書き込みエラーの場合、部分的に作成されたファイルを削除
		if removeErr := os.Remove(filename); removeErr != nil {
			fmt.Fprintf(os.Stderr, "警告: 不完全なファイル '%s' の削除に失敗しました: %v\n", filename, removeErr)
		}
		return fmt.Errorf("ファイル '%s' への書き込みに失敗しました: %w", filename, writeErr)
	}

	// 出力完了メッセージ
	formatName := strings.ToUpper(format.String())
	fmt.Printf("📄 %s出力が完了しました: %s\n", formatName, filename)
	fmt.Printf("   - 総行数: %d行 (ヘッダー含む)\n", len(data.Rows)+1)
	fmt.Printf("   - 列数: %d列\n", len(data.Headers))
	fmt.Printf("   - ファイルサイズ: ")

	// ファイルサイズを取得して表示
	if fileInfo, err := file.Stat(); err == nil {
		fmt.Printf("%.2f KB\n", float64(fileInfo.Size())/1024)
	} else {
		fmt.Printf("不明\n")
	}

	return nil
}

// validateFilePath はファイルパスの妥当性を検証する
func (o *OutputServiceImpl) validateFilePath(filename string) error {
	// 空文字チェック
	if strings.TrimSpace(filename) == "" {
		return fmt.Errorf("ファイル名が空です")
	}

	// 危険な文字のチェック
	dangerousChars := []string{"\x00", "\n", "\r"}
	for _, char := range dangerousChars {
		if strings.Contains(filename, char) {
			return fmt.Errorf("ファイル名に無効な文字が含まれています")
		}
	}

	// 拡張子のチェック（.csv または .json を推奨）
	lowerFilename := strings.ToLower(filename)
	if !strings.HasSuffix(lowerFilename, ".csv") && !strings.HasSuffix(lowerFilename, ".json") {
		fmt.Fprintf(os.Stderr, "⚠️  ファイル拡張子が .csv または .json ではありません: %s\n", filename)
	}

	return nil
}

// ensureDirectoryExists は出力ディレクトリが存在することを確認し、必要に応じて作成する
func (o *OutputServiceImpl) ensureDirectoryExists(filename string) error {
	dir := filepath.Dir(filename)

	// カレントディレクトリの場合は何もしない
	if dir == "." {
		return nil
	}

	// ディレクトリが存在するかチェック
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "📁 ディレクトリを作成します: %s\n", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("ディレクトリ '%s' の作成に失敗しました: %w", dir, err)
		}
	}

	return nil
}

// handleFileCreationError はファイル作成エラーを詳細に処理する
func (o *OutputServiceImpl) handleFileCreationError(filename string, err error) error {
	// 権限エラーの場合
	if os.IsPermission(err) {
		return fmt.Errorf("ファイル '%s' への書き込み権限がありません: %w", filename, err)
	}

	// ディスク容量不足の場合
	if strings.Contains(err.Error(), "no space left on device") {
		return fmt.Errorf("ディスク容量が不足しています。ファイル '%s' を作成できません: %w", filename, err)
	}

	// ファイル名が長すぎる場合
	if strings.Contains(err.Error(), "file name too long") {
		return fmt.Errorf("ファイル名が長すぎます: %s", filename)
	}

	// その他のエラー
	return fmt.Errorf("ファイル '%s' の作成に失敗しました: %w", filename, err)
}

// writeRecords はJSONレコード配列をWriterに出力する
// 要件4.6, 4.9: 構造化されたJSON配列の生成、UTF-8エンコーディング対応
func (jw *JSONWriter) writeRecords(records []JSONRecord, writer io.Writer) error {
	// JSON エンコーダーを作成（UTF-8エンコーディング）
	encoder := json.NewEncoder(writer)

	// エンコーダーの設定
	if jw.compactOutput {
		encoder.SetIndent("", "")
	} else {
		encoder.SetIndent("", jw.indent)
	}

	encoder.SetEscapeHTML(jw.escapeHTML)

	// JSON配列として出力
	if err := encoder.Encode(records); err != nil {
		return fmt.Errorf("JSON書き込み中にエラーが発生しました: %w", err)
	}

	return nil
}

// writeRecordsWithOptions はオプション付きでJSONレコード配列を出力する
func (jw *JSONWriter) writeRecordsWithOptions(records []JSONRecord, writer io.Writer, options JSONWriteOptions) error {
	// 一時的にオプションを適用
	originalIndent := jw.indent
	originalEscapeHTML := jw.escapeHTML
	originalCompactOutput := jw.compactOutput

	if options.Indent != nil {
		jw.indent = *options.Indent
	}
	if options.EscapeHTML != nil {
		jw.escapeHTML = *options.EscapeHTML
	}
	if options.CompactOutput != nil {
		jw.compactOutput = *options.CompactOutput
	}

	// 出力実行
	err := jw.writeRecords(records, writer)

	// 設定を元に戻す
	jw.indent = originalIndent
	jw.escapeHTML = originalEscapeHTML
	jw.compactOutput = originalCompactOutput

	return err
}

// JSONWriteOptions はJSON出力のオプションを定義する
type JSONWriteOptions struct {
	Indent        *string
	EscapeHTML    *bool
	CompactOutput *bool
	SortKeys      *bool
}

// OutputOptions は統合出力オプションを定義する
// 要件4.7, 4.8: 統合出力サービスの更新
type OutputOptions struct {
	// 基本オプション
	OutputPath string
	Format     OutputFormat

	// ファイル出力オプション
	OverwriteExisting bool
	CreateDirectories bool
	FilePermissions   os.FileMode

	// 表示オプション
	ShowProgress bool
	ShowSummary  bool
	QuietMode    bool

	// 形式固有オプション
	CSVOptions  *CSVWriteOptions
	JSONOptions *JSONWriteOptions
}

// CSVWriteOptions はCSV出力のオプションを定義する
type CSVWriteOptions struct {
	Delimiter     rune
	IncludeHeader bool
	Encoding      string
}

// validateJSONOutput は出力されたJSONの妥当性を検証する
func (jw *JSONWriter) validateJSONOutput(data []byte) error {
	var records []JSONRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("出力されたJSONが無効です: %w", err)
	}

	// 基本的な構造の検証
	for i, record := range records {
		if record.Dimensions == nil {
			return fmt.Errorf("レコード %d の dimensions が nil です", i+1)
		}
		if record.Metrics == nil {
			return fmt.Errorf("レコード %d の metrics が nil です", i+1)
		}
		if record.Metadata.RetrievedAt == "" {
			return fmt.Errorf("レコード %d の retrieved_at が空です", i+1)
		}
	}

	return nil
}

// formatJSONForDisplay は表示用にJSONを整形する
func (jw *JSONWriter) formatJSONForDisplay(records []JSONRecord) (string, error) {
	var buf bytes.Buffer

	// 表示用の設定で出力
	options := JSONWriteOptions{
		Indent:        stringPtr("  "),
		EscapeHTML:    boolPtr(false),
		CompactOutput: boolPtr(false),
	}

	if err := jw.writeRecordsWithOptions(records, &buf, options); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// stringPtr は文字列のポインタを返すヘルパー関数
func stringPtr(s string) *string {
	return &s
}

// boolPtr はboolのポインタを返すヘルパー関数
func boolPtr(b bool) *bool {
	return &b
}

// WriteWithOptions は詳細なオプション付きで出力する
// 要件4.7, 4.8: 統合出力サービスの更新、形式対応
func (o *OutputServiceImpl) WriteWithOptions(data *analytics.ReportData, options OutputOptions) error {
	// オプションの妥当性を検証
	if err := o.ValidateOutputOptions(options); err != nil {
		return fmt.Errorf("出力オプションの検証に失敗しました: %w", err)
	}

	// データの妥当性を検証
	if err := o.ValidateData(data); err != nil {
		return fmt.Errorf("出力データの検証に失敗しました: %w", err)
	}

	// プログレス表示
	if options.ShowProgress && !options.QuietMode {
		formatName := strings.ToUpper(options.Format.String())
		if options.OutputPath == "" || options.OutputPath == "-" {
			fmt.Fprintf(os.Stderr, "📊 %s出力を標準出力に書き込み中...\n", formatName)
		} else {
			fmt.Fprintf(os.Stderr, "📄 %s出力をファイル '%s' に書き込み中...\n", formatName, options.OutputPath)
		}
	}

	// 出力先に応じて処理を分岐
	if options.OutputPath == "" || options.OutputPath == "-" {
		return o.writeToConsoleWithOptions(data, options)
	} else {
		return o.writeToFileWithOptions(data, options)
	}
}

// ValidateOutputOptions は出力オプションの妥当性を検証する
// 要件4.7, 4.8: 統合出力サービスの更新
func (o *OutputServiceImpl) ValidateOutputOptions(options OutputOptions) error {
	// 出力形式の検証
	if options.Format != FormatCSV && options.Format != FormatJSON {
		return fmt.Errorf("サポートされていない出力形式です: %v", options.Format)
	}

	// ファイルパスの検証（ファイル出力の場合）
	if options.OutputPath != "" && options.OutputPath != "-" {
		if err := o.validateFilePath(options.OutputPath); err != nil {
			return fmt.Errorf("出力ファイルパスが無効です: %w", err)
		}
	}

	// ファイル権限の検証
	if options.FilePermissions != 0 && (options.FilePermissions < 0o400 || options.FilePermissions > 0o777) {
		return fmt.Errorf("無効なファイル権限です: %o", options.FilePermissions)
	}

	// CSV固有オプションの検証
	if options.Format == FormatCSV && options.CSVOptions != nil {
		if options.CSVOptions.Delimiter == 0 {
			return fmt.Errorf("CSVデリミタが指定されていません")
		}
		if options.CSVOptions.Encoding != "" && options.CSVOptions.Encoding != "UTF-8" {
			return fmt.Errorf("サポートされていないエンコーディングです: %s", options.CSVOptions.Encoding)
		}
	}

	return nil
}

// writeToConsoleWithOptions はオプション付きで標準出力に書き込む
func (o *OutputServiceImpl) writeToConsoleWithOptions(data *analytics.ReportData, options OutputOptions) error {
	// サマリー表示
	if options.ShowSummary && !options.QuietMode {
		summary := o.GetOutputSummary(data, options.Format)
		fmt.Fprintf(os.Stderr, "%s\n", summary)
	}

	// 形式に応じて出力
	switch options.Format {
	case FormatCSV:
		if options.CSVOptions != nil {
			return o.writeCSVWithOptions(data, os.Stdout, *options.CSVOptions)
		}
		return o.WriteCSV(data, os.Stdout)
	case FormatJSON:
		if options.JSONOptions != nil {
			return o.jsonWriter.writeRecordsWithOptions(o.convertToJSONRecords(data), os.Stdout, *options.JSONOptions)
		}
		return o.WriteJSON(data, os.Stdout)
	default:
		return fmt.Errorf("サポートされていない出力形式です: %v", options.Format)
	}
}

// writeToFileWithOptions はオプション付きでファイルに書き込む
func (o *OutputServiceImpl) writeToFileWithOptions(data *analytics.ReportData, options OutputOptions) error {
	// ディレクトリ作成
	if options.CreateDirectories {
		if err := o.ensureDirectoryExists(options.OutputPath); err != nil {
			return fmt.Errorf("出力ディレクトリの作成に失敗しました: %w", err)
		}
	}

	// ファイル存在確認
	if !options.OverwriteExisting {
		if _, err := os.Stat(options.OutputPath); err == nil {
			return fmt.Errorf("ファイル '%s' は既に存在します。上書きするには OverwriteExisting オプションを有効にしてください", options.OutputPath)
		}
	}

	// ファイル作成
	file, err := os.Create(options.OutputPath)
	if err != nil {
		return o.handleFileCreationError(options.OutputPath, err)
	}
	defer file.Close()

	// ファイル権限設定
	if options.FilePermissions != 0 {
		if err := file.Chmod(options.FilePermissions); err != nil {
			fmt.Fprintf(os.Stderr, "警告: ファイル権限の設定に失敗しました: %v\n", err)
		}
	}

	// 形式に応じて書き込み
	var writeErr error
	switch options.Format {
	case FormatCSV:
		if options.CSVOptions != nil {
			writeErr = o.writeCSVWithOptions(data, file, *options.CSVOptions)
		} else {
			writeErr = o.WriteCSV(data, file)
		}
	case FormatJSON:
		if options.JSONOptions != nil {
			writeErr = o.jsonWriter.writeRecordsWithOptions(o.convertToJSONRecords(data), file, *options.JSONOptions)
		} else {
			writeErr = o.WriteJSON(data, file)
		}
	default:
		writeErr = fmt.Errorf("サポートされていない出力形式です: %v", options.Format)
	}

	if writeErr != nil {
		// エラー時にファイルを削除
		if removeErr := os.Remove(options.OutputPath); removeErr != nil {
			fmt.Fprintf(os.Stderr, "警告: 不完全なファイル '%s' の削除に失敗しました: %v\n", options.OutputPath, removeErr)
		}
		return fmt.Errorf("ファイル '%s' への書き込みに失敗しました: %w", options.OutputPath, writeErr)
	}

	// 完了メッセージ
	if !options.QuietMode {
		formatName := strings.ToUpper(options.Format.String())
		fmt.Printf("✅ %s出力が完了しました: %s\n", formatName, options.OutputPath)

		if options.ShowSummary {
			summary := o.GetOutputSummary(data, options.Format)
			fmt.Printf("%s\n", summary)
		}
	}

	return nil
}

// writeCSVWithOptions はオプション付きでCSVを書き込む
func (o *OutputServiceImpl) writeCSVWithOptions(data *analytics.ReportData, writer io.Writer, options CSVWriteOptions) error {
	csvWriter := csv.NewWriter(writer)
	if options.Delimiter != 0 {
		csvWriter.Comma = options.Delimiter
	}
	defer csvWriter.Flush()

	// ヘッダー書き込み
	if options.IncludeHeader && len(data.Headers) > 0 {
		if err := csvWriter.Write(data.Headers); err != nil {
			return fmt.Errorf("ヘッダー行の書き込みに失敗しました: %w", err)
		}
	}

	// データ行書き込み
	for i, row := range data.Rows {
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("データ行 %d の書き込みに失敗しました: %w", i+1, err)
		}
	}

	return csvWriter.Error()
}

// convertToJSONRecords はReportDataをJSONRecords配列に変換する
func (o *OutputServiceImpl) convertToJSONRecords(data *analytics.ReportData) []JSONRecord {
	records := make([]JSONRecord, 0, len(data.Rows))
	retrievedAt := time.Now().UTC().Format(time.RFC3339)
	totalRecords := len(data.Rows)

	for recordIndex, row := range data.Rows {
		if len(row) != len(data.Headers) {
			continue
		}

		dimensions, metrics := o.createKeyValuePairs(data.Headers, row)
		propertyID := o.extractPropertyID(row, data.Headers)
		streamID := o.extractStreamID(row, data.Headers)

		record := JSONRecord{
			Dimensions: dimensions,
			Metrics:    metrics,
			Metadata: JSONMetadata{
				RetrievedAt:  retrievedAt,
				PropertyID:   propertyID,
				StreamID:     streamID,
				DateRange:    data.Summary.DateRange,
				RecordIndex:  recordIndex + 1,
				TotalRecords: totalRecords,
				OutputFormat: "json",
				ToolVersion:  "ga-tool-v1.0",
			},
		}

		records = append(records, record)
	}

	return records
}

// GetOutputSummary は出力データのサマリー情報を取得する
func (o *OutputServiceImpl) GetOutputSummary(data *analytics.ReportData, format OutputFormat) string {
	if data == nil {
		return "データなし"
	}

	formatName := strings.ToUpper(format.String())
	summary := fmt.Sprintf("%s出力サマリー:\n", formatName)
	summary += fmt.Sprintf("  - 総レコード数: %d\n", data.Summary.TotalRows)

	if format == FormatCSV {
		summary += fmt.Sprintf("  - 出力行数: %d行 (ヘッダー含む)\n", len(data.Rows)+1)
	} else {
		summary += fmt.Sprintf("  - 出力レコード数: %d\n", len(data.Rows))
	}

	summary += fmt.Sprintf("  - 列数: %d列\n", len(data.Headers))
	summary += fmt.Sprintf("  - 期間: %s\n", data.Summary.DateRange)

	if len(data.Summary.Properties) > 0 {
		summary += fmt.Sprintf("  - プロパティ数: %d\n", len(data.Summary.Properties))
		if len(data.Summary.Properties) <= 3 {
			summary += fmt.Sprintf("  - プロパティ: %v\n", data.Summary.Properties)
		}
	}

	return summary
}

// processHeaders はヘッダーを処理してpagePathをfullURLに変更し、pagePathのインデックスを返す
func (o *OutputServiceImpl) processHeaders(headers []string) ([]string, int) {
	processedHeaders := make([]string, len(headers))
	pagePathIndex := -1

	for i, header := range headers {
		if strings.ToLower(header) == "pagepath" {
			processedHeaders[i] = "fullURL"
			pagePathIndex = i // 最後に見つかったインデックスを保持
		} else {
			processedHeaders[i] = header
		}
	}

	return processedHeaders, pagePathIndex
}

// processRow はデータ行を処理してURL結合を行う
func (o *OutputServiceImpl) processRow(row []string, pagePathIndex int, urlProcessor *url.URLProcessor, headers []string) []string {
	if pagePathIndex == -1 || pagePathIndex >= len(row) {
		// pagePathが見つからない場合はそのまま返す
		return row
	}

	processedRow := make([]string, len(row))
	copy(processedRow, row)

	// ストリームIDを取得
	streamID := o.extractStreamIDFromRow(row, headers)

	// pagePathとベースURLを結合
	pagePath := row[pagePathIndex]
	fullURL := urlProcessor.ProcessPagePath(streamID, pagePath)
	processedRow[pagePathIndex] = fullURL

	return processedRow
}

// extractStreamIDFromRow は行データからストリームIDを抽出する
func (o *OutputServiceImpl) extractStreamIDFromRow(row []string, headers []string) string {
	for i, header := range headers {
		if strings.ToLower(header) == "stream_id" && i < len(row) {
			return row[i]
		}
	}
	return ""
}

// processRowForJSON はJSON出力用にデータ行を処理してURL結合を行う
func (o *OutputServiceImpl) processRowForJSON(row []string, headers []string, urlProcessor *url.URLProcessor) []string {
	processedRow := make([]string, len(row))
	copy(processedRow, row)

	// pagePathのインデックスを探す
	pagePathIndex := -1
	for i, header := range headers {
		if strings.ToLower(header) == "pagepath" {
			pagePathIndex = i
			break
		}
	}

	if pagePathIndex == -1 || pagePathIndex >= len(row) {
		// pagePathが見つからない場合はそのまま返す
		return processedRow
	}

	// ストリームIDを取得
	streamID := o.extractStreamIDFromRow(row, headers)

	// pagePathとベースURLを結合
	pagePath := row[pagePathIndex]
	fullURL := urlProcessor.ProcessPagePath(streamID, pagePath)
	processedRow[pagePathIndex] = fullURL

	return processedRow
}
