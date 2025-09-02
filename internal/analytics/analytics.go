package analytics

import (
	"context"

	"github.com/ymotongpoo/ga/internal/config"
)

// AnalyticsService はGoogle Analytics 4データ取得を提供するインターフェース
type AnalyticsService interface {
	// GetReportData は指定された設定に基づいてレポートデータを取得する
	GetReportData(ctx context.Context, config *config.Config) (*ReportData, error)
}

// GA4Client はGoogle Analytics 4 APIクライアント
type GA4Client struct {
	// service *analyticsreporting.Service // TODO: 後で追加
	config *config.Config
}

// ReportData はレポートデータを表す構造体
type ReportData struct {
	Headers []string
	Rows    [][]string
	Summary ReportSummary
}

// ReportSummary はレポートサマリーを表す構造体
type ReportSummary struct {
	TotalRows  int
	DateRange  string
	Properties []string
}

// GAReportResponse はGoogle Analytics APIレスポンスを表す構造体
type GAReportResponse struct {
	DimensionHeaders []DimensionHeader
	MetricHeaders    []MetricHeader
	Rows             []ReportRow
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
func NewAnalyticsService() AnalyticsService {
	return &AnalyticsServiceImpl{}
}

// GetReportData は指定された設定に基づいてレポートデータを取得する
func (a *AnalyticsServiceImpl) GetReportData(ctx context.Context, config *config.Config) (*ReportData, error) {
	// TODO: 実装予定
	return nil, nil
}