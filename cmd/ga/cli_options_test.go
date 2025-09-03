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

package main

import (
	"strings"
	"testing"
)

func TestParseArgs_FormatOption(t *testing.T) {
	app := NewCLIApp()

	testCases := []struct {
		name        string
		args        []string
		expected    string
		shouldError bool
	}{
		{
			name:        "Default format (csv)",
			args:        []string{"--config", "test.yaml"},
			expected:    "csv",
			shouldError: false,
		},
		{
			name:        "CSV format",
			args:        []string{"--config", "test.yaml", "--format", "csv"},
			expected:    "csv",
			shouldError: false,
		},
		{
			name:        "JSON format",
			args:        []string{"--config", "test.yaml", "--format", "json"},
			expected:    "json",
			shouldError: false,
		},
		{
			name:        "CSV format (uppercase)",
			args:        []string{"--config", "test.yaml", "--format", "CSV"},
			expected:    "",
			shouldError: true,
		},
		{
			name:        "JSON format (uppercase)",
			args:        []string{"--config", "test.yaml", "--format", "JSON"},
			expected:    "",
			shouldError: true,
		},
		{
			name:        "Invalid format",
			args:        []string{"--config", "test.yaml", "--format", "xml"},
			expected:    "",
			shouldError: true,
		},
		{
			name:        "Invalid format (empty)",
			args:        []string{"--config", "test.yaml", "--format", ""},
			expected:    "",
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options, err := app.parseArgs(tc.args)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for args %v, but got none", tc.args)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for args %v: %v", tc.args, err)
				return
			}

			if options.OutputFormat != tc.expected {
				t.Errorf("Expected format %s, got %s", tc.expected, options.OutputFormat)
			}
		})
	}
}

func TestParseArgs_OutputOption(t *testing.T) {
	app := NewCLIApp()

	testCases := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "Default output (empty)",
			args:     []string{"--config", "test.yaml"},
			expected: "",
		},
		{
			name:     "CSV output file",
			args:     []string{"--config", "test.yaml", "--output", "data.csv"},
			expected: "data.csv",
		},
		{
			name:     "JSON output file",
			args:     []string{"--config", "test.yaml", "--output", "data.json"},
			expected: "data.json",
		},
		{
			name:     "Output to stdout (-)",
			args:     []string{"--config", "test.yaml", "--output", "-"},
			expected: "-",
		},
		{
			name:     "Output with path",
			args:     []string{"--config", "test.yaml", "--output", "/tmp/output.csv"},
			expected: "/tmp/output.csv",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options, err := app.parseArgs(tc.args)

			if err != nil {
				t.Errorf("Unexpected error for args %v: %v", tc.args, err)
				return
			}

			if options.OutputPath != tc.expected {
				t.Errorf("Expected output path %s, got %s", tc.expected, options.OutputPath)
			}
		})
	}
}

func TestParseArgs_CombinedOptions(t *testing.T) {
	app := NewCLIApp()

	testCases := []struct {
		name           string
		args           []string
		expectedFormat string
		expectedOutput string
		expectedConfig string
		shouldError    bool
	}{
		{
			name:           "CSV format with output file",
			args:           []string{"--config", "custom.yaml", "--format", "csv", "--output", "data.csv"},
			expectedFormat: "csv",
			expectedOutput: "data.csv",
			expectedConfig: "custom.yaml",
			shouldError:    false,
		},
		{
			name:           "JSON format with output file",
			args:           []string{"--config", "custom.yaml", "--format", "json", "--output", "data.json"},
			expectedFormat: "json",
			expectedOutput: "data.json",
			expectedConfig: "custom.yaml",
			shouldError:    false,
		},
		{
			name:           "JSON format to stdout",
			args:           []string{"--config", "test.yaml", "--format", "json", "--output", "-"},
			expectedFormat: "json",
			expectedOutput: "-",
			expectedConfig: "test.yaml",
			shouldError:    false,
		},
		{
			name:           "All options with debug",
			args:           []string{"--config", "test.yaml", "--format", "json", "--output", "out.json", "--debug"},
			expectedFormat: "json",
			expectedOutput: "out.json",
			expectedConfig: "test.yaml",
			shouldError:    false,
		},
		{
			name:           "Invalid format with valid output",
			args:           []string{"--config", "test.yaml", "--format", "xml", "--output", "data.xml"},
			expectedFormat: "",
			expectedOutput: "",
			expectedConfig: "",
			shouldError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options, err := app.parseArgs(tc.args)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for args %v, but got none", tc.args)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for args %v: %v", tc.args, err)
				return
			}

			if options.OutputFormat != tc.expectedFormat {
				t.Errorf("Expected format %s, got %s", tc.expectedFormat, options.OutputFormat)
			}

			if options.OutputPath != tc.expectedOutput {
				t.Errorf("Expected output path %s, got %s", tc.expectedOutput, options.OutputPath)
			}

			if options.ConfigPath != tc.expectedConfig {
				t.Errorf("Expected config path %s, got %s", tc.expectedConfig, options.ConfigPath)
			}
		})
	}
}

func TestParseArgs_ErrorMessages(t *testing.T) {
	app := NewCLIApp()

	testCases := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name:          "Invalid format error message",
			args:          []string{"--config", "test.yaml", "--format", "xml"},
			expectedError: "無効な出力形式です: xml (csv または json を指定してください)",
		},
		{
			name:          "Empty format error message",
			args:          []string{"--config", "test.yaml", "--format", ""},
			expectedError: "無効な出力形式です:  (csv または json を指定してください)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := app.parseArgs(tc.args)

			if err == nil {
				t.Errorf("Expected error for args %v, but got none", tc.args)
				return
			}

			if !strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf("Expected error containing '%s', got: %v", tc.expectedError, err)
			}
		})
	}
}

func TestCLIOptions_DefaultValues(t *testing.T) {
	app := NewCLIApp()

	// 最小限の引数でテスト
	options, err := app.parseArgs([]string{})
	if err != nil {
		t.Fatalf("Unexpected error with minimal args: %v", err)
	}

	// デフォルト値の確認
	if options.ConfigPath != "ga.yaml" {
		t.Errorf("Expected default config path 'ga.yaml', got %s", options.ConfigPath)
	}

	if options.OutputPath != "" {
		t.Errorf("Expected default output path '', got %s", options.OutputPath)
	}

	if options.OutputFormat != "csv" {
		t.Errorf("Expected default format 'csv', got %s", options.OutputFormat)
	}

	if options.Debug != false {
		t.Errorf("Expected default debug false, got %v", options.Debug)
	}

	if options.Help != false {
		t.Errorf("Expected default help false, got %v", options.Help)
	}

	if options.Version != false {
		t.Errorf("Expected default version false, got %v", options.Version)
	}

	if options.Login != false {
		t.Errorf("Expected default login false, got %v", options.Login)
	}
}
