package config

// ConfigService は設定ファイル処理を提供するインターフェース
type ConfigService interface {
	// LoadConfig は指定されたパスから設定ファイルを読み込む
	LoadConfig(path string) (*Config, error)

	// ValidateConfig は設定ファイルの内容を検証する
	ValidateConfig(config *Config) error
}

// Config はアプリケーション設定を表す構造体
type Config struct {
	StartDate  string     `yaml:"start_date"`
	EndDate    string     `yaml:"end_date"`
	Account    string     `yaml:"account"`
	Properties []Property `yaml:"properties"`
}

// Property はGoogle Analytics プロパティを表す構造体
type Property struct {
	ID      string   `yaml:"property"`
	Streams []Stream `yaml:"streams"`
}

// Stream はGoogle Analytics ストリームを表す構造体
type Stream struct {
	ID         string   `yaml:"stream"`
	Dimensions []string `yaml:"dimensions"`
	Metrics    []string `yaml:"metrics"`
}

// ConfigServiceImpl はConfigServiceの実装
type ConfigServiceImpl struct{}

// NewConfigService は新しいConfigServiceを作成する
func NewConfigService() ConfigService {
	return &ConfigServiceImpl{}
}

// LoadConfig は指定されたパスから設定ファイルを読み込む
func (c *ConfigServiceImpl) LoadConfig(path string) (*Config, error) {
	// TODO: 実装予定
	return nil, nil
}

// ValidateConfig は設定ファイルの内容を検証する
func (c *ConfigServiceImpl) ValidateConfig(config *Config) error {
	// TODO: 実装予定
	return nil
}