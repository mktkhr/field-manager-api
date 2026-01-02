# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# 依存関係のキャッシュ
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# ビルド(h3-goがCGOを必要とするため有効化)
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /cluster-worker ./cmd/cluster-worker

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /cluster-worker /app/cluster-worker

# 非rootユーザーで実行
RUN adduser -D -g '' appuser
USER appuser

ENTRYPOINT ["/app/cluster-worker"]
