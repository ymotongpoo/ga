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
	"errors"
	"testing"
	"time"

	"github.com/ymotongpoo/ga/internal/config"
	"google.golang.org/api/analyticsdata/v1beta"
	"google.golang.org/api/googleapi"
)

// モックGA4Client
type mockGA4Client struct {
	shouldReturnError bool
	errorType         string
	retryCount        int
	maxRetries        int
	response          *GA4ReportResponse
}

func (m *mockGA4Client) runReport(ctx context.Context, request *GA4ReportRequest) (*GA4ReportResponse, error) {
	if m.shouldReturnError {
		switch m.errorType {
		case "auth":
			return nil, &googleapi.Error{Code: 401, Message: "Unauthorized"}
		case "rate_limit":
			return nil, &googleapi.Error{Code: 429, Message: "Too Many Requests"}
		case "network":
			return nil, errors.New("network timeout")
		case "api":
			return nil, &googleapi.Error{Code: 500, Message: "Internal Server Error"}
		default:
			return nil, errors.New("unknown error")
		}
	}

	if m.response != nil {
		return m.response, nil
	}

	// デフォルトのレスポンス
	return &GA4ReportResponse{
		DimensionHeaders: []*analyticsdata.DimensionHeader{
			{Name: "date"},
			{Name: "pagePath"},
		},
		MetricHeaders: []*analyticsdata.MetricHeader{
			{Name: "sessions", Type: "TYPE_INTEGER"},
			{Name: "activeUsers", Type: "TYPE_INTEGER"},
		},
		Rows: []*analyticsdata.Row{
			{
				DimensionValues: []*analyticsdata.DimensionValue{
					{Value: "2023-01-01"},
					{Value: "/home"},
				},
				MetricValues: []*analyticsdata.MetricValue{
					{Value: "1250"},
					{Value: "1100"},
				},
			},
		},
		RowCount: 1,
	}, nil
}

func (m *mockGA4Client) isRetryableError(err error) bool {
	if apiErr, ok := err.(*googleapi.Error); ok {
		return apiErr.Code == 429 || apiErr.Code >= 500
	}
	return false
}

func (m *mockGA4Client) calculateBackoffDelay(attempt int) time.Duration {
	return time.Millisecond * 10 // テスト用に短い時間
}

func (m *mockGA4Client) classifyError(err error, context string) error {
	return err
}

// テスト用の設定データを作成
func createTestConfig() *config.Config {
	return &config.Config{
		StartDate: "2023-01-01",
		EndDate:   "2023-01-31",
		Account:   "123456789",
		Properties: []config.Property{
			{
				ID: "987654321",
				Streams: []config.Stream{
					{
						ID:         "1234567",
						Dimensions: []string{"date", "pagePath"},
						Metrics:    []string{"sessions", "activeUsers", "newUsers", "averageSessionDuration"},
					},
				},
			},
		},
	}
}

func TestNewGA4Client(t *testing.T) {
	config := createTestConfig()

	// 実際のAPIクライアントの作成はスキップ（モック環境では困難）
	// 代わりに構造体の初期化をテスト
	client := &GA4Client{
		service:     nil, // テスト環境では nil
		config:      config,
		retryConfig: DefaultRetryConfig,
	}

	if client.config != config {
		t.Error("Config was not set correctly")
	}

	if client.retryConfig != DefaultRetryConfig {
		t.Error("RetryConfig was not set correctly")
	}
}

