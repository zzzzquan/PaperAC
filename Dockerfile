# Stage 1: Front-end Build
FROM node:18-alpine AS frontend-builder
WORKDIR /app/web
# Coping web directory
COPY web/package*.json ./
RUN npm install
COPY web/ .
RUN npm run build

# Stage 2: Back-end Build
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app/server
ENV GOPROXY=https://goproxy.cn,direct

# Install git for dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY server/go.mod server/go.sum ./
RUN go mod download

# Copy source code
COPY server/ .
# Build static binary with CGO_ENABLED=0
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Stage 3: Final Image
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create data directory
RUN mkdir -p /data/tmp

# Copy Backend Binary
COPY --from=backend-builder /app/server/main .

# Copy Frontend Dist
# Note: Vite output is usually in dist/
COPY --from=frontend-builder /app/web/dist ./dist

# Expose /data volume
VOLUME ["/data"]

# Expose Port
EXPOSE 8080

# Run
CMD ["./main"]
