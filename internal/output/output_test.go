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
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ymotongpoo/ga/internal/analytics"
)

// テスト用のサンプルデータを作成
func createTestReportData() *analytics.ReportData {
	return &analytics.ReportData{
		Headers: []string{"property_id", "date", "pagePath", "sessions", "activeUsers", "newUsers", "averageSessionDuration"},
		Rows: [][]string{
			{"123456789", "2023-01-01", "/home", "1250", "1100", "850", "120.5"},
			{"123456789", "2023-01-01", "/about", "450", "420", "380", "95.2"},
			{"123456789", "2023-01-02", "/home", "1180", "1050", "780", "115.8"},
			{"123456789", "2023-01-02", "/contact", "320", "300", "250", "88.4"},
		},
		Summary: analytics.ReportSummary{
			TotalRows:  4,
			DateRange:  "2023-01-01 - 2023-01-02",
			Properties: []string{"123456789"},
		},
	}
}

// 空のデータを作成
func createEmptyReportData() *analytics.ReportData {
	return &analytics.ReportData{
		Headers: []string{"property_id", "date", "sessions"},
		Rows:    [][]string{},
		Summary: analytics.ReportSummary{
			TotalRows:  0,
			DateRange:  "2023-01-01 - 2023-01-01",
			Properties: []string{},
		},
	}
}

// 不正なデータを作成（列数が一致しない）
func createInvalidReportData() *analytics.ReportData {
	return &analytics.ReportData{
		Headers: []string{"property_id", "date", "sessions"},
		Rows: [][]string{
			{"123456789", "2023-01-01", "1250"},          // 正常
			{"123456789", "2023-01-02"},                  // 列数不足
			{"123456789", "2023-01-03", "1180", "extra"}, // 列数過多
		},
		Summary: analytics.ReportSummary{
			TotalRows:  3,
			DateRange:  "2023-01-01 - 2023-01-03",
			Properties: []string{"123456789"},
		},
	}
}

func TestNewOutputService(t *testing.T) {
	service := NewOutputService()
	if service == nil {
		t.Fatal("NewOutputService() returned nil")
	}

	// 型アサーションでOutputServiceImplかチェック
	impl, ok := service.(*OutputServiceImpl)
	if !ok {
		t.Fatal("NewOutputService() did not return *OutputServiceImpl")
	}

	if impl.csvWriter == nil {
		t.Fatal("csvWriter is nil")
	}

	if impl.csvWriter.encoding != "UTF-8" {
		t.Errorf("Expected encoding UTF-8, got %s", impl.csvWriter.encoding)
	}

	if impl.csvWriter.delimiter != ',' {
		t.Errorf("Expected delimiter ',', got %c", impl.csvWriter.delimiter)
	}
}

