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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ymotongpoo/ga/internal/analytics"
)

// OutputService はデータ出力を提供するインターフェース
type OutputService interface {
	// WriteCSV はReportDataをCSV形式でWriterに出力する
	WriteCSV(data *analytics.ReportData, writer io.Writer) error
	// WriteToFile はReportDataをCSV形式でファイルに出力する
	WriteToFile(data *analytics.ReportData, filename string) error
	// WriteToConsole はReportDataをCSV形式で標準出力に出力する
	WriteToConsole(data *analytics.ReportData) error
	// WriteOutput は出力先に応じて適切な出力方法を選択する
	WriteOutput(data *analytics.ReportData, outputPath string) error
}

// CSVWriter はCSV出力を行う構造体
type CSVWriter struct {
	encoding  string
	delimiter rune
}

// OutputServiceImpl はOutputServiceの実装
type OutputServiceImpl struct {
	csvWriter *CSVWriter
}

// NewOutputService は新しいOutputServiceを作成する
func NewOutputService() OutputService {
	return &OutputServiceImpl{
		csvWriter: &CSVWriter{
			encoding:  "UTF-8",
			delimiter: ',',
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

// WriteToFile はReportDataをCSV形式でファイルに出力する
func (o *OutputServiceImpl) WriteToFile(data *analytics.ReportData, filename string) error {
	return o.WriteToFileWithErrorHandling(data, filename)
}

// WriteToConsole はReportDataをCSV形式で標準出力に出力する
func (o *OutputServiceImpl) WriteToConsole(data *analytics.ReportData) error {
	// 標準出力への書き込み前にサマリー情報を標準エラー出力に表示
	fmt.Fprintf(os.Stderr, "📊 CSV出力を標準出力に書き込みます...\n")
	fmt.Fprintf(os.Stderr, "   - 総行数: %d行 (ヘッダー含む)\n", len(data.Rows)+1)
	fmt.Fprintf(os.Stderr, "   - 列数: %d列\n", len(data.Headers))
	fmt.Fprintf(os.Stderr, "   - 期間: %s\n", data.Summary.DateRange)
	fmt.Fprintf(os.Stderr, "\n")

	// CSV形式で標準出力に書き込み
	if err := o.WriteCSV(data, os.Stdout); err != nil {
		return fmt.Errorf("標準出力への書き込みに失敗しました: %w", err)
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

// WriteOutput は出力先に応じて適切な出力方法を選択する
func (o *OutputServiceImpl) WriteOutput(data *analytics.ReportData, outputPath string) error {
	// データの妥当性を検証
	if err := o.ValidateData(data); err != nil {
		return fmt.Errorf("出力データの検証に失敗しました: %w", err)
	}

	// 出力先が指定されていない場合は標準出力
	if outputPath == "" || outputPath == "-" {
		return o.WriteToConsole(data)
	}

	// ファイル出力の場合
	return o.WriteToFileWithErrorHandling(data, outputPath)
}

// WriteToFileWithErrorHandling はエラーハンドリングを強化したファイル出力
func (o *OutputServiceImpl) WriteToFileWithErrorHandling(data *analytics.ReportData, filename string) error {
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

	// CSV形式で書き込み
	if err := o.WriteCSV(data, file); err != nil {
		// 書き込みエラーの場合、部分的に作成されたファイルを削除
		if removeErr := os.Remove(filename); removeErr != nil {
			fmt.Fprintf(os.Stderr, "警告: 不完全なファイル '%s' の削除に失敗しました: %v\n", filename, removeErr)
		}
		return fmt.Errorf("ファイル '%s' への書き込みに失敗しました: %w", filename, err)
	}

	// 出力完了メッセージ
	fmt.Printf("📄 CSV出力が完了しました: %s\n", filename)
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

	// 拡張子のチェック（.csvを推奨）
	if !strings.HasSuffix(strings.ToLower(filename), ".csv") {
		fmt.Fprintf(os.Stderr, "⚠️  ファイル拡張子が .csv ではありません: %s\n", filename)
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
func (o *OutputServiceImpl) GetOutputSummary(data *analytics.ReportData) string {
	if data == nil {
		return "データなし"
	}

	summary := fmt.Sprintf("CSV出力サマリー:\n")
	summary += fmt.Sprintf("  - 総レコード数: %d\n", data.Summary.TotalRows)
	summary += fmt.Sprintf("  - 出力行数: %d行 (ヘッダー含む)\n", len(data.Rows)+1)
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