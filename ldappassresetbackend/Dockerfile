# 使用 Golang 官方镜像
FROM golang:1.21-alpine AS builder
#
# # 设置工作目录
WORKDIR /app
#
# # 将当前目录的内容复制到工作目录
COPY . .
#
ENV GOPROXY=https://goproxy.cn,direct
# # 下载依赖并编译 Go 项目
RUN go mod tidy
RUN go build -o main .
#
# # 运行阶段
FROM alpine:latest 
#
# # 安装必要的依赖（如 libc6-compat 等）
RUN apk --no-cache add ca-certificates
#
# # 设置工作目录
WORKDIR /root/
#
# # 从构建阶段复制二进制文件
COPY --from=builder /app .
#
# # 复制 Vue3 构建文件到 Go 项目目录下
COPY ./public /app/public
#
# COPY ./config/config.yaml /app/config
#
# # 暴露应用端口
EXPOSE 8000
#
# # 启动应用
CMD ["./main"]
# # 设置国内 Go 模块代理