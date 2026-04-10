# 直接使用已有的基础镜像
FROM golang:1.25.5-alpine

# 设置工作目录
WORKDIR /app

# 设置 Go 代理（这行只是设置环境变量，不会访问外网）
ENV GOPROXY=https://goproxy.cn,direct

# 复制依赖文件并下载（利用缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8080

# 启动程序
CMD ["./main"]