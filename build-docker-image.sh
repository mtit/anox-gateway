#!/usr/bin/env sh
set -eu

docker build --build-arg GOPROXY=https://goproxy.cn,direct -t mtit/anox-gateway:latest .

echo ""
echo "==> Build complete: mtit/anox-gateway:latest"
echo ""
echo "启动示例:"
echo ""
cat <<'EXAMPLE'
# 确保网络存在
docker network inspect dev-net >/dev/null 2>&1 || docker network create dev-net

# 首次或更新后启动
docker rm -f anox-gateway 2>/dev/null || true

docker run -d \
  --name anox-gateway \
  --restart unless-stopped \
  --network dev-net \
  -p 8080:8080 \
  -e ANOX_URL=anox-server:8848 \
  -e HTTP_HOST=0.0.0.0 \
  -e HTTP_PORT=8080 \
  -e JWT_SECRET=change-me \
  mtit/anox-gateway:latest
EXAMPLE
