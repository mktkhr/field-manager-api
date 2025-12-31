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
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /import-processor ./cmd/import-processor

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /import-processor /app/import-processor

# 非rootユーザーで実行
RUN adduser -D -g '' appuser
USER appuser

ENTRYPOINT ["/app/import-processor"]
