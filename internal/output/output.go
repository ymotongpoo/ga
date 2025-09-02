package output

import (
	"io"

	"github.com/ymotongpoo/ga/internal/analytics"
)

// OutputService はデータ出力機能を提供するインターフェース
type OutputService interface {
	// WriteCSV はReportDataをCSV形式でWriterに出力する
	WriteCSV(data *analytics.ReportData, writer io.Writer) error

	// WriteToFile はReportDataを指定されたファイルに出力する
	WriteToFile(data *analytics.ReportData, filename string) error

	// WriteToConsole はReportDataを標準出力に出力する
	WriteToConsole(data *analytics.ReportData) error
}

// CSVWriter はCSV出力機能を提供する構造体
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
			encoding:  "utf-8",
			delimiter: ',',
		},
	}
}

// WriteCSV はReportDataをCSV形式でWriterに出力する
func (o *OutputServiceImpl) WriteCSV(data *analytics.ReportData, writer io.Writer) error {
	// TODO: 実装予定
	return nil
}

// WriteToFile はReportDataを指定されたファイルに出力する
func (o *OutputServiceImpl) WriteToFile(data *analytics.ReportData, filename string) error {
	// TODO: 実装予定
	return nil
}

// WriteToConsole はReportDataを標準出力に出力する
func (o *OutputServiceImpl) WriteToConsole(data *analytics.ReportData) error {
	// TODO: 実装予定
	return nil
}