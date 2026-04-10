# 使用已有的 golang:1.25.5-alpine
FROM golang:1.25.5-alpine

# 设置工作目录
WORKDIR /app

# 设置 Go 代理（国内加速）
ENV GOPROXY=https://goproxy.cn,direct

# 复制依赖文件并下载（利用缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .


# 设置时区
ENV TZ=Asia/Shanghai

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/main .

# 暴露端口（根据你的 Go 项目实际端口修改）
EXPOSE 8080

# 启动程序
CMD ["./main"]