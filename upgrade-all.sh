#!/bin/bash
set -euo pipefail

WORKSPACE="/Users/pangxingzhong/workspace/subs-check-all"
REPO="$WORKSPACE/subs-check"
BASE_DIR="/Users/pangxingzhong/Library/Mobile Documents/com~apple~CloudDocs/docker/subcheck"
IMAGE_NAME="subs-check:local"
PROJECTS=(mianfei-subscheck zhuque-subscheck fufei-subscheck test-subcheck)

echo "=== 使用本地仓构建 $IMAGE_NAME ==="
(cd "$REPO" && docker build --pull --no-cache -t "$IMAGE_NAME" .)

for project in "${PROJECTS[@]}"; do
  echo "=== 重启 $project ==="
  (cd "$BASE_DIR/$project" && docker compose up -d --remove-orphans)
done

echo "=== 验证容器状态 ==="
docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}"
