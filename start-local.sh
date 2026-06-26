#!/bin/bash

# “工作伙伴人工智能”一键启动脚本 (极致优化版)
# 适合中小学生与初学者使用，支持自动安装依赖、自动创建数据库

echo "🚀 工作伙伴人工智能 启动助手正在初始化..."

# 1. 检查基础环境是否已安装
command -v go >/dev/null 2>&1 || { echo "❌ 错误: 未安装 编程语言环境。请前往 https://go.dev 安装。"; exit 1; }
command -v node >/dev/null 2>&1 || { echo "❌ 错误: 未安装 网页运行环境。请前往 https://nodejs.org 安装。"; exit 1; }
command -v npm >/dev/null 2>&1 || { echo "❌ 错误: 未安装 软件包管理器。"; exit 1; }

# 2. 检查并生成 环境变量文件
if [ ! -f .env ]; then
    echo "📝 正在根据模板生成 环境变量 配置文件..."
    cp .env.example .env
    # 自动替换为本地开发最简配置
    sed -i '' 's/APP_PROFILE=docker/APP_PROFILE=local/g' .env
    sed -i '' 's/DB_HOST=mysql/DB_HOST=127.0.0.1/g' .env
    sed -i '' 's/MILVUS_HOST=milvus-standalone/MILVUS_HOST=127.0.0.1/g' .env
    sed -i '' 's/SERVER_PORT=8080/SERVER_PORT=18080/g' .env
    sed -i '' 's/NEXT_PUBLIC_BACKEND_PORT=8080/NEXT_PUBLIC_BACKEND_PORT=18080/g' .env
    echo "✅ 环境变量文件已生成。请确保您的本地 数据库 (地址 127.0.0.1 端口 3306) 已启动。"
fi

# 3. 检查 数据库 是否可连接
echo "🔍 正在检查 数据库 连接状态..."
# 使用 网络检查工具 探测 3306 端口
nc -z -w 2 127.0.0.1 3306 >/dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "⚠️ 警告: 无法连接到本地 3306 端口。请确认 数据库 服务已启动。"
    echo "如果 数据库 服务未运行，后端程序将无法成功启动。"
else
    echo "✅ 数据库 端口响应正常。"
fi

# 4. 后端服务准备
echo "📡 正在同步后端依赖库..."
cd backend
go mod tidy
echo "✅ 后端依赖准备就绪。"

# 启动后端服务（在后台运行并记录日志）
echo "🎬 正在启动后端服务 (监听端口 18080)..."
go run main.go > ../后端运行日志.log 2>&1 &
BACKEND_PID=$!
cd ..

# 5. 前端服务准备
echo "🌐 正在检查前端网页运行环境..."
cd frontend
if [ ! -d "node_modules" ]; then
    echo "📦 正在安装网页依赖组件 (预计耗时数分钟)..."
    npm install
fi

# 6. 启动前端服务
echo "🚀 正在启动前端服务 (访问端口 3000)..."
echo "💡 温馨提示：如果浏览器打开后显示 404，请稍等片刻并刷新页面，等待后端完全启动。"
npm run dev

# 脚本退出时自动关闭后端进程
trap "kill $BACKEND_PID" EXIT
