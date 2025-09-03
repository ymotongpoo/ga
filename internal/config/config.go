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

package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

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
	BaseURL    string   `yaml:"base_url,omitempty"`
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
	// ファイルの存在確認
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("設定ファイルが見つかりません: %s", path)
	}

	// ファイル読み込み
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	// YAML解析
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("YAML形式が不正です: %w", err)
	}

	return &config, nil
}

// ValidateConfig は設定ファイルの内容を検証する
func (c *ConfigServiceImpl) ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("設定が空です")
	}

	// 必須項目の検証
	if err := c.validateRequiredFields(config); err != nil {
		return err
	}

	// 日付形式の検証
	if err := c.validateDateFormat(config); err != nil {
		return err
	}

	// ID形式の検証
	if err := c.validateIDFormats(config); err != nil {
		return err
	}

	// プロパティとストリームの検証
	if err := c.validatePropertiesAndStreams(config); err != nil {
		return err
	}

	return nil
}

// validateRequiredFields は必須項目の存在を検証する
func (c *ConfigServiceImpl) validateRequiredFields(config *Config) error {
	if strings.TrimSpace(config.StartDate) == "" {
		return fmt.Errorf("start_date は必須項目です")
	}
	if strings.TrimSpace(config.EndDate) == "" {
		return fmt.Errorf("end_date は必須項目です")
	}
	if strings.TrimSpace(config.Account) == "" {
		return fmt.Errorf("account は必須項目です")
	}
	if len(config.Properties) == 0 {
		return fmt.Errorf("properties は必須項目です")
	}
	return nil
}

// validateDateFormat は日付形式を検証する
func (c *ConfigServiceImpl) validateDateFormat(config *Config) error {
	dateFormat := "2006-01-02"

	// start_date の検証
	if _, err := time.Parse(dateFormat, config.StartDate); err != nil {
		return fmt.Errorf("start_date の形式が不正です（YYYY-MM-DD形式で入力してください）: %s", config.StartDate)
	}

	// end_date の検証
	if _, err := time.Parse(dateFormat, config.EndDate); err != nil {
		return fmt.Errorf("end_date の形式が不正です（YYYY-MM-DD形式で入力してください）: %s", config.EndDate)
	}

	// 日付の論理的検証（開始日 <= 終了日）
	startDate, _ := time.Parse(dateFormat, config.StartDate)
	endDate, _ := time.Parse(dateFormat, config.EndDate)
	if startDate.After(endDate) {
		return fmt.Errorf("start_date は end_date より前の日付である必要があります")
	}

	return nil
}

// validateIDFormats はID形式を検証する
func (c *ConfigServiceImpl) validateIDFormats(config *Config) error {
	// アカウントIDの検証（数字のみ）
	if matched, _ := regexp.MatchString(`^\d+$`, config.Account); !matched {
		return fmt.Errorf("account ID の形式が不正です（数字のみ）: %s", config.Account)
	}

	return nil
}

// validatePropertiesAndStreams はプロパティとストリームの検証を行う
func (c *ConfigServiceImpl) validatePropertiesAndStreams(config *Config) error {
	for i, property := range config.Properties {
		// プロパティIDの検証
		if strings.TrimSpace(property.ID) == "" {
			return fmt.Errorf("properties[%d].property は必須項目です", i)
		}
		if matched, _ := regexp.MatchString(`^\d+$`, property.ID); !matched {
			return fmt.Errorf("properties[%d].property ID の形式が不正です（数字のみ）: %s", i, property.ID)
		}

		// ストリームの検証
		if len(property.Streams) == 0 {
			return fmt.Errorf("properties[%d].streams は必須項目です", i)
		}

		for j, stream := range property.Streams {
			// ストリームIDの検証
			if strings.TrimSpace(stream.ID) == "" {
				return fmt.Errorf("properties[%d].streams[%d].stream は必須項目です", i, j)
			}
			if matched, _ := regexp.MatchString(`^\d+$`, stream.ID); !matched {
				return fmt.Errorf("properties[%d].streams[%d].stream ID の形式が不正です（数字のみ）: %s", i, j, stream.ID)
			}

			// base_urlの検証（オプション項目）
			if err := c.validateBaseURL(stream.BaseURL, i, j); err != nil {
				return err
			}

			// ディメンションとメトリクスの検証
			if len(stream.Dimensions) == 0 {
				return fmt.Errorf("properties[%d].streams[%d].dimensions は必須項目です", i, j)
			}
			if len(stream.Metrics) == 0 {
				return fmt.Errorf("properties[%d].streams[%d].metrics は必須項目です", i, j)
			}

			// 有効なメトリクスの検証
			validMetrics := map[string]bool{
				"sessions":               true,
				"activeUsers":            true,
				"newUsers":               true,
				"averageSessionDuration": true,
			}

			for _, metric := range stream.Metrics {
				if !validMetrics[metric] {
					return fmt.Errorf("properties[%d].streams[%d] に無効なメトリクスが含まれています: %s", i, j, metric)
				}
			}
		}
	}

	return nil
}

// validateBaseURL はbase_urlの妥当性を検証する
func (c *ConfigServiceImpl) validateBaseURL(baseURL string, propertyIndex, streamIndex int) error {
	// base_urlは省略可能なので、空文字列の場合は検証をスキップ
	if strings.TrimSpace(baseURL) == "" {
		return nil
	}

	// URLの形式検証（http://またはhttps://で始まる）
	urlPattern := `^https?://[^\s/$.?#].[^\s]*$`
	matched, err := regexp.MatchString(urlPattern, baseURL)
	if err != nil {
		return fmt.Errorf("properties[%d].streams[%d].base_url の検証中にエラーが発生しました: %w", propertyIndex, streamIndex, err)
	}
	if !matched {
		return fmt.Errorf("properties[%d].streams[%d].base_url の形式が不正です（http://またはhttps://で始まる有効なURLを入力してください）: %s", propertyIndex, streamIndex, baseURL)
	}

	return nil
}
