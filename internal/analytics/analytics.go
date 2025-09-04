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

package analytics

import (
	"context"
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ymotongpoo/ga/internal/config"
	"github.com/ymotongpoo/ga/internal/errors"
	"golang.org/x/oauth2"
	"google.golang.org/api/analyticsdata/v1beta"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// AnalyticsService はGoogle Analytics 4データ取得を提供するインターフェース
type AnalyticsService interface {
	// GetReportData は指定された設定に基づいてレポートデータを取得する
	GetReportData(ctx context.Context, config *config.Config) (*ReportData, error)
}

// GA4Client はGoogle Analytics 4 APIクライアント
type GA4Client struct {
	service     *analyticsdata.Service
	config      *config.Config
	retryConfig *RetryConfig
}

// ReportData はレポートデータを表す構造体
type ReportData struct {
	Headers    []string
	Rows       [][]string
	Summary    ReportSummary
	StreamURLs map[string]string // ストリームID -> ベースURL のマッピング
}

// ReportSummary はレポートサマリーを表す構造体
type ReportSummary struct {
	TotalRows  int
	DateRange  string
	Properties []string
}

// GA4ReportRequest はGA4 APIリクエストを表す構造体
type GA4ReportRequest struct {
	PropertyID string
	StreamID   string // URL結合機能のために追加
	StartDate  string
	EndDate    string
	Dimensions []string
	Metrics    []string
}

// MetricMapping はメトリクス名のマッピングを定義
var MetricMapping = map[string]string{
	"sessions":               "sessions",
	"activeUsers":            "activeUsers",
	"newUsers":               "newUsers",
	"averageSessionDuration": "averageSessionDuration",
}

// RetryConfig はリトライ設定を表す構造体
type RetryConfig struct {
	MaxRetries      int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []int // HTTPステータスコード
}

// DefaultRetryConfig はデフォルトのリトライ設定
var DefaultRetryConfig = &RetryConfig{
	MaxRetries:    3,
	BaseDelay:     1 * time.Second,
	MaxDelay:      30 * time.Second,
	BackoffFactor: 2.0,
	RetryableErrors: []int{
		429, // Too Many Requests
		500, // Internal Server Error
		502, // Bad Gateway
		503, // Service Unavailable
		504, // Gateway Timeout
	},
}

// GA4ReportResponse はGoogle Analytics Data APIレスポンスを表す構造体
type GA4ReportResponse struct {
	DimensionHeaders []*analyticsdata.DimensionHeader
	MetricHeaders    []*analyticsdata.MetricHeader
	Rows             []*analyticsdata.Row
	RowCount         int64
}

// DimensionHeader はディメンションヘッダーを表す構造体
type DimensionHeader struct {
	Name string
}

// MetricHeader はメトリクスヘッダーを表す構造体
type MetricHeader struct {
	Name string
	Type string
}

// ReportRow はレポート行を表す構造体
type ReportRow struct {
	Dimensions []string
	Metrics    []MetricValue
}

// MetricValue はメトリクス値を表す構造体
type MetricValue struct {
	Value string
}

// AnalyticsServiceImpl はAnalyticsServiceの実装
type AnalyticsServiceImpl struct {
	client *GA4Client
}

// NewAnalyticsService は新しいAnalyticsServiceを作成する
func NewAnalyticsService(ctx context.Context, token *oauth2.Token, config *config.Config) (AnalyticsService, error) {
	client, err := NewGA4Client(ctx, token, config)
	if err != nil {
		return nil, fmt.Errorf("GA4クライアントの作成に失敗しました: %w", err)
	}

	return &AnalyticsServiceImpl{
		client: client,
	}, nil
}

// NewGA4Client は新しいGA4クライアントを作成する
func NewGA4Client(ctx context.Context, token *oauth2.Token, config *config.Config) (*GA4Client, error) {
	// OAuth2トークンソースを作成
	tokenSource := oauth2.StaticTokenSource(token)

	// Analytics Data APIサービスを作成
	service, err := analyticsdata.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("Analytics Data APIサービスの作成に失敗しました: %w", err)
	}

	return &GA4Client{
		service:     service,
		config:      config,
		retryConfig: DefaultRetryConfig,
	}, nil
}

// isRetryableError はエラーがリトライ可能かどうかを判定する
func (c *GA4Client) isRetryableError(err error) bool {
	// Google API エラーの場合
	if apiErr, ok := err.(*googleapi.Error); ok {
		for _, code := range c.retryConfig.RetryableErrors {
			if apiErr.Code == code {
				return true
			}
		}
		return false
	}

	// ネットワークエラーの場合
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout() || netErr.Temporary()
	}

	// 文字列ベースのエラーチェック
	errStr := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"timeout",
		"connection reset",
		"connection refused",
		"temporary failure",
		"service unavailable",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// calculateBackoffDelay は指数バックオフによる待機時間を計算する
