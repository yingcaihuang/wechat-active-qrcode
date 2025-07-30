#!/bin/bash

# 微信活码管理系统 Docker 部署脚本

echo "=== 微信活码管理系统 Docker 部署 ==="

# 读取BASE_URL参数
BASE_URL=${1:-"http://localhost:8083"}
echo "🌐 使用BASE_URL: $BASE_URL"

# 检查Docker是否运行
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

echo "✅ Docker运行正常"

# 导出环境变量
export BASE_URL=$BASE_URL

# 停止现有容器
echo "🛑 停止现有容器..."
docker-compose -f docker-compose.debian.yml down 2>/dev/null || true

# 删除旧镜像
echo "🗑️  清理旧镜像..."
docker-compose -f docker-compose.debian.yml down --rmi all 2>/dev/null || true

# 构建新镜像
echo "🔨 构建新镜像..."
docker-compose -f docker-compose.debian.yml build --no-cache

if [ $? -ne 0 ]; then
    echo "❌ 镜像构建失败"
    exit 1
fi

echo "✅ 镜像构建成功"

# 启动服务
echo "🚀 启动服务..."
docker-compose -f docker-compose.debian.yml up -d

if [ $? -ne 0 ]; then
    echo "❌ 服务启动失败"
    exit 1
fi

echo "✅ 服务启动成功"

# 等待服务就绪
echo "⏳ 等待服务就绪..."
sleep 5

# 检查服务状态
echo "🔍 检查服务状态..."
docker-compose -f docker-compose.debian.yml ps

# 检查健康状态
echo "🏥 检查健康状态..."
for i in {1..10}; do
    if curl -f http://localhost:8083/health >/dev/null 2>&1; then
        echo "✅ 服务健康检查通过"
        break
    fi
    
    if [ $i -eq 10 ]; then
        echo "❌ 服务健康检查失败"
        echo "查看日志："
        docker-compose -f docker-compose.debian.yml logs --tail=20
        exit 1
    fi
    
    echo "等待服务启动... ($i/10)"
    sleep 3
done

echo ""
echo "🎉 部署完成！"
echo "📱 访问地址: http://localhost:8083"
echo "🌐 当前BASE_URL: $BASE_URL"
echo "📊 健康检查: http://localhost:8083/health"
echo ""
echo "📋 管理命令:"
echo "  查看日志: docker-compose -f docker-compose.debian.yml logs -f"
echo "  停止服务: docker-compose -f docker-compose.debian.yml down"
echo "  重启服务: docker-compose -f docker-compose.debian.yml restart"
echo ""
echo "💡 使用自定义域名部署:"
echo "  ./deploy-docker.sh http://your-domain.com"
echo ""
