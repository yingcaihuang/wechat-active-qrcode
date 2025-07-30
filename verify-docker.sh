#!/bin/bash

# 验证Docker镜像中的文件结构

echo "=== 验证Docker镜像中的静态资源 ==="

# 获取容器ID
CONTAINER_ID=$(docker-compose -f docker-compose.debian.yml ps -q wechat-qrcode)

if [ -z "$CONTAINER_ID" ]; then
    echo "❌ 容器未运行，请先执行 ./deploy-docker.sh"
    exit 1
fi

echo "📦 容器ID: $CONTAINER_ID"

# 检查web目录结构
echo "🔍 检查web目录结构..."
docker exec $CONTAINER_ID find /root/web -type f | head -20

echo ""
echo "🔍 检查assets目录..."
docker exec $CONTAINER_ID ls -la /root/web/assets/ 2>/dev/null || echo "❌ assets目录不存在"

echo ""
echo "🔍 检查CSS文件..."
docker exec $CONTAINER_ID ls -la /root/web/assets/css/ 2>/dev/null || echo "❌ CSS目录不存在"

echo ""
echo "🔍 检查JS文件..."
docker exec $CONTAINER_ID ls -la /root/web/assets/js/ 2>/dev/null || echo "❌ JS目录不存在"

echo ""
echo "🔍 检查字体文件..."
docker exec $CONTAINER_ID ls -la /root/web/assets/fonts/ 2>/dev/null || echo "❌ fonts目录不存在"

echo ""
echo "🔍 检查index.html中的引用..."
docker exec $CONTAINER_ID grep -n "assets/" /root/web/index.html || echo "❌ 未找到assets引用"

echo ""
echo "🌐 测试静态文件访问..."
echo "CSS文件: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:8083/assets/css/bootstrap.min.css)"
echo "JS文件: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:8083/assets/js/bootstrap.bundle.min.js)"
echo "字体文件: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:8083/assets/fonts/bootstrap-icons.woff2)"

echo ""
echo "✅ 验证完成"
