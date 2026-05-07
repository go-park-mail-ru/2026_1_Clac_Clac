#!/bin/bash

# Вспомогательные функции
info() {
    echo -e "\033[1;36m[INFO]\033[0m $1"
}

ok() {
    echo -e "\033[1;32m[OK]\033[0m $1"
}

error() {
    echo -e "\033[1;31m[ERR]\033[0m $1"
}

# Настройки

# dry run для тестов
DRY_RUN=true

if [ "$DRY_RUN" = true ]; then
    error "DRY RUN"
fi

DEPLOYMENT_DIR="deployments"
SHARED_PATTERNS="pkg/"

SERVICES=("appeal" "authorization" "board" "facade" "mail_sender" "rate_limiter" "user")

# Логика

# $1 - Таргет для билдинга
build() {
    if [ "$DRY_RUN" = true ]; then
        info "rebuild $1"
    else
        ok "rebuild"
    fi
}

# $1 -- CHANGED_FILES
check_shared_dirs() {
    if echo "$1" | grep -qE "$SHARED_PATTERNS"; then
        info "shared changed, rebuild all"
        for svc in "${SERVICES[@]}"; do
            build "$svc"
        done
        exit 0
    fi
}

# Первый параметр -- относительно чего делаем git diff
if [ -n "$1" ]; then
    BASE_DIFF_REF="$1"
else
    error "no base diff provided"
    exit 1
fi

# Что реально изменилось
CHANGED_FILES=$(git diff --name-only $BASE_DIFF_REF)

info "checking diff from: $BASE_DIFF_REF"

# Проверка и сборка
check_shared_dirs "$CHANGED_FILES"