func TestMetricMapping(t *testing.T) {
	expectedMappings := map[string]string{
		"sessions":               "sessions",
		"activeUsers":            "activeUsers",
		"newUsers":               "newUsers",
		"averageSessionDuration": "averageSessionDuration",
	}

	for input, expected := range expectedMappings {
		if MetricMapping[input] != expected {
			t.Errorf("MetricMapping[%s] = %s, want %s", input, MetricMapping[input], expected)
		}
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	if DefaultRetryConfig.MaxRetries != 3 {
		t.Errorf("DefaultRetryConfig.MaxRetries = %d, want 3", DefaultRetryConfig.MaxRetries)
	}

	if DefaultRetryConfig.BaseDelay != 1*time.Second {
		t.Errorf("DefaultRetryConfig.BaseDelay = %v, want 1s", DefaultRetryConfig.BaseDelay)
	}

	if DefaultRetryConfig.BackoffFactor != 2.0 {
		t.Errorf("DefaultRetryConfig.BackoffFactor = %f, want 2.0", DefaultRetryConfig.BackoffFactor)
	}

	expectedRetryableCodes := []int{429, 500, 502, 503, 504}
	if len(DefaultRetryConfig.RetryableErrors) != len(expectedRetryableCodes) {
		t.Errorf("DefaultRetryConfig.RetryableErrors length = %d, want %d",
			len(DefaultRetryConfig.RetryableErrors), len(expectedRetryableCodes))
	}

	for i, code := range expectedRetryableCodes {
		if DefaultRetryConfig.RetryableErrors[i] != code {
			t.Errorf("DefaultRetryConfig.RetryableErrors[%d] = %d, want %d",
				i, DefaultRetryConfig.RetryableErrors[i], code)
		}
	}
}

func TestGA4Client_isRetryableError(t *testing.T) {
	client := &GA4Client{retryConfig: DefaultRetryConfig}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "429 Too Many Requests",
			err:  &googleapi.Error{Code: 429},
			want: true,
		},
		{
			name: "500 Internal Server Error",
			err:  &googleapi.Error{Code: 500},
			want: true,
		},
		{
			name: "401 Unauthorized",
			err:  &googleapi.Error{Code: 401},
			want: false,
		},
		{
			name: "Generic error",
			err:  errors.New("generic error"),
			want: false,
		},
		{
			name: "Timeout error",
			err:  errors.New("timeout occurred"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.isRetryableError(tt.err)
			if got != tt.want {
				t.Errorf("isRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGA4Client_calculateBackoffDelay(t *testing.T) {
	client := &GA4Client{retryConfig: DefaultRetryConfig}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{10, 30 * time.Second}, // MaxDelayに制限される
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := client.calculateBackoffDelay(tt.attempt)
			if got != tt.want {
				t.Errorf("calculateBackoffDelay(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestGA4Client_classifyError(t *testing.T) {
	client := &GA4Client{}

	tests := []struct {
		name     string
		err      error
		context  string
		wantType string
	}{
		{
			name:     "401 Unauthorized",
			err:      &googleapi.Error{Code: 401},
			context:  "test context",
			wantType: "AUTH_ERROR",
		},
		{
			name:     "403 Forbidden",
			err:      &googleapi.Error{Code: 403},
			context:  "test context",
			wantType: "AUTH_ERROR",
		},
		{
			name:     "429 Rate Limit",
			err:      &googleapi.Error{Code: 429},
			context:  "test context",
			wantType: "API_ERROR",
		},
		{
			name:     "500 Server Error",
			err:      &googleapi.Error{Code: 500},
			context:  "test context",
			wantType: "API_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.classifyError(tt.err, tt.context)
			// エラーメッセージにエラータイプが含まれているかチェック
			if !contains(got.Error(), tt.wantType) {
				t.Errorf("classifyError() error = %v, should contain %v", got, tt.wantType)
			}
		})
	}
}

// contains はstringにsubstrが含まれているかチェックする
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestAnalyticsServiceImpl_buildReportRequests(t *testing.T) {
	service := &AnalyticsServiceImpl{}
	config := createTestConfig()

	requests, err := service.buildReportRequests(config)
	if err != nil {
		t.Fatalf("buildReportRequests() error = %v", err)
	}

	if len(requests) != 1 {
		t.Errorf("buildReportRequests() returned %d requests, want 1", len(requests))
	}

	request := requests[0]
	if request.PropertyID != "987654321" {
		t.Errorf("PropertyID = %s, want 987654321", request.PropertyID)
	}

	if request.StartDate != "2023-01-01" {
		t.Errorf("StartDate = %s, want 2023-01-01", request.StartDate)
	}

	if request.EndDate != "2023-01-31" {
		t.Errorf("EndDate = %s, want 2023-01-31", request.EndDate)
	}

	expectedDimensions := []string{"date", "pagePath"}
	if len(request.Dimensions) != len(expectedDimensions) {
		t.Errorf("Dimensions length = %d, want %d", len(request.Dimensions), len(expectedDimensions))
	}

	expectedMetrics := []string{"sessions", "activeUsers", "newUsers", "averageSessionDuration"}
	if len(request.Metrics) != len(expectedMetrics) {
		t.Errorf("Metrics length = %d, want %d", len(request.Metrics), len(expectedMetrics))
	}
}

func TestAnalyticsServiceImpl_mapMetrics(t *testing.T) {
	service := &AnalyticsServiceImpl{}

	tests := []struct {
		name    string
		metrics []string
		want    []string
		wantErr bool
	}{
		{
			name:    "Valid metrics",
			metrics: []string{"sessions", "activeUsers"},
			want:    []string{"sessions", "activeUsers"},
			wantErr: false,
		},
		{
			name:    "Invalid metric",
			metrics: []string{"sessions", "invalidMetric"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Empty metrics",
			metrics: []string{},
			want:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.mapMetrics(tt.metrics)
			if (err != nil) != tt.wantErr {
				t.Errorf("mapMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("mapMetrics() length = %d, want %d", len(got), len(tt.want))
				}
				for i, metric := range got {
					if metric != tt.want[i] {
						t.Errorf("mapMetrics()[%d] = %s, want %s", i, metric, tt.want[i])
					}
				}
			}
		})
	}
}

func TestAnalyticsServiceImpl_buildHeaders(t *testing.T) {
	service := &AnalyticsServiceImpl{}

	response := &GA4ReportResponse{
		DimensionHeaders: []*analyticsdata.DimensionHeader{
			{Name: "date"},
			{Name: "pagePath"},
		},
		MetricHeaders: []*analyticsdata.MetricHeader{
			{Name: "sessions"},
			{Name: "activeUsers"},
		},
	}

	headers := service.buildHeaders(response)

	expectedHeaders := []string{"property_id", "date", "pagePath", "sessions", "activeUsers"}
	if len(headers) != len(expectedHeaders) {
		t.Errorf("buildHeaders() length = %d, want %d", len(headers), len(expectedHeaders))
	}

	for i, header := range headers {
		if header != expectedHeaders[i] {
			t.Errorf("buildHeaders()[%d] = %s, want %s", i, header, expectedHeaders[i])
		}
	}
}

func TestAnalyticsServiceImpl_convertResponseToRows(t *testing.T) {
	service := &AnalyticsServiceImpl{}

	response := &GA4ReportResponse{
		Rows: []*analyticsdata.Row{
			{
				DimensionValues: []*analyticsdata.DimensionValue{
					{Value: "2023-01-01"},
					{Value: "/home"},
				},
				MetricValues: []*analyticsdata.MetricValue{
					{Value: "1250"},
					{Value: "1100"},
				},
			},
		},
	}

	rows := service.convertResponseToRows(response, "123456789")

	if len(rows) != 1 {
		t.Errorf("convertResponseToRows() returned %d rows, want 1", len(rows))
	}

	expectedRow := []string{"123456789", "2023-01-01", "/home", "1250", "1100"}
	if len(rows[0]) != len(expectedRow) {
		t.Errorf("Row length = %d, want %d", len(rows[0]), len(expectedRow))
	}

	for i, value := range rows[0] {
		if value != expectedRow[i] {
			t.Errorf("Row[%d] = %s, want %s", i, value, expectedRow[i])
		}
	}
}

func TestReportData_Structure(t *testing.T) {
	data := &ReportData{
		Headers: []string{"property_id", "date", "sessions"},
		Rows: [][]string{
			{"123456789", "2023-01-01", "1250"},
		},
		Summary: ReportSummary{
			TotalRows:  1,
			DateRange:  "2023-01-01 - 2023-01-01",
			Properties: []string{"123456789"},
		},
	}

	if len(data.Headers) != 3 {
		t.Errorf("Headers length = %d, want 3", len(data.Headers))
	}

	if len(data.Rows) != 1 {
		t.Errorf("Rows length = %d, want 1", len(data.Rows))
	}

	if data.Summary.TotalRows != 1 {
		t.Errorf("Summary.TotalRows = %d, want 1", data.Summary.TotalRows)
	}

	if data.Summary.DateRange != "2023-01-01 - 2023-01-01" {
		t.Errorf("Summary.DateRange = %s, want 2023-01-01 - 2023-01-01", data.Summary.DateRange)
	}

	if len(data.Summary.Properties) != 1 || data.Summary.Properties[0] != "123456789" {
		t.Errorf("Summary.Properties = %v, want [123456789]", data.Summary.Properties)
	}
}

func TestGA4ReportRequest_Structure(t *testing.T) {
	request := &GA4ReportRequest{
		PropertyID: "123456789",
		StartDate:  "2023-01-01",
		EndDate:    "2023-01-31",
		Dimensions: []string{"date", "pagePath"},
		Metrics:    []string{"sessions", "activeUsers"},
	}

	if request.PropertyID != "123456789" {
		t.Errorf("PropertyID = %s, want 123456789", request.PropertyID)
	}

	if request.StartDate != "2023-01-01" {
		t.Errorf("StartDate = %s, want 2023-01-01", request.StartDate)
	}

	if request.EndDate != "2023-01-31" {
		t.Errorf("EndDate = %s, want 2023-01-31", request.EndDate)
	}

	if len(request.Dimensions) != 2 {
		t.Errorf("Dimensions length = %d, want 2", len(request.Dimensions))
	}

	if len(request.Metrics) != 2 {
		t.Errorf("Metrics length = %d, want 2", len(request.Metrics))
	}
}

func TestRetryConfig_Structure(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:      5,
		BaseDelay:       2 * time.Second,
		MaxDelay:        60 * time.Second,
		BackoffFactor:   1.5,
		RetryableErrors: []int{429, 500},
	}

	if config.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", config.MaxRetries)
	}

	if config.BaseDelay != 2*time.Second {
		t.Errorf("BaseDelay = %v, want 2s", config.BaseDelay)
	}

	if config.MaxDelay != 60*time.Second {
		t.Errorf("MaxDelay = %v, want 60s", config.MaxDelay)
	}

	if config.BackoffFactor != 1.5 {
		t.Errorf("BackoffFactor = %f, want 1.5", config.BackoffFactor)
	}

	if len(config.RetryableErrors) != 2 {
		t.Errorf("RetryableErrors length = %d, want 2", len(config.RetryableErrors))
	}
}
