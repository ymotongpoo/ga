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
	"testing"
)

func TestProcessPagePath(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		pagePath string
		expected string
	}{
		{
			name:     "絶対URL（https）はそのまま返す",
			baseURL:  "https://example.com",
			pagePath: "https://other.com/page",
			expected: "https://other.com/page",
		},
		{
			name:     "絶対URL（http）はそのまま返す",
			baseURL:  "https://example.com",
			pagePath: "http://other.com/page",
			expected: "http://other.com/page",
		},
		{
			name:     "ベースURLが空の場合はpagePathをそのまま返す",
			baseURL:  "",
			pagePath: "/home",
			expected: "/home",
		},
		{
			name:     "ベースURLが空白の場合はpagePathをそのまま返す",
			baseURL:  "   ",
			pagePath: "/home",
			expected: "/home",
		},
		{
			name:     "pagePathが空の場合はベースURLのみ返す",
			baseURL:  "https://example.com",
			pagePath: "",
			expected: "https://example.com",
		},
		{
			name:     "pagePathが空白の場合はベースURLのみ返す",
			baseURL:  "https://example.com",
			pagePath: "   ",
			expected: "https://example.com",
		},
		{
			name:     "通常のURL結合",
			baseURL:  "https://example.com",
			pagePath: "/home",
			expected: "https://example.com/home",
		},
		{
			name:     "ベースURLの末尾スラッシュを処理",
			baseURL:  "https://example.com/",
			pagePath: "/home",
			expected: "https://example.com/home",
		},
		{
			name:     "pagePathの先頭スラッシュなし",
			baseURL:  "https://example.com",
			pagePath: "home",
			expected: "https://example.com/home",
		},
		{
			name:     "両方にスラッシュがある場合の重複処理",
			baseURL:  "https://example.com/",
			pagePath: "/home",
			expected: "https://example.com/home",
		},
		{
			name:     "複雑なパス",
			baseURL:  "https://example.com/app",
			pagePath: "/users/profile",
			expected: "https://example.com/app/users/profile",
		},
		{
			name:     "ベースURLにパスが含まれる場合",
			baseURL:  "https://example.com/api/v1/",
			pagePath: "data",
			expected: "https://example.com/api/v1/data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProcessPagePath(tt.baseURL, tt.pagePath)
			if result != tt.expected {
				t.Errorf("ProcessPagePath(%q, %q) = %q, want %q", tt.baseURL, tt.pagePath, result, tt.expected)
			}
		})
	}
}

func TestURLProcessor(t *testing.T) {
	t.Run("NewURLProcessor", func(t *testing.T) {
		streamURLs := map[string]string{
			"stream1": "https://example.com",
			"stream2": "https://other.com",
		}

		processor := NewURLProcessor(streamURLs)
		if processor == nil {
			t.Fatal("NewURLProcessor() returned nil")
		}

		if len(processor.streamURLs) != 2 {
			t.Errorf("streamURLs length = %d, want 2", len(processor.streamURLs))
		}
	})

	t.Run("NewURLProcessor with nil map", func(t *testing.T) {
		processor := NewURLProcessor(nil)
		if processor == nil {
			t.Fatal("NewURLProcessor() returned nil")
		}

		if processor.streamURLs == nil {
			t.Error("streamURLs should not be nil")
		}
	})

	t.Run("ProcessPagePath with URLProcessor", func(t *testing.T) {
		streamURLs := map[string]string{
			"stream1": "https://example.com",
			"stream2": "https://other.com",
		}

		processor := NewURLProcessor(streamURLs)

		result := processor.ProcessPagePath("stream1", "/home")
		expected := "https://example.com/home"
		if result != expected {
			t.Errorf("ProcessPagePath() = %q, want %q", result, expected)
		}

		result = processor.ProcessPagePath("stream2", "/about")
		expected = "https://other.com/about"
		if result != expected {
			t.Errorf("ProcessPagePath() = %q, want %q", result, expected)
		}

		// 存在しないストリームID
		result = processor.ProcessPagePath("nonexistent", "/page")
		expected = "/page"
		if result != expected {
			t.Errorf("ProcessPagePath() = %q, want %q", result, expected)
		}
	})

	t.Run("SetStreamURL and GetStreamURL", func(t *testing.T) {
		processor := NewURLProcessor(nil)

		processor.SetStreamURL("stream1", "https://example.com")
		result := processor.GetStreamURL("stream1")
		expected := "https://example.com"
		if result != expected {
			t.Errorf("GetStreamURL() = %q, want %q", result, expected)
		}

		// 存在しないストリームID
		result = processor.GetStreamURL("nonexistent")
		if result != "" {
			t.Errorf("GetStreamURL() = %q, want empty string", result)
		}
	})

	t.Run("GetAllStreamURLs", func(t *testing.T) {
		streamURLs := map[string]string{
			"stream1": "https://example.com",
			"stream2": "https://other.com",
		}

		processor := NewURLProcessor(streamURLs)
		result := processor.GetAllStreamURLs()

		if len(result) != 2 {
			t.Errorf("GetAllStreamURLs() length = %d, want 2", len(result))
		}

		if result["stream1"] != "https://example.com" {
			t.Errorf("GetAllStreamURLs()[stream1] = %q, want %q", result["stream1"], "https://example.com")
		}

		// 返されたマップを変更しても元のマップに影響しないことを確認
		result["stream1"] = "modified"
		if processor.GetStreamURL("stream1") == "modified" {
			t.Error("GetAllStreamURLs() should return a copy, not the original map")
		}
	})
}