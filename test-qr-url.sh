#!/bin/bash

# 测试二维码URL是否正确的脚本

echo "=== 测试二维码URL正确性 ==="

# 读取当前配置的BASE_URL
if [ -f ".env" ]; then
    source .env
fi

BASE_URL=${BASE_URL:-"http://localhost:8083"}
echo "🌐 当前BASE_URL: $BASE_URL"

# 测试健康检查
echo "🏥 测试健康检查..."
curl -s "$BASE_URL/health" > /dev/null
if [ $? -eq 0 ]; then
    echo "✅ 服务正常运行"
else
    echo "❌ 服务未运行，请先启动服务"
    exit 1
fi

echo ""
echo "📋 运行完整测试来验证二维码URL..."

# 编译并运行测试
if [ -f "active_qr_test.go" ]; then
    echo "🔧 编译测试程序..."
    go build -o qr_test active_qr_test.go
    
    if [ $? -eq 0 ]; then
        echo "🧪 运行测试..."
        ./qr_test
        
        # 清理
        rm -f qr_test
    else
        echo "❌ 编译失败"
        exit 1
    fi
else
    echo "❌ 测试文件不存在"
    exit 1
fi

echo ""
echo "💡 提示:"
echo "  - 检查生成的二维码是否包含正确的BASE_URL"
echo "  - 生成的二维码应该指向: $BASE_URL/r/[短码]"
echo "  - 复制功能应该复制相同的URL"
