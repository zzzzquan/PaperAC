# Build Stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 设置 Go Proxy 加速国内构建
ENV GOPROXY=https://goproxy.cn,direct

# 安装基础编译依赖
RUN apk add --no-cache git

# 缓存依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并构建
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Run Stage
FROM alpine:latest

WORKDIR /app

# 安装必要的运行时依赖 (证书、时区)
RUN apk --no-cache add ca-certificates tzdata

# 从构建层复制二进制文件
COPY --from=builder /app/main .
# 复制 Assets (如字体)
COPY --from=builder /app/assets ./assets
# 复制可能的 .env (生产环境建议通过 docker-compose environment 注入，这里可选)

# 暴露端口
EXPOSE 8080

# 启动
CMD ["./main"]
