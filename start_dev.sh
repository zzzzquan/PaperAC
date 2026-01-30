#!/bin/bash

# PaperAC 本地开发启动脚本
# 同时启动后端 API 和前端 Vite 服务

# 捕获退出信号，关闭子进程
trap 'kill $(jobs -p)' EXIT

echo "🚀 正在启动 PaperAC 开发环境..."

# 确保进入脚本所在目录
cd "$(dirname "$0")" || exit 1


# 1. 启动后端
echo "📦 [Backend] 启动中..."
cd server
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi
# 使用 mock 模式确保验证码打印到控制台
export MAILER_MODE="mock"
go run ./cmd/api/main.go &
BACKEND_PID=$!
cd ..

# 等待一秒让后端初始化
sleep 2

# 2. 启动前端
echo "🎨 [Frontend] 启动中..."
cd web
npm run dev &
FRONTEND_PID=$!
cd ..

echo "✅ 服务已启动！"
echo "   后端日志将直接显示在下方 (验证码在这里看)"
echo "   按 Ctrl+C 停止所有服务"
echo "---------------------------------------------------"

# 等待子进程
wait
