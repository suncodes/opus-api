# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 配置 Go 代理加速下载（使用国内镜像）
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV GOSUMDB=off

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Run stage
FROM python:3.9-slim

WORKDIR /app

# Install any system dependencies if needed
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy binary from builder
COPY --from=builder /app/server .

# Copy Python startup script
COPY app.py .

# Create logs directory
RUN mkdir -p /app/logs

# Hugging Face Spaces uses port 7860
EXPOSE 7860

# Set environment variable
ENV PORT=7860

# Run server via Python script
CMD ["python", "app.py"]