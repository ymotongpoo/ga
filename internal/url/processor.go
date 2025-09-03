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

package url

import (
	"fmt"
	"strings"
)

// URLProcessor はURL結合処理を提供する構造体
type URLProcessor struct {
	streamURLs map[string]string // ストリームID -> ベースURL のマッピング
}

// NewURLProcessor は新しいURLProcessorを作成する
func NewURLProcessor(streamURLs map[string]string) *URLProcessor {
	if streamURLs == nil {
		streamURLs = make(map[string]string)
	}
	return &URLProcessor{
		streamURLs: streamURLs,
	}
}

// ProcessPagePath はストリームIDとpagePathを受け取り、適切なフルURLを返す
func (up *URLProcessor) ProcessPagePath(streamID, pagePath string) string {
	baseURL := up.streamURLs[streamID]
	return ProcessPagePath(baseURL, pagePath)
}

// ProcessPagePath はベースURLとpagePathを結合してフルURLを生成する
// この関数は以下のルールに従って処理を行う:
// 1. pagePathが絶対URL（http://またはhttps://で始まる）の場合はそのまま返す
// 2. baseURLが設定されていない場合はpagePathをそのまま返す
// 3. pagePathが空文字列またはnullの場合はbaseURLのみを返す
// 4. スラッシュの重複を適切に処理してURL結合を行う
func ProcessPagePath(baseURL, pagePath string) string {
	// pagePathが絶対URLの場合はそのまま返す
	if strings.HasPrefix(pagePath, "http://") || strings.HasPrefix(pagePath, "https://") {
		return pagePath
	}

	// ベースURLが設定されていない場合はpagePathをそのまま返す
	if strings.TrimSpace(baseURL) == "" {
		return pagePath
	}

	// pagePathが空文字列またはnullの場合はベースURLのみ返す
	if strings.TrimSpace(pagePath) == "" {
		return baseURL
	}

	// スラッシュの重複を処理してURL結合
	baseURL = strings.TrimSuffix(baseURL, "/")
	pagePath = strings.TrimPrefix(pagePath, "/")

	return fmt.Sprintf("%s/%s", baseURL, pagePath)
}

// SetStreamURL はストリームIDに対応するベースURLを設定する
func (up *URLProcessor) SetStreamURL(streamID, baseURL string) {
	up.streamURLs[streamID] = baseURL
}

// GetStreamURL はストリームIDに対応するベースURLを取得する
func (up *URLProcessor) GetStreamURL(streamID string) string {
	return up.streamURLs[streamID]
}

// GetAllStreamURLs は全てのストリームURLマッピングを取得する
func (up *URLProcessor) GetAllStreamURLs() map[string]string {
	// コピーを返してカプセル化を保つ
	result := make(map[string]string)
	for k, v := range up.streamURLs {
		result[k] = v
	}
	return result
}