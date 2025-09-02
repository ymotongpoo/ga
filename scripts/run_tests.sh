#!/bin/bash
# Copyright 2025 Yoshi Yamaguchi
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# テスト実行スクリプト
# このスクリプトは単体テスト、統合テスト、エンドツーエンドテストを実行し、
# カバレッジレポートを生成します。

set -e

# 色付きの出力用
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# プロジェクトルートディレクトリに移動
cd "$(dirname "$0")/.."

echo -e "${BLUE}=== Google Analytics Tool テスト実行 ===${NC}"
echo

# 依存関係の確認
echo -e "${YELLOW}依存関係を確認中...${NC}"
go mod tidy
go mod download
echo -e "${GREEN}✓ 依存関係の確認完了${NC}"
echo

# ビルドテスト
echo -e "${YELLOW}ビルドテストを実行中...${NC}"
if go build -o /tmp/ga_build_test ./cmd/ga; then
    echo -e "${GREEN}✓ ビルドテスト成功${NC}"
    rm -f /tmp/ga_build_test
else
    echo -e "${RED}✗ ビルドテスト失敗${NC}"
    exit 1
fi
echo

# 単体テストの実行
echo -e "${YELLOW}単体テストを実行中...${NC}"
echo "対象パッケージ:"
echo "  - internal/auth"
echo "  - internal/config"
echo "  - internal/analytics"
echo "  - internal/errors"
echo "  - internal/logger"
echo "  - internal/output"
echo "  - cmd/ga"
echo

# カバレッジファイルの準備
mkdir -p coverage
rm -f coverage/*.out

# 各パッケージの単体テストを実行
UNIT_TEST_PACKAGES=(
    "./internal/auth"
    "./internal/config"
    "./internal/analytics"
    "./internal/errors"
    "./internal/logger"
    "./internal/output"
    "./cmd/ga"
)

UNIT_TEST_FAILED=0

for package in "${UNIT_TEST_PACKAGES[@]}"; do
    package_name=$(basename "$package")
    echo -e "${BLUE}Testing $package...${NC}"

    if go test -v -race -coverprofile="coverage/${package_name}.out" "$package"; then
        echo -e "${GREEN}✓ $package テスト成功${NC}"
    else
        echo -e "${RED}✗ $package テスト失敗${NC}"
        UNIT_TEST_FAILED=1
    fi
    echo
done

if [ $UNIT_TEST_FAILED -eq 1 ]; then
    echo -e "${RED}単体テストで失敗があります${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 全ての単体テスト成功${NC}"
echo

# 統合テストの実行
echo -e "${YELLOW}統合テストを実行中...${NC}"
if go test -v -race -coverprofile="coverage/integration.out" ./tests -run "TestConfigServiceIntegration|TestAuthServiceIntegration|TestOutputServiceIntegration|TestLoggerIntegration|TestErrorHandlingIntegration|TestServiceInteraction|TestCompleteWorkflow"; then
    echo -e "${GREEN}✓ 統合テスト成功${NC}"
else
    echo -e "${RED}✗ 統合テスト失敗${NC}"
    exit 1
fi
echo

# エンドツーエンドテストの実行
echo -e "${YELLOW}エンドツーエンドテストを実行中...${NC}"
if go test -v -race ./tests -run "TestCLI_"; then
    echo -e "${GREEN}✓ エンドツーエンドテスト成功${NC}"
else
    echo -e "${RED}✗ エンドツーエンドテスト失敗${NC}"
    exit 1
fi
echo

# カバレッジレポートの生成
echo -e "${YELLOW}カバレッジレポートを生成中...${NC}"

# 全てのカバレッジファイルを結合
echo "mode: atomic" > coverage/total.out
for file in coverage/*.out; do
    if [ "$file" != "coverage/total.out" ] && [ -f "$file" ]; then
        tail -n +2 "$file" >> coverage/total.out
    fi
done

# カバレッジ統計を表示
if command -v go &> /dev/null; then
    echo -e "${BLUE}=== カバレッジ統計 ===${NC}"
    go tool cover -func=coverage/total.out | tail -1
    echo

    # HTMLレポートの生成
    if go tool cover -html=coverage/total.out -o coverage/coverage.html; then
        echo -e "${GREEN}✓ HTMLカバレッジレポートを生成しました: coverage/coverage.html${NC}"
    else
        echo -e "${YELLOW}⚠ HTMLカバレッジレポートの生成に失敗しました${NC}"
    fi
fi

# パッケージ別カバレッジの表示
echo -e "${BLUE}=== パッケージ別カバレッジ ===${NC}"
for package in "${UNIT_TEST_PACKAGES[@]}"; do
    package_name=$(basename "$package")
    if [ -f "coverage/${package_name}.out" ]; then
        coverage=$(go tool cover -func="coverage/${package_name}.out" | tail -1 | awk '{print $3}')
        echo -e "${package}: ${coverage}"
    fi
done
echo

# ベンチマークテストの実行（オプション）
if [ "$1" = "--bench" ]; then
    echo -e "${YELLOW}ベンチマークテストを実行中...${NC}"
    go test -bench=. -benchmem ./internal/...
    echo
fi

# テスト結果のサマリー
echo -e "${BLUE}=== テスト結果サマリー ===${NC}"
echo -e "${GREEN}✓ ビルドテスト: 成功${NC}"
echo -e "${GREEN}✓ 単体テスト: 成功${NC}"
echo -e "${GREEN}✓ 統合テスト: 成功${NC}"
echo -e "${GREEN}✓ エンドツーエンドテスト: 成功${NC}"
echo -e "${GREEN}✓ カバレッジレポート: 生成完了${NC}"
echo

# 推奨事項の表示
echo -e "${BLUE}=== 推奨事項 ===${NC}"
echo "1. カバレッジレポートを確認してください: coverage/coverage.html"
echo "2. 80%以上のカバレッジを目標にしてください"
echo "3. 失敗したテストがある場合は修正してください"
echo "4. 新しい機能を追加する際は対応するテストも追加してください"
echo

echo -e "${GREEN}🎉 全てのテストが正常に完了しました！${NC}"