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

# Google Analytics 4 データ取得ツール - 使用例スクリプト
# このスクリプトは、gaツールの様々な使用方法を示します

set -e

# 色付きの出力用
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ヘルプ関数
show_help() {
    echo -e "${BLUE}Google Analytics 4 データ取得ツール - 使用例${NC}"
    echo ""
    echo "使用方法: $0 [オプション]"
    echo ""
    echo "オプション:"
    echo "  help          このヘルプを表示"
    echo "  basic         基本的な使用例を実行"
    echo "  advanced      高度な使用例を実行"
    echo "  batch         バッチ処理の例を実行"
    echo "  all           すべての例を実行"
    echo ""
}

# 基本的な使用例
run_basic_examples() {
    echo -e "${GREEN}=== 基本的な使用例 ===${NC}"

    echo -e "${YELLOW}1. ヘルプの表示${NC}"
    ./ga --help
    echo ""

    echo -e "${YELLOW}2. バージョン情報の表示${NC}"
    ./ga --version
    echo ""

    echo -e "${YELLOW}3. デフォルト設定でのデータ取得（標準出力）${NC}"
    if [ -f "ga.yaml" ]; then
        echo "設定ファイル ga.yaml を使用してデータを取得します..."
        ./ga
    else
        echo -e "${RED}エラー: ga.yaml が見つかりません。ga.yaml.example をコピーして設定してください。${NC}"
    fi
    echo ""
}

# 高度な使用例
run_advanced_examples() {
    echo -e "${GREEN}=== 高度な使用例 ===${NC}"

    echo -e "${YELLOW}1. カスタム設定ファイルを使用${NC}"
    if [ -f "ga.yaml.example" ]; then
        cp ga.yaml.example custom_config.yaml
        echo "カスタム設定ファイル custom_config.yaml を作成しました"
        ./ga --config custom_config.yaml --output custom_data.csv
        echo "データを custom_data.csv に出力しました"
        rm -f custom_config.yaml custom_data.csv
    fi
    echo ""

    echo -e "${YELLOW}2. デバッグモードでの実行${NC}"
    if [ -f "ga.yaml" ]; then
        ./ga --debug --output debug_output.csv
        echo "デバッグ情報付きでデータを取得し、debug_output.csv に出力しました"
        rm -f debug_output.csv
    fi
    echo ""

    echo -e "${YELLOW}3. パイプを使用した処理${NC}"
    if [ -f "ga.yaml" ]; then
        echo "最初の10行を表示:"
        ./ga | head -10
        echo ""
        echo "セッション数でソート（降順）:"
        ./ga | sort -t',' -k4 -nr | head -5
    fi
    echo ""
}

# バッチ処理の例
run_batch_examples() {
    echo -e "${GREEN}=== バッチ処理の例 ===${NC}"

    echo -e "${YELLOW}1. 月次レポートの生成${NC}"

    # 現在の年月を取得
    current_year=$(date +%Y)
    current_month=$(date +%m)

    # 前月の計算
    if [ "$current_month" -eq 1 ]; then
        prev_month=12
        prev_year=$((current_year - 1))
    else
        prev_month=$((current_month - 1))
        prev_year=$current_year
    fi

    # 前月の開始日と終了日
    start_date=$(printf "%04d-%02d-01" $prev_year $prev_month)

    # 月末日の計算
    if [ "$prev_month" -eq 2 ]; then
        # うるう年の判定
        if [ $((prev_year % 4)) -eq 0 ] && ([ $((prev_year % 100)) -ne 0 ] || [ $((prev_year % 400)) -eq 0 ]); then
            end_day=29
        else
            end_day=28
        fi
    elif [ "$prev_month" -eq 4 ] || [ "$prev_month" -eq 6 ] || [ "$prev_month" -eq 9 ] || [ "$prev_month" -eq 11 ]; then
        end_day=30
    else
        end_day=31
    fi

    end_date=$(printf "%04d-%02d-%02d" $prev_year $prev_month $end_day)

    echo "前月（${start_date} ～ ${end_date}）のレポートを生成します..."

    # 一時的な設定ファイルを作成
    if [ -f "ga.yaml.example" ]; then
        sed -e "s/start_date: \".*\"/start_date: \"$start_date\"/" \
            -e "s/end_date: \".*\"/end_date: \"$end_date\"/" \
            ga.yaml.example > monthly_config.yaml

        output_file="monthly_report_${prev_year}_$(printf "%02d" $prev_month).csv"

        echo "設定ファイル: monthly_config.yaml"
        echo "出力ファイル: $output_file"

        # 実際にはコメントアウト（認証が必要なため）
        # ./ga --config monthly_config.yaml --output "$output_file"
        echo "（実際の実行はコメントアウトされています）"

        rm -f monthly_config.yaml
    fi
    echo ""

    echo -e "${YELLOW}2. 複数期間の比較レポート${NC}"
    echo "今月と前月のデータを比較するスクリプト例:"

    cat << 'EOF'
#!/bin/bash
# 比較レポート生成スクリプト

# 今月のデータ取得
./ga --config config_current_month.yaml --output current_month.csv

# 前月のデータ取得
./ga --config config_previous_month.yaml --output previous_month.csv

# 比較分析（例：Pythonスクリプトを呼び出し）
# python3 compare_reports.py current_month.csv previous_month.csv
EOF
    echo ""
}

# 環境チェック
check_environment() {
    echo -e "${BLUE}環境チェック中...${NC}"

    # gaコマンドの存在確認
    if [ ! -f "./ga" ]; then
        echo -e "${RED}エラー: gaコマンドが見つかりません。先にビルドしてください:${NC}"
        echo "  go build -o ga cmd/ga/main.go"
        exit 1
    fi

    # 環境変数の確認
    if [ -z "$GA_CLIENT_ID" ] || [ -z "$GA_CLIENT_SECRET" ]; then
        echo -e "${YELLOW}警告: OAuth認証用の環境変数が設定されていません${NC}"
        echo "以下の環境変数を設定してください:"
        echo "  export GA_CLIENT_ID=\"your-client-id\""
        echo "  export GA_CLIENT_SECRET=\"your-client-secret\""
        echo ""
    fi

    # 設定ファイルの確認
    if [ ! -f "ga.yaml" ] && [ ! -f "ga.yaml.example" ]; then
        echo -e "${RED}エラー: 設定ファイルが見つかりません${NC}"
        exit 1
    fi

    echo -e "${GREEN}環境チェック完了${NC}"
    echo ""
}

# メイン処理
main() {
    case "${1:-help}" in
        "help")
            show_help
            ;;
        "basic")
            check_environment
            run_basic_examples
            ;;
        "advanced")
            check_environment
            run_advanced_examples
            ;;
        "batch")
            check_environment
            run_batch_examples
            ;;
        "all")
            check_environment
            run_basic_examples
            run_advanced_examples
            run_batch_examples
            ;;
        *)
            echo -e "${RED}不明なオプション: $1${NC}"
            show_help
            exit 1
            ;;
    esac
}

# スクリプト実行
main "$@"