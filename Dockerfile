FROM golang:1.21-alpine AS builder

# 安装编译依赖
RUN apk --no-cache add gcc musl-dev sqlite-dev

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates sqlite curl

# 创建应用目录
WORKDIR /root/

# 从builder阶段复制二进制文件
COPY --from=builder /app/main .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 复制web静态文件
COPY --from=builder /app/web ./web

# 创建数据目录
RUN mkdir -p data

# 暴露端口
EXPOSE 8083

# 启动应用
CMD ["./main"] 