## 🐳 Docker 開発メモ

### 起動（基本：バックグラウンド）

```bash
# next-go-monorepo（docker-compose.yml がある場所）で
docker compose up -d --build
```

```bash
docker compose logs -f
```

```bash
docker compose down
```

## ▶️ Go 単体で起動（Docker を使わない）

### 1) DB だけ起動（docker-compose）

```bash
# リポジトリルート（docker-compose.yml がある場所）で
docker compose up -d db
```

### 2) API をローカルで起動（go run）

```bash
# apps/go-api（main.go がある場所）で
export DATABASE_URL='postgres://nextgo:nextgo@localhost:5432/nextgo_dev?sslmode=disable'
export JWT_SECRET='dev-secret-change-me'

go run .
```

### 3) 動作確認

```bash
# 別ターミナルで（場所はどこでもOK）
curl -sS http://localhost:8080/health
```

### 4) 片付け（停止）

```bash
# リポジトリルート（docker-compose.yml がある場所）で
docker compose down
```
