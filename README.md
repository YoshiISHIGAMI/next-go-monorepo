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
curl -sS http://localhost:8080/ping
```

### 4) 片付け（停止）

```bash
# リポジトリルート（docker-compose.yml がある場所）で
docker compose down
```

## Railway Deploy Memo (Go API + Postgres)

### 1) Create Project

- Railway: New Project → Empty Project

### 2) Add Postgres

- `+ Create` → Database → PostgreSQL
- Confirm Postgres is Online

### 3) Initialize DB (run init.sql)

- Postgres → Connect → Public Network → copy Connection URL (show)
- Run from repo root:

```bash
docker run --rm -it \
  -v "$PWD:/work" -w /work \
  postgres:17 \
  psql "<CONNECTION_URL>" -f docker/db/init.sql
```

### 4) Add API Service (monorepo)

- Create → Service → GitHub Repo
- Root Directory: /apps/go-api
- Build Method: Dockerfile
- Dockerfile Path: /apps/go-api/Dockerfile

### 5) Set Environment Variables (API service → Variables)

- `DATABASE_URL` = `${{ Postgres.DATABASE_URL }}`
- `JWT_SECRET` = random 32+ chars

### 6) Deploy & Generate Domain

- Deployments → Deploy
- Settings → Public Networking → Generate Domain (Port: 8080)

### 7) Smoke Test (Production URL)

```bash
BASE_URL="https://<your-domain>"

# ping
curl -i "$BASE_URL/ping"

# signup (dummy)
curl -i -X POST "$BASE_URL/auth/signup" \
  -H "Content-Type: application/json" \
  -d '{"email":"dummy@example.com","password":"dummyPass1234"}'

# login -> token -> me (one-liner)
BASE_URL="https://<your-domain>"; TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" -H "Content-Type: application/json" -d '{"email":"dummy@example.com","password":"dummyPass1234"}' | python -c 'import sys,json; print(json.load(sys.stdin)["token"])'); curl -i "$BASE_URL/auth/me" -H "Authorization: Bearer $TOKEN"
```
