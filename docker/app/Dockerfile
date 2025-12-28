# syntax=docker/dockerfile:1

FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# 依存関係のインストール
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ビルド
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /server \
    ./cmd/server/main.go

FROM gcr.io/distroless/static:nonroot

# ビルド成果物をコピー
COPY --from=builder /server /server

# ポート公開
EXPOSE 8080

# 非rootユーザーで実行
USER nonroot:nonroot

ENTRYPOINT ["/server"]
