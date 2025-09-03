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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ymotongpoo/ga/internal/analytics"
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
func ParseOutputFormat(format string) (OutputFormat, error) {
	switch strings.ToLower(format) {
	case "csv":
		return FormatCSV, nil
	case "json":
		return FormatJSON, nil
	default:
		return FormatCSV, fmt.Errorf("無効な出力形式: %s (csv または json を指定してください)", format)
	}
}

// OutputService はデータ出力を提供するインターフェース
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
}

// CSVWriter はCSV出力を行う構造体
type CSVWriter struct {
	encoding  string
	delimiter rune
}

// JSONWriter はJSON出力を行う構造体
type JSONWriter struct {
	encoding string
	indent   string
}

// JSONRecord はJSON出力用のレコード構造体
type JSONRecord struct {
	Dimensions map[string]string `json:"dimensions"`
	Metrics    map[string]string `json:"metrics"`
	Metadata   JSONMetadata      `json:"metadata"`
}

// JSONMetadata はJSON出力用のメタデータ構造体
type JSONMetadata struct {
	RetrievedAt string `json:"retrieved_at"`
	PropertyID  string `json:"property_id,omitempty"`
	StreamID    string `json:"stream_id,omitempty"`
	DateRange   string `json:"date_range"`
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
			encoding: "UTF-8",
			indent:   "  ",
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

	// ヘッダー行を書き込み
	if len(data.Headers) > 0 {
		if err := csvWriter.Write(data.Headers); err != nil {
			return fmt.Errorf("ヘッダー行の書き込みに失敗しました: %w", err)
		}
	}

	// データ行を書き込み
	for i, row := range data.Rows {
		if err := csvWriter.Write(row); err != nil {
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
func (o *OutputServiceImpl) WriteJSON(data *analytics.ReportData, writer io.Writer) error {
	if data == nil {
		return fmt.Errorf("出力データがnilです")
	}

	// JSON レコードの配列を作成
	var records []JSONRecord

	// 現在時刻を取得
	retrievedAt := time.Now().UTC().Format(time.RFC3339)

	// 各データ行をJSONレコードに変換
	for _, row := range data.Rows {
		if len(row) != len(data.Headers) {
			continue // 不正な行はスキップ
		}

		record := JSONRecord{
			Dimensions: make(map[string]string),
			Metrics:    make(map[string]string),
			Metadata: JSONMetadata{
				RetrievedAt: retrievedAt,
				DateRange:   data.Summary.DateRange,
			},
		}

		// ヘッダーと値をマッピング
		for i, header := range data.Headers {
			value := ""
			if i < len(row) {
				value = row[i]
			}

			// ディメンションとメトリクスを分類
			// 一般的なディメンション名をチェック
			if isDimension(header) {
				record.Dimensions[header] = value
			} else {
				record.Metrics[header] = value
			}
		}

		records = append(records, record)
	}

	// JSON エンコーダーを作成
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", o.jsonWriter.indent)

	// JSON配列として出力
	if err := encoder.Encode(records); err != nil {
		return fmt.Errorf("JSON書き込み中にエラーが発生しました: %w", err)
	}

	return nil
}

// isDimension はヘッダー名がディメンションかどうかを判定する
func isDimension(header string) bool {
	// 一般的なディメンション名のリスト
	dimensions := []string{
		"date", "pagePath", "fullURL", "country", "city", "browser",
		"operatingSystem", "deviceCategory", "channelGrouping", "source",
		"medium", "campaign", "landingPage", "exitPage", "eventName",
	}

	headerLower := strings.ToLower(header)
	for _, dim := range dimensions {
		if strings.ToLower(dim) == headerLower {
			return true
		}
	}

	// メトリクス名の場合はfalseを返す
	metrics := []string{
		"sessions", "activeUsers", "newUsers", "averageSessionDuration",
		"engagementRateDuration", "bounceRate", "pageviews", "screenPageViews",
		"eventCount", "conversions", "totalRevenue",
	}

	for _, metric := range metrics {
		if strings.ToLower(metric) == headerLower {
			return false
		}
	}

	// 不明な場合はディメンションとして扱う
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