func (c *GA4Client) calculateBackoffDelay(attempt int) time.Duration {
	delay := time.Duration(float64(c.retryConfig.BaseDelay) * math.Pow(c.retryConfig.BackoffFactor, float64(attempt)))
	if delay > c.retryConfig.MaxDelay {
		delay = c.retryConfig.MaxDelay
	}
	return delay
}

// classifyError はエラーを分類してGAErrorを作成する
func (c *GA4Client) classifyError(err error, context string) error {
	if apiErr, ok := err.(*googleapi.Error); ok {
		switch apiErr.Code {
		case 401:
			return errors.NewAuthError(fmt.Sprintf("認証エラー: %s", context), err)
		case 403:
			return errors.NewAuthError(fmt.Sprintf("アクセス権限エラー: %s", context), err)
		case 429:
			return errors.NewAPIError(fmt.Sprintf("API制限エラー: %s", context), err)
		case 400:
			return errors.NewAPIError(fmt.Sprintf("リクエストエラー: %s", context), err)
		default:
			return errors.NewAPIError(fmt.Sprintf("APIエラー (コード: %d): %s", apiErr.Code, context), err)
		}
	}

	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			return errors.NewAPIError(fmt.Sprintf("タイムアウトエラー: %s", context), err)
		}
		return errors.NewAPIError(fmt.Sprintf("ネットワークエラー: %s", context), err)
	}

	return errors.NewAPIError(fmt.Sprintf("予期しないエラー: %s", context), err)
}

// GetReportData は指定された設定に基づいてレポートデータを取得する
func (a *AnalyticsServiceImpl) GetReportData(ctx context.Context, config *config.Config) (*ReportData, error) {
	if a.client == nil {
		return nil, fmt.Errorf("GA4クライアントが初期化されていません")
	}

	// 設定からレポートリクエストを作成
	requests, err := a.buildReportRequests(config)
	if err != nil {
		return nil, fmt.Errorf("レポートリクエストの作成に失敗しました: %w", err)
	}

	// 複数プロパティの並行処理
	return a.fetchDataConcurrently(ctx, requests, config)
}

// fetchDataConcurrently は複数のプロパティから並行してデータを取得する
func (a *AnalyticsServiceImpl) fetchDataConcurrently(ctx context.Context, requests []*GA4ReportRequest, config *config.Config) (*ReportData, error) {
	type result struct {
		response   *GA4ReportResponse
		propertyID string
		streamID   string // ストリームIDを追加
		err        error
	}

	// プログレス表示の初期化
	fmt.Printf("データ取得を開始します... (%d プロパティ)\n", len(requests))

	// 結果チャネル
	resultChan := make(chan result, len(requests))
	var wg sync.WaitGroup

	// プログレス追跡
	completed := 0
	progressChan := make(chan string, len(requests))

	// 各リクエストを並行実行
	for i, request := range requests {
		wg.Add(1)
		go func(req *GA4ReportRequest, index int) {
			defer wg.Done()

			fmt.Printf("[%d/%d] プロパティ %s のデータを取得中...\n", index+1, len(requests), req.PropertyID)

			response, err := a.client.runReport(ctx, req)

			if err != nil {
				progressChan <- fmt.Sprintf("プロパティ %s: エラー", req.PropertyID)
			} else {
				progressChan <- fmt.Sprintf("プロパティ %s: %d レコード取得完了", req.PropertyID, response.RowCount)
			}

			resultChan <- result{
				response:   response,
				propertyID: req.PropertyID,
				streamID:   req.StreamID, // ストリームIDを追加
				err:        err,
			}
		}(request, i)
	}

	// プログレス表示用のgoroutine
	go func() {
		for progress := range progressChan {
			completed++
			fmt.Printf("[%d/%d] %s\n", completed, len(requests), progress)
		}
	}()

	// 全ての goroutine の完了を待つ
	go func() {
		wg.Wait()
		close(resultChan)
		close(progressChan)
	}()

	// 結果を収集
	var allRows [][]string
	var headers []string
	var properties []string
	totalRows := 0
	var errors []error

	for res := range resultChan {
		if res.err != nil {
			errors = append(errors, fmt.Errorf("プロパティ %s のデータ取得に失敗しました: %w", res.propertyID, res.err))
			continue
		}

		// 初回のみヘッダーを設定
		if len(headers) == 0 {
			headers = a.buildHeaders(res.response)
		}

		// データ行を変換（ストリームIDも含める）
		rows := a.convertResponseToRows(res.response, res.propertyID, res.streamID)
		allRows = append(allRows, rows...)
		properties = append(properties, res.propertyID)
		totalRows += int(res.response.RowCount)
	}

	// エラーがある場合は最初のエラーを返す
	if len(errors) > 0 {
		return nil, errors[0]
	}

	// 完了通知
	fmt.Printf("\n✅ データ取得が完了しました!\n")
	fmt.Printf("📊 取得結果:\n")
	fmt.Printf("   - 総レコード数: %d\n", totalRows)
	fmt.Printf("   - 対象プロパティ数: %d\n", len(properties))
	fmt.Printf("   - 期間: %s - %s\n", config.StartDate, config.EndDate)

	if len(properties) > 1 {
		fmt.Printf("   - プロパティ一覧:\n")
		for _, prop := range properties {
			fmt.Printf("     * %s\n", prop)
		}
	}
	fmt.Println()

	// StreamURLsマッピングを構築
	streamURLs := make(map[string]string)
	for _, property := range config.Properties {
		for _, stream := range property.Streams {
			if stream.BaseURL != "" {
				streamURLs[stream.ID] = stream.BaseURL
			}
		}
	}

	return &ReportData{
		Headers:    headers,
		Rows:       allRows,
		StreamURLs: streamURLs,
		Summary: ReportSummary{
			TotalRows:  totalRows,
			DateRange:  fmt.Sprintf("%s - %s", config.StartDate, config.EndDate),
			Properties: properties,
		},
	}, nil
}

