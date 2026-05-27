#!/bin/bash

# Использовать в ТОЛЬКО крайних случаях!!!

echo -e "[DANGER] Использовать на свой страх и риск"
echo -e "[DANGER] Предварительно надо войти в docker аккаунт"

SERVICES=("appeal" "authorization" "board" "facade" "mail_sender" "rate_limiter" "user" "realtime")

for svc in "${SERVICES[@]}"; do
    docker build --platform linux/amd64 \
        -f "deployments/${svc}/Dockerfile" \
        -t "nisakoo/nexus-${svc}:latest" \
        .
done
