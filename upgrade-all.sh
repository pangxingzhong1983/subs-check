#!/bin/bash
set -euo pipefail

WORKSPACE="/Users/pangxingzhong/workspace/subs-check-all"
REPO="$WORKSPACE/subs-check"
BASE_DIR="/Users/pangxingzhong/Library/Mobile Documents/com~apple~CloudDocs/docker/subcheck"
IMAGE_NAME="subs-check:local"
PROJECTS=(mianfei-subscheck zhuque-subscheck fufei-subscheck test-subcheck)

echo "=== 使用代码目录: $REPO ==="
echo "=== 当前 HEAD: $(git -C "$REPO" rev-parse --short HEAD 2>/dev/null || echo unknown) ==="
echo "=== 工作区改动: ==="
git -C "$REPO" status --short || true

echo "=== 停止现有容器 ==="
for project in "${PROJECTS[@]}"; do
  (cd "$BASE_DIR/$project" && docker compose down) || true
done

echo "=== 清理旧镜像 ==="
docker image rm -f "$IMAGE_NAME" || true

echo "=== 清空 output 目录 ==="
for project in "${PROJECTS[@]}"; do
  rm -rf "$BASE_DIR/$project/output/"* || true
done

echo "=== 本地构建镜像 $IMAGE_NAME ==="
COMMIT=$(git -C "$REPO" rev-parse HEAD 2>/dev/null || echo unknown)
VERSION=$(date +%Y%m%d%H%M%S)
(cd "$REPO" && docker build --pull --no-cache \
  --build-arg GITHUB_SHA="$COMMIT" \
  --build-arg VERSION="$VERSION" \
  -t "$IMAGE_NAME" .)

for project in "${PROJECTS[@]}"; do
  echo "=== 重启 $project ==="
  (cd "$BASE_DIR/$project" && docker compose up -d --remove-orphans)
done

echo "=== 验证容器状态 ==="
docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}"