// buildReportRequests は設定からレポートリクエストを構築する
func (a *AnalyticsServiceImpl) buildReportRequests(config *config.Config) ([]*GA4ReportRequest, error) {
	var requests []*GA4ReportRequest

	for _, property := range config.Properties {
		for _, stream := range property.Streams {
			// メトリクス名を検証・マッピング
			mappedMetrics, err := a.mapMetrics(stream.Metrics)
			if err != nil {
				return nil, fmt.Errorf("プロパティ %s のメトリクスマッピングに失敗しました: %w", property.ID, err)
			}

			request := &GA4ReportRequest{
				PropertyID: property.ID,
				StreamID:   stream.ID, // ストリームIDを追加
				StartDate:  config.StartDate,
				EndDate:    config.EndDate,
				Dimensions: stream.Dimensions,
				Metrics:    mappedMetrics,
			}
			requests = append(requests, request)
		}
	}

	if len(requests) == 0 {
		return nil, fmt.Errorf("有効なプロパティ設定が見つかりません")
	}

	return requests, nil
}

// mapMetrics はメトリクス名をGA4 API用にマッピングする
func (a *AnalyticsServiceImpl) mapMetrics(metrics []string) ([]string, error) {
	var mappedMetrics []string

	for _, metric := range metrics {
		if mappedName, exists := MetricMapping[metric]; exists {
			mappedMetrics = append(mappedMetrics, mappedName)
		} else {
			return nil, fmt.Errorf("サポートされていないメトリクスです: %s", metric)
		}
	}

	return mappedMetrics, nil
}

// runReport はGA4 APIを呼び出してレポートを実行する（リトライ機能付き）
func (c *GA4Client) runReport(ctx context.Context, request *GA4ReportRequest) (*GA4ReportResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		// リトライの場合は待機
		if attempt > 0 {
			delay := c.calculateBackoffDelay(attempt - 1)
			fmt.Printf("リトライ %d/%d: %v後に再試行します...\n", attempt, c.retryConfig.MaxRetries, delay)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		response, err := c.executeReport(ctx, request)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// リトライ可能なエラーかチェック
		if !c.isRetryableError(err) {
			break
		}

		// 最後の試行の場合はリトライしない
		if attempt == c.retryConfig.MaxRetries {
			break
		}
	}

	// エラーを分類して返す
	context := fmt.Sprintf("プロパティ %s のレポート取得", request.PropertyID)
	return nil, c.classifyError(lastErr, context)
}

// executeReport は実際のAPI呼び出しを実行する
func (c *GA4Client) executeReport(ctx context.Context, request *GA4ReportRequest) (*GA4ReportResponse, error) {
	// ディメンションを構築
	var dimensions []*analyticsdata.Dimension
	for _, dim := range request.Dimensions {
		dimensions = append(dimensions, &analyticsdata.Dimension{
			Name: dim,
		})
	}

	// メトリクスを構築
	var metrics []*analyticsdata.Metric
	for _, metric := range request.Metrics {
		metrics = append(metrics, &analyticsdata.Metric{
			Name: metric,
		})
	}

	// 日付範囲を構築
	dateRanges := []*analyticsdata.DateRange{
		{
			StartDate: request.StartDate,
			EndDate:   request.EndDate,
		},
	}

	// レポートリクエストを構築
	reportRequest := &analyticsdata.RunReportRequest{
		Dimensions: dimensions,
		Metrics:    metrics,
		DateRanges: dateRanges,
	}

	// APIを呼び出し
	propertyPath := fmt.Sprintf("properties/%s", request.PropertyID)
	response, err := c.service.Properties.RunReport(propertyPath, reportRequest).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	return &GA4ReportResponse{
		DimensionHeaders: response.DimensionHeaders,
		MetricHeaders:    response.MetricHeaders,
		Rows:             response.Rows,
		RowCount:         response.RowCount,
	}, nil
}

