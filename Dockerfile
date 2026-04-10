# 多阶段构建 Dockerfile
# Stage 1: 构建阶段
FROM golang:1.24-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum（如果存在）
COPY go.mod go.sum* ./

# 下载依赖（利用 Docker 缓存层）
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
# -ldflags="-s -w" 去除调试信息，减小二进制文件大小
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o idle_rpg .


# Stage 2: 运行阶段
FROM alpine:latest

# 安装 ca-certificates 用于 HTTPS 连接
RUN apk --no-cache add ca-certificates tzdata

# 设置时区（可选）
ENV TZ=Asia/Shanghai

# 创建非 root 用户运行
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/idle_rpg .

# 修改文件权限
RUN chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 健康检查（可选）
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD pgrep idle_rpg || exit 1

# 暴露端口（如果未来需要添加 Web 接口）
# EXPOSE 8080

# 设置环境变量默认值
ENV MYSQL_HOST=mysql
ENV MYSQL_PORT=3306
ENV MYSQL_USER=rpguser
ENV MYSQL_PASSWORD=
ENV MYSQL_DATABASE=idle_rpg
ENV REDIS_ADDR=redis:6379
ENV REDIS_PASSWORD=
ENV REDIS_DB=0

# 启动应用
ENTRYPOINT ["./idle_rpg"]