func TestWriteCSV(t *testing.T) {
	service := NewOutputService()
	data := createTestReportData()

	var buf bytes.Buffer
	err := service.WriteCSV(data, &buf)
	if err != nil {
		t.Fatalf("WriteCSV() failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// ヘッダー行をチェック
	expectedHeader := "property_id,date,pagePath,sessions,activeUsers,newUsers,averageSessionDuration"
	if lines[0] != expectedHeader {
		t.Errorf("Expected header: %s, got: %s", expectedHeader, lines[0])
	}

	// データ行数をチェック
	expectedLines := len(data.Rows) + 1 // ヘッダー + データ行
	if len(lines) != expectedLines {
		t.Errorf("Expected %d lines, got %d", expectedLines, len(lines))
	}

	// 最初のデータ行をチェック
	expectedFirstRow := "123456789,2023-01-01,/home,1250,1100,850,120.5"
	if lines[1] != expectedFirstRow {
		t.Errorf("Expected first data row: %s, got: %s", expectedFirstRow, lines[1])
	}
}

func TestWriteCSV_EmptyData(t *testing.T) {
	service := NewOutputService()
	data := createEmptyReportData()

	var buf bytes.Buffer
	err := service.WriteCSV(data, &buf)
	if err != nil {
		t.Fatalf("WriteCSV() with empty data failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// ヘッダーのみが出力されることを確認
	if len(lines) != 1 {
		t.Errorf("Expected 1 line (header only), got %d", len(lines))
	}

	expectedHeader := "property_id,date,sessions"
	if lines[0] != expectedHeader {
		t.Errorf("Expected header: %s, got: %s", expectedHeader, lines[0])
	}
}

func TestWriteCSV_NilData(t *testing.T) {
	service := NewOutputService()

	var buf bytes.Buffer
	err := service.WriteCSV(nil, &buf)
	if err == nil {
		t.Fatal("WriteCSV() with nil data should return error")
	}

	expectedError := "出力データがnilです"
	if err.Error() != expectedError {
		t.Errorf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestWriteToFile(t *testing.T) {
	service := NewOutputService()
	data := createTestReportData()

	// 一時ファイル名を生成
	tempFile := "test_output.csv"
	defer os.Remove(tempFile) // テスト後にファイルを削除

	err := service.WriteToFile(data, tempFile, FormatCSV)
	if err != nil {
		t.Fatalf("WriteToFile() failed: %v", err)
	}

	// ファイルが作成されたことを確認
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// ファイル内容を読み込んで検証
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	expectedLines := len(data.Rows) + 1 // ヘッダー + データ行
	if len(lines) != expectedLines {
		t.Errorf("Expected %d lines in file, got %d", expectedLines, len(lines))
	}

	// ヘッダー行をチェック
	expectedHeader := "property_id,date,pagePath,sessions,activeUsers,newUsers,averageSessionDuration"
	if lines[0] != expectedHeader {
		t.Errorf("Expected header in file: %s, got: %s", expectedHeader, lines[0])
	}
}

func TestWriteToFile_EmptyFilename(t *testing.T) {
	service := NewOutputService()
	data := createTestReportData()

	err := service.WriteToFile(data, "", FormatCSV)
	if err == nil {
		t.Fatal("WriteToFile() with empty filename should return error")
	}

	expectedError := "ファイル名が指定されていません"
	if err.Error() != expectedError {
		t.Errorf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestWriteToFile_InvalidPath(t *testing.T) {
	service := NewOutputService()
	data := createTestReportData()

	// 存在しないディレクトリへの書き込みを試行
	invalidPath := "/nonexistent/directory/output.csv"
	err := service.WriteToFile(data, invalidPath, FormatCSV)
	if err == nil {
		t.Fatal("WriteToFile() with invalid path should return error")
	}

	// エラーメッセージに期待する文字列が含まれているかチェック
	if !strings.Contains(err.Error(), "失敗しました") {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestWriteToConsole(t *testing.T) {
	service := NewOutputService()
	data := createTestReportData()

	// 標準出力をキャプチャするためのバッファ
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// 標準エラー出力もキャプチャ
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	err := service.WriteToConsole(data, FormatCSV)

	// 標準出力を復元
	w.Close()
	os.Stdout = oldStdout

	// 標準エラー出力を復元
	wErr.Close()
	os.Stderr = oldStderr

	if err != nil {
		t.Fatalf("WriteToConsole() failed: %v", err)
	}

	// 標準出力の内容を読み取り
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// 標準エラー出力の内容を読み取り
	bufErr := make([]byte, 1024)
	nErr, _ := rErr.Read(bufErr)
	stderrOutput := string(bufErr[:nErr])
	_ = stderrOutput // 使用していない変数の警告を回避

	// 標準出力にCSVデータが含まれているかチェック
	if !strings.Contains(output, "property_id,date,pagePath") {
		t.Error("Standard output does not contain expected CSV header")
	}

	// 標準エラー出力にサマリー情報が含まれているかチェック
	if !strings.Contains(stderrOutput, "CSV出力を標準出力に書き込みます") {
		t.Error("Standard error does not contain expected summary message")
	}
}

func TestValidateData(t *testing.T) {
	service := NewOutputService().(*OutputServiceImpl)

	// 正常なデータのテスト
	validData := createTestReportData()
	err := service.ValidateData(validData)
	if err != nil {
		t.Errorf("ValidateData() with valid data failed: %v", err)
	}

	// nilデータのテスト
	err = service.ValidateData(nil)
	if err == nil {
		t.Error("ValidateData() with nil data should return error")
	}

	// 空のヘッダーのテスト
	emptyHeaderData := &analytics.ReportData{
		Headers: []string{},
		Rows:    [][]string{},
	}
	err = service.ValidateData(emptyHeaderData)
	if err == nil {
		t.Error("ValidateData() with empty headers should return error")
	}

	// 列数が一致しないデータのテスト
	invalidData := createInvalidReportData()
	err = service.ValidateData(invalidData)
	if err == nil {
		t.Error("ValidateData() with invalid column count should return error")
	}
}

func TestGetOutputSummary(t *testing.T) {
	service := NewOutputService().(*OutputServiceImpl)

	// 正常なデータのテスト
	data := createTestReportData()
	summary := service.GetOutputSummary(data, FormatCSV)

	expectedStrings := []string{
		"CSV出力サマリー:",
		"総レコード数: 4",
		"出力行数: 5行",
		"列数: 7列",
		"期間: 2023-01-01 - 2023-01-02",
		"プロパティ数: 1",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(summary, expected) {
			t.Errorf("Summary does not contain expected string: %s\nActual summary: %s", expected, summary)
		}
	}

	// nilデータのテスト
	nilSummary := service.GetOutputSummary(nil, FormatCSV)
	if nilSummary != "データなし" {
		t.Errorf("Expected 'データなし' for nil data, got: %s", nilSummary)
	}
}

func TestCSVEncoding(t *testing.T) {
	service := NewOutputService()

	// 日本語を含むテストデータ
	japaneseData := &analytics.ReportData{
		Headers: []string{"プロパティID", "日付", "ページパス", "セッション数"},
		Rows: [][]string{
			{"123456789", "2023-01-01", "/ホーム", "1250"},
			{"123456789", "2023-01-01", "/会社概要", "450"},
		},
		Summary: analytics.ReportSummary{
			TotalRows:  2,
			DateRange:  "2023-01-01 - 2023-01-01",
			Properties: []string{"123456789"},
		},
	}

	var buf bytes.Buffer
	err := service.WriteCSV(japaneseData, &buf)
	if err != nil {
		t.Fatalf("WriteCSV() with Japanese data failed: %v", err)
	}

	output := buf.String()

	// 日本語文字が含まれているかチェック
	if !strings.Contains(output, "プロパティID") {
		t.Error("Japanese characters not properly encoded in CSV output")
	}

	if !strings.Contains(output, "ホーム") {
		t.Error("Japanese characters in data not properly encoded in CSV output")
	}
}

func TestWriteOutput_ToConsole(t *testing.T) {
	service := NewOutputService()
	data := createTestReportData()

	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// 標準エラー出力もキャプチャ
	oldStderr := os.Stderr
	_, wErr, _ := os.Pipe()
	os.Stderr = wErr

	// 空文字列（標準出力）
	err := service.WriteOutput(data, "", FormatCSV)
	if err != nil {
		t.Fatalf("WriteOutput() to console failed: %v", err)
	}

	// "-"（標準出力）
	err = service.WriteOutput(data, "-", FormatCSV)
	if err != nil {
		t.Fatalf("WriteOutput() to console with '-' failed: %v", err)
	}

	// 標準出力を復元
	w.Close()
	os.Stdout = oldStdout

	// 標準エラー出力を復元
	wErr.Close()
	os.Stderr = oldStderr

	// 標準出力の内容を読み取り
	buf := make([]byte, 2048)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// CSVデータが2回出力されているかチェック（空文字列と"-"の両方）
	headerCount := strings.Count(output, "property_id,date,pagePath")
	if headerCount != 2 {
		t.Errorf("Expected CSV header to appear 2 times, got %d", headerCount)
	}
}

func TestWriteOutput_ToFile(t *testing.T) {
	service := NewOutputService()
	data := createTestReportData()

	// 一時ファイル名を生成
	tempFile := "test_write_output.csv"
	defer os.Remove(tempFile)

	err := service.WriteOutput(data, tempFile, FormatCSV)
	if err != nil {
		t.Fatalf("WriteOutput() to file failed: %v", err)
	}

	// ファイルが作成されたことを確認
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// ファイル内容を読み込んで検証
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	expectedLines := len(data.Rows) + 1
	if len(lines) != expectedLines {
		t.Errorf("Expected %d lines in file, got %d", expectedLines, len(lines))
	}
}

func TestWriteOutput_InvalidData(t *testing.T) {
	service := NewOutputService()

	// nilデータのテスト
	err := service.WriteOutput(nil, "test.csv", FormatCSV)
	if err == nil {
		t.Error("WriteOutput() with nil data should return error")
	}

	// 無効なデータのテスト
	invalidData := createInvalidReportData()
	err = service.WriteOutput(invalidData, "test.csv", FormatCSV)
	if err == nil {
		t.Error("WriteOutput() with invalid data should return error")
	}
}

func TestWriteToFileWithErrorHandling_DirectoryCreation(t *testing.T) {
	service := NewOutputService().(*OutputServiceImpl)
	data := createTestReportData()

	// 存在しないディレクトリを含むパス
	testDir := "test_dir"
	testFile := testDir + "/output.csv"
	defer os.RemoveAll(testDir) // テスト後にディレクトリを削除

	err := service.WriteToFileWithErrorHandling(data, testFile, FormatCSV)
	if err != nil {
		t.Fatalf("WriteToFileWithErrorHandling() with directory creation failed: %v", err)
	}

	// ファイルが作成されたことを確認
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Output file was not created in new directory")
	}

	// ディレクトリが作成されたことを確認
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Fatal("Directory was not created")
	}
}

func TestValidateFilePath(t *testing.T) {
	service := NewOutputService().(*OutputServiceImpl)

	// 正常なファイルパス
	validPaths := []string{
		"output.csv",
		"data/output.csv",
		"../output.csv",
		"output_2023.csv",
	}

	for _, path := range validPaths {
		err := service.validateFilePath(path)
		if err != nil {
			t.Errorf("validateFilePath() failed for valid path '%s': %v", path, err)
		}
	}

	// 無効なファイルパス
	invalidPaths := []string{
		"",         // 空文字列
		"   ",      // 空白のみ
		"file\x00", // null文字
		"file\n",   // 改行文字
		"file\r",   // キャリッジリターン
	}

	for _, path := range invalidPaths {
		err := service.validateFilePath(path)
		if err == nil {
			t.Errorf("validateFilePath() should fail for invalid path '%s'", path)
		}
	}
}

func TestHandleFileCreationError(t *testing.T) {
	service := NewOutputService().(*OutputServiceImpl)

	// 権限エラーのテスト
	permErr := &os.PathError{Op: "open", Path: "/test", Err: os.ErrPermission}
	err := service.handleFileCreationError("test.csv", permErr)
	if !strings.Contains(err.Error(), "書き込み権限がありません") {
		t.Errorf("Expected permission error message, got: %s", err.Error())
	}

	// 一般的なエラーのテスト
	genericErr := fmt.Errorf("generic error")
	err = service.handleFileCreationError("test.csv", genericErr)
	if !strings.Contains(err.Error(), "作成に失敗しました") {
		t.Errorf("Expected generic error message, got: %s", err.Error())
	}
}

// TestMultiStreamURLJoining は複数ストリームでのURL結合機能をテストする
func TestMultiStreamURLJoining(t *testing.T) {
	// 複数ストリームのテストデータを作成
	testData := &analytics.ReportData{
		Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions", "activeUsers"},
		Rows: [][]string{
			// ストリーム1のデータ
			{"320031301", "3282358539", "2024-01-01", "/entry/2024/01/01/120000", "10", "8"},
			{"320031301", "3282358539", "2024-01-01", "/entry/2024/01/02/120000", "15", "12"},
			// ストリーム2のデータ
			{"321189208", "3803158344", "2024-01-01", "/articles/go-tutorial", "20", "18"},
			{"321189208", "3803158344", "2024-01-01", "/articles/python-basics", "25", "22"},
		},
		StreamURLs: map[string]string{
			"3282358539": "https://ymotongpoo.hatenablog.com",
			"3803158344": "https://zenn.com/ymotongpoo",
		},
		Summary: analytics.ReportSummary{
			TotalRows:  4,
			DateRange:  "2024-01-01 - 2024-01-01",
			Properties: []string{"320031301", "321189208"},
		},
	}

	// 出力サービスを作成
	outputService := NewOutputService()

	// CSV出力をテスト
	var buf bytes.Buffer
	err := outputService.WriteCSV(testData, &buf)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// ヘッダー行をチェック（pagePathがfullURLに変換されているか）
	expectedHeader := "property_id,stream_id,date,fullURL,sessions,activeUsers"
	if lines[0] != expectedHeader {
		t.Errorf("Expected header: %s, got: %s", expectedHeader, lines[0])
	}

	// 各行のURL結合結果をチェック
	expectedRows := []string{
		"320031301,3282358539,2024-01-01,https://ymotongpoo.hatenablog.com/entry/2024/01/01/120000,10,8",
		"320031301,3282358539,2024-01-01,https://ymotongpoo.hatenablog.com/entry/2024/01/02/120000,15,12",
		"321189208,3803158344,2024-01-01,https://zenn.com/ymotongpoo/articles/go-tutorial,20,18",
		"321189208,3803158344,2024-01-01,https://zenn.com/ymotongpoo/articles/python-basics,25,22",
	}

	for i, expectedRow := range expectedRows {
		if i+1 >= len(lines) {
			t.Fatalf("Expected at least %d lines, got %d", i+2, len(lines))
		}
		if lines[i+1] != expectedRow {
			t.Errorf("Row %d: expected %s, got %s", i+1, expectedRow, lines[i+1])
		}
	}
}

// TestURLJoiningWithDifferentStreamURLs は異なるストリームURLでのURL結合をテストする
func TestURLJoiningWithDifferentStreamURLs(t *testing.T) {
	testData := &analytics.ReportData{
		Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
		Rows: [][]string{
			// ストリーム1のデータ
			{"320031301", "3282358539", "2024-01-01", "/entry/2024/01/01/120000", "100"},
			{"320031301", "3282358539", "2024-01-01", "/entry/2024/01/02/120000", "150"},
			// ストリーム2のデータ
			{"321189208", "3803158344", "2024-01-01", "/articles/golang-tips", "200"},
			{"321189208", "3803158344", "2024-01-01", "/articles/web-development", "250"},
		},
		StreamURLs: map[string]string{
			"3282358539": "https://ymotongpoo.hatenablog.com/",
			"3803158344": "https://zenn.com/ymotongpoo/",
		},
		Summary: analytics.ReportSummary{
			TotalRows: 4,
			DateRange: "2024-01-01 to 2024-01-01",
		},
	}

	// 出力サービスを作成
	service := NewOutputService()

	// CSV出力をテスト
	var buf bytes.Buffer
	err := service.WriteCSV(testData, &buf)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// 期待される結果
	expectedLines := []string{
		"property_id,stream_id,date,fullURL,sessions",
		"320031301,3282358539,2024-01-01,https://ymotongpoo.hatenablog.com/entry/2024/01/01/120000,100",
		"320031301,3282358539,2024-01-01,https://ymotongpoo.hatenablog.com/entry/2024/01/02/120000,150",
		"321189208,3803158344,2024-01-01,https://zenn.com/ymotongpoo/articles/golang-tips,200",
		"321189208,3803158344,2024-01-01,https://zenn.com/ymotongpoo/articles/web-development,250",
	}

	if len(lines) != len(expectedLines) {
		t.Fatalf("Expected %d lines, got %d", len(expectedLines), len(lines))
	}

	for i, expectedLine := range expectedLines {
		if lines[i] != expectedLine {
			t.Errorf("Line %d: expected %s, got %s", i, expectedLine, lines[i])
		}
	}
}

// TestURLJoiningWithMissingStreamURL はStreamURLが設定されていない場合のテスト
func TestURLJoiningWithMissingStreamURL(t *testing.T) {
	testData := &analytics.ReportData{
		Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
		Rows: [][]string{
			{"320031301", "3282358539", "2024-01-01", "/entry/2024/01/01/120000", "100"},
			{"321189208", "unknown_stream", "2024-01-01", "/articles/golang-tips", "200"},
		},
		StreamURLs: map[string]string{
			"3282358539": "https://ymotongpoo.hatenablog.com/",
			// "unknown_stream" は意図的に設定しない
		},
		Summary: analytics.ReportSummary{
			TotalRows: 2,
			DateRange: "2024-01-01 to 2024-01-01",
		},
	}

	service := NewOutputService()
	var buf bytes.Buffer
	err := service.WriteCSV(testData, &buf)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// 期待される結果（unknown_streamの場合はpagePathがそのまま残る）
	expectedLines := []string{
		"property_id,stream_id,date,fullURL,sessions",
		"320031301,3282358539,2024-01-01,https://ymotongpoo.hatenablog.com/entry/2024/01/01/120000,100",
		"321189208,unknown_stream,2024-01-01,/articles/golang-tips,200",
	}

	if len(lines) != len(expectedLines) {
		t.Fatalf("Expected %d lines, got %d", len(expectedLines), len(lines))
	}

	for i, expectedLine := range expectedLines {
		if lines[i] != expectedLine {
			t.Errorf("Line %d: expected %s, got %s", i, expectedLine, lines[i])
		}
	}
}

// TestJSONOutputWithMultiStream は複数ストリームでのJSON出力をテストする
func TestJSONOutputWithMultiStream(t *testing.T) {
	testData := &analytics.ReportData{
		Headers: []string{"property_id", "stream_id", "date", "pagePath", "sessions"},
		Rows: [][]string{
			{"320031301", "3282358539", "2024-01-01", "/entry/2024/01/01/120000", "100"},
			{"321189208", "3803158344", "2024-01-01", "/articles/golang-tips", "200"},
		},
		StreamURLs: map[string]string{
			"3282358539": "https://ymotongpoo.hatenablog.com/",
			"3803158344": "https://zenn.com/ymotongpoo/",
		},
		Summary: analytics.ReportSummary{
			TotalRows: 2,
			DateRange: "2024-01-01 to 2024-01-01",
		},
	}

	service := NewOutputService()
	var buf bytes.Buffer
	err := service.WriteJSON(testData, &buf)
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	output := buf.String()

	// JSON出力にURL結合結果が含まれているかチェック
	expectedURLs := []string{
		"https://ymotongpoo.hatenablog.com/entry/2024/01/01/120000",
		"https://zenn.com/ymotongpoo/articles/golang-tips",
	}

	for _, expectedURL := range expectedURLs {
		if !strings.Contains(output, expectedURL) {
			t.Errorf("JSON output does not contain expected URL: %s", expectedURL)
		}
	}

	// fullURLキーが含まれているかチェック
	if !strings.Contains(output, "\"fullURL\"") {
		t.Error("JSON output does not contain 'fullURL' key")
	}

	// ストリームIDがメタデータに含まれているかチェック
	if !strings.Contains(output, "\"stream_id\"") {
		t.Error("JSON output does not contain 'stream_id' in metadata")
	}
}