// buildHeaders はレスポンスからCSVヘッダーを構築する
func (a *AnalyticsServiceImpl) buildHeaders(response *GA4ReportResponse) []string {
	var headers []string

	// プロパティIDを追加
	headers = append(headers, "property_id")

	// ストリームIDを追加（URL結合機能のために必要）
	headers = append(headers, "stream_id")

	// ディメンションヘッダーを追加
	for _, dimHeader := range response.DimensionHeaders {
		headers = append(headers, dimHeader.Name)
	}

	// メトリクスヘッダーを追加
	for _, metricHeader := range response.MetricHeaders {
		headers = append(headers, metricHeader.Name)
	}

	return headers
}

// convertResponseToRows はAPIレスポンスをCSV行に変換する
func (a *AnalyticsServiceImpl) convertResponseToRows(response *GA4ReportResponse, propertyID, streamID string) [][]string {
	var rows [][]string

	for _, row := range response.Rows {
		var csvRow []string

		// プロパティIDを追加
		csvRow = append(csvRow, propertyID)

		// ストリームIDを追加（URL結合機能のために必要）
		csvRow = append(csvRow, streamID)

		// ディメンション値を追加
		for _, dimValue := range row.DimensionValues {
			csvRow = append(csvRow, dimValue.Value)
		}

		// メトリクス値を追加
		for _, metricValue := range row.MetricValues {
			csvRow = append(csvRow, metricValue.Value)
		}

		rows = append(rows, csvRow)
	}

	return rows
}

// GetSessionMetrics は指定されたプロパティからセッション関連メトリクスを取得する
func (a *AnalyticsServiceImpl) GetSessionMetrics(ctx context.Context, propertyID, startDate, endDate string, dimensions []string) (*ReportData, error) {
	fmt.Printf("セッションメトリクスを取得中... (プロパティ: %s)\n", propertyID)

	// セッション関連の標準メトリクス
	metrics := []string{"sessions", "activeUsers", "newUsers", "averageSessionDuration"}

	request := &GA4ReportRequest{
		PropertyID: propertyID,
		StreamID:   "", // セッションメトリクスではストリームIDなし
		StartDate:  startDate,
		EndDate:    endDate,
		Dimensions: dimensions,
		Metrics:    metrics,
	}

	response, err := a.client.runReport(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("セッションメトリクスの取得に失敗しました: %w", err)
	}

	headers := a.buildHeaders(response)
	rows := a.convertResponseToRows(response, propertyID, "") // ストリームIDなし

	fmt.Printf("✅ セッションメトリクス取得完了: %d レコード\n", response.RowCount)

	return &ReportData{
		Headers:    headers,
		Rows:       rows,
		StreamURLs: make(map[string]string), // 空のマッピング
		Summary: ReportSummary{
			TotalRows:  int(response.RowCount),
			DateRange:  fmt.Sprintf("%s - %s", startDate, endDate),
			Properties: []string{propertyID},
		},
	}, nil
}

// GetUserMetrics は指定されたプロパティからユーザー関連メトリクスを取得する
func (a *AnalyticsServiceImpl) GetUserMetrics(ctx context.Context, propertyID, startDate, endDate string, dimensions []string) (*ReportData, error) {
	fmt.Printf("ユーザーメトリクスを取得中... (プロパティ: %s)\n", propertyID)

	// ユーザー関連の標準メトリクス
	metrics := []string{"activeUsers", "newUsers"}

	request := &GA4ReportRequest{
		PropertyID: propertyID,
		StreamID:   "", // ユーザーメトリクスではストリームIDなし
		StartDate:  startDate,
		EndDate:    endDate,
		Dimensions: dimensions,
		Metrics:    metrics,
	}

	response, err := a.client.runReport(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("ユーザーメトリクスの取得に失敗しました: %w", err)
	}

	headers := a.buildHeaders(response)
	rows := a.convertResponseToRows(response, propertyID, "") // ストリームIDなし

	fmt.Printf("✅ ユーザーメトリクス取得完了: %d レコード\n", response.RowCount)

	return &ReportData{
		Headers:    headers,
		Rows:       rows,
		StreamURLs: make(map[string]string), // 空のマッピング
		Summary: ReportSummary{
			TotalRows:  int(response.RowCount),
			DateRange:  fmt.Sprintf("%s - %s", startDate, endDate),
			Properties: []string{propertyID},
		},
	}, nil
}
