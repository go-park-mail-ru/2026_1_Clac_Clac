#!/bin/bash

# Вспомогательные функции
info() { echo -e "\033[1;36m[INFO]\033[0m $1"; }
ok() { echo -e "\033[1;32m[OK]\033[0m $1"; }
error() { echo -e "\033[1;31m[ERR]\033[0m $1"; }

# Настройки
# dry run для тестов
DRY_RUN=${DRY_RUN:-true}
DEPLOYMENT_DIR="deployments"
SHARED_PATTERNS="pkg/"
SERVICES=("appeal" "authorization" "board" "facade" "mail_sender" "rate_limiter" "user")
DOCKER_USER=${DOCKER_USER:-"nisakoo"}

# Логика

# $1 -- Таргет для билдинга
build_service() {
    local svc=$1
    local tag=${IMAGE_TAG:-"latest"}
    if [ "$DRY_RUN" = true ]; then
        info "$svc build skip (tag: $tag)"
    else
        docker build --platform linux/amd64 -f "$DEPLOYMENT_DIR/$svc/Dockerfile" -t "$DOCKER_USER/nexus-$svc:$tag" . || { error "build failed for $svc"; exit 1; }
        docker push "$DOCKER_USER/nexus-$svc:$tag" || { error "push failed for $svc"; exit 1; }

        # Обновляем values.yaml для ArgoCD
        if [ -f "deployments/nexus/values.yaml" ]; then
            # Экранируем слеши в имени пользователя для sed
            local escaped_user=$(echo $DOCKER_USER | sed 's/\//\\\//g')
            sed -i "s|image: $escaped_user/nexus-$svc:.*|image: $DOCKER_USER/nexus-$svc:$tag|g" deployments/nexus/values.yaml
            info "updated values.yaml for $svc with tag $tag"
        fi
    fi

    ok "done $svc (tag: $tag)"
}

# Основной код
[ "$DRY_RUN" = true ] && error "--- DRY RUN MODE ENABLED ---"

# Первый параметр -- относительно чего делаем git diff
BASE_DIFF_REF="$1"
# Защита от нулей GitHub
if [[ -z "$BASE_DIFF_REF" || "$BASE_DIFF_REF" == "0000000000000000000000000000000000000000" ]]; then
    BASE_DIFF_REF="HEAD~1"
fi

# Что реально изменилось
CHANGED_FILES=$(git diff --name-only "$BASE_DIFF_REF")

info "checking diff from: $BASE_DIFF_REF"

if [ -z "$CHANGED_FILES" ]; then
    info "no files changed since $BASE_DIFF_REF"
    exit 0
fi

# Проверка и сборка
if echo "$CHANGED_FILES" | grep -qE "$SHARED_PATTERNS"; then
    info "shared files changed, full rebuild..."
    for svc in "${SERVICES[@]}"; do
        build_service "$svc"
    done
else
    for svc in "${SERVICES[@]}"; do
        if echo "$CHANGED_FILES" | grep -qE "^$DEPLOYMENT_DIR/$svc/|^$svc/"; then
            build_service "$svc"
        fi
    done
fi

ok "all done"
