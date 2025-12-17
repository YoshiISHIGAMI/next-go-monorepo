# --- build stage ---
FROM golang:1.25 AS builder

WORKDIR /app

# 依存関係を先に入れてキャッシュを効かせる
COPY go.mod go.sum ./
RUN go mod download

# ソースをコピー
COPY . .

# ビルド（静的バイナリ寄り）
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# --- runtime stage ---
FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /app/server /server

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/server"]