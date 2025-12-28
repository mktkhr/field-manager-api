#!/usr/bin/env bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }
print_info() { echo -e "📦 $1"; }

copy_env_file() {
  local source="$1"
  local dest="$2"

  if [ -f "$dest" ]; then
    print_warning "$dest は既に存在します (スキップ)"
    return
  fi

  if [ ! -f "$source" ]; then
    print_error "$source が見つかりません"
    exit 1
  fi

  cp "$source" "$dest"
  print_success "$dest を作成しました"
}

check_command() {
  command -v "$1" >/dev/null 2>&1
}

add_asdf_plugin() {
  local plugin="$1"

  if asdf plugin list 2>/dev/null | grep -q "^${plugin}$"; then
    print_warning "asdf plugin '$plugin' は既に追加されています (スキップ)"
  else
    asdf plugin add "$plugin"
    print_success "asdf plugin '$plugin' を追加しました"
  fi
}

install_go_tool() {
  local name="$1"
  local package="$2"

  if check_command "$name"; then
    print_warning "$name は既にインストールされています (スキップ)"
  else
    print_info "$name をインストール中..."
    go install "$package"
    print_success "$name をインストールしました"
  fi
}

install_npm_tool() {
  local name="$1"
  local package="$2"

  if check_command "$name"; then
    print_warning "$name は既にインストールされています (スキップ)"
  else
    print_info "$name をインストール中..."
    npm install -g "$package"
    print_success "$name をインストールしました"
  fi
}

setup_env_files() {
  echo ""
  print_info "環境ファイルを準備中..."
  echo ""

  copy_env_file "docker/.env.sample" "docker/.env"
}

setup_asdf() {
  echo ""
  print_info "asdf のセットアップ中..."
  echo ""

  if ! check_command asdf; then
    print_error "asdf がインストールされていません"
    echo "  インストール方法: https://asdf-vm.com/guide/getting-started.html"
    exit 1
  fi

  add_asdf_plugin "golang"
  add_asdf_plugin "nodejs"

  echo ""
  print_info "asdf install を実行中..."
  asdf install
  print_success "asdf install が完了しました"
}

setup_dev_tools() {
  echo ""
  print_info "開発ツールをインストール中..."
  echo ""

  # Go tools
  install_go_tool "oapi-codegen" "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest"
  install_go_tool "air" "github.com/air-verse/air@latest"
  install_go_tool "gosec" "github.com/securego/gosec/v2/cmd/gosec@latest"
  install_go_tool "sqlc" "github.com/sqlc-dev/sqlc/cmd/sqlc@latest"

  # golangci-lint
  if check_command golangci-lint; then
    print_warning "golangci-lint は既にインストールされています (スキップ)"
  else
    print_info "golangci-lint をインストール中..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin"
    print_success "golangci-lint をインストールしました"
  fi

  # npm tools
  install_npm_tool "redocly" "@redocly/cli"
}

main() {
  echo "======================================"
  echo "  field-manager-api 初期セットアップ"
  echo "======================================"

  setup_env_files
  setup_asdf
  setup_dev_tools

  echo ""
  echo "======================================"
  print_success "初期セットアップが完了しました"
  echo "======================================"
  echo ""
  echo "次のステップ:"
  echo "  1. Dockerコンテナを起動してください:"
  echo "     cd docker && docker compose up -d"
  echo ""
  echo "  2. アプリケーションを起動してください:"
  echo "     make run (または) make dev"
}

main "$@"
