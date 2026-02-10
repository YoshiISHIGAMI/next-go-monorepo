# next-go-monorepo

Go API + Next.js のフルスタックモノレポテンプレート

## アーキテクチャ

```
┌─────────────────────────────────────────────────────────────┐
│                        Client                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Next.js (Frontend)                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  App Router │  │  Auth.js    │  │  Server Components  │  │
│  │  (Pages)    │  │  (OAuth)    │  │  (SSR/RSC)          │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│                         │ BFF Pattern                       │
└─────────────────────────┼───────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    Go API (Backend)                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Echo      │  │   slog      │  │   OpenAPI           │  │
│  │   (Router)  │  │   (Logger)  │  │   (API Spec)        │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────┼───────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    PostgreSQL                               │
└─────────────────────────────────────────────────────────────┘
```

## ディレクトリ構成

```
next-go-monorepo/
├── apps/
│   ├── go-api/                    # Go API
│   │   ├── internal/
│   │   │   ├── handler/           # HTTPハンドラー
│   │   │   ├── middleware/        # ミドルウェア (auth, logger, error)
│   │   │   ├── model/             # データモデル
│   │   │   ├── apperror/          # 共通エラー
│   │   │   └── logger/            # slog設定
│   │   ├── main.go                # エントリポイント
│   │   ├── openapi.yaml           # API仕様
│   │   ├── Dockerfile             # 本番用
│   │   └── Dockerfile.dev         # 開発用 (air)
│   │
│   └── next-app/                  # Next.js
│       └── src/
│           ├── app/               # App Router (ページ)
│           ├── components/ui/     # shadcn/ui
│           └── shared/
│               ├── api/           # OpenAPI生成型
│               ├── config/        # 環境変数
│               └── lib/           # ユーティリティ
│
├── docker/
│   └── db/init.sql                # DB初期化
│
├── .github/workflows/ci.yml       # CI設定
├── docker-compose.yml
├── Makefile
└── CLAUDE.md                      # 開発ガイド
```

## 技術スタック

| レイヤー | 技術 |
|---------|------|
| Frontend | Next.js 16, React 19, Tailwind CSS v4, shadcn/ui |
| Backend | Go 1.25, Echo, slog |
| Database | PostgreSQL 17 |
| Auth | Auth.js v5 (GitHub OAuth) |
| API Spec | OpenAPI 3.0 |
| Testing | Vitest, Playwright, go test |
| CI | GitHub Actions |
| Dev Tools | Biome, air (hot reload), Docker |

---

## クイックスタート

```bash
# 1. 環境変数を設定
cp apps/next-app/.env.example apps/next-app/.env.local
# AUTH_SECRET, AUTH_GITHUB_ID, AUTH_GITHUB_SECRET を設定

# 2. Docker で Go API + DB を起動
make dev

# 3. Next.js を起動（別ターミナル）
cd apps/next-app && pnpm install && pnpm dev

# 4. アクセス
# Next.js: http://localhost:3000
# Go API:  http://localhost:8080
```

## Makefile コマンド

```bash
make dev          # Docker (API + DB) を起動
make down         # Docker を停止
make logs         # ログを表示
make test         # 全テスト実行
make lint         # Lint チェック
make generate     # OpenAPI から型生成
```

---

## ローカル開発

### 1) 環境変数を設定

```bash
cp apps/next-app/.env.example apps/next-app/.env.local
```

`.env.local` を編集:
- `AUTH_SECRET`: `npx auth secret` で生成
- `AUTH_GITHUB_ID` / `AUTH_GITHUB_SECRET`: GitHub OAuth App から取得
- `API_BASE_URL`: `http://localhost:8080` (サーバ専用)

### 2) Docker で Go API + DB を起動

```bash
make dev
# または
docker compose up -d
```

Go API は air による hot reload が有効。コード変更で自動再起動。

### 3) Next.js を起動

```bash
cd apps/next-app
pnpm install
pnpm dev
```

### 4) 動作確認

1. http://localhost:3000 にアクセス
2. 「Sign in with GitHub」でログイン
3. /me ページでユーザー情報が表示されればOK

---

## テスト

```bash
# Next.js ユニットテスト
pnpm --filter next-app test:run

# Next.js E2Eテスト
pnpm --filter next-app test:e2e

# Go テスト
cd apps/go-api && go test ./...

# Lint
pnpm biome:check
cd apps/go-api && go vet ./...
```

---

## 本番デプロイ (Railway)

### 1) プロジェクト作成

Railway で New Project → Empty Project

### 2) PostgreSQL 追加

- `+ Create` → Database → PostgreSQL
- 起動を確認

### 3) DB 初期化

```bash
# Postgres の Public URL を取得して実行
docker run --rm -it \
  -v "$PWD:/work" -w /work \
  postgres:17 \
  psql "<CONNECTION_URL>" -f docker/db/init.sql
```

### 4) Go API デプロイ

- Create → Service → GitHub Repo
- Root Directory: `/apps/go-api`
- Dockerfile Path: `/apps/go-api/Dockerfile`

環境変数:
- `DATABASE_URL` = `${{ Postgres.DATABASE_URL }}`
- `JWT_SECRET` = ランダム文字列 (32文字以上)

### 5) Next.js デプロイ (Vercel)

- Vercel で GitHub Repo をインポート
- Root Directory: `apps/next-app`

環境変数:
- `AUTH_SECRET`
- `AUTH_GITHUB_ID`
- `AUTH_GITHUB_SECRET`
- `API_BASE_URL` = Railway の API URL
  - Server Components / Route Handler からのみ使用（ブラウザには露出しない）

> **TODO**: 本番デプロイ時は Go API の CORS 設定、認証トークンの受け渡し方法を要検討

### 6) GitHub OAuth App 更新

本番用の Callback URL を追加:
- `https://<your-vercel-domain>/api/auth/callback/github`

---

## 新しいアプリを作る手順

### このテンプレートを使う場合

```bash
# 1. テンプレートをコピー
git clone <this-repo> my-new-app
cd my-new-app
rm -rf .git
git init

# 2. プロジェクト名を変更
# - package.json の name
# - go.mod の module 名
# - README.md

# 3. 環境変数を設定
cp apps/next-app/.env.example apps/next-app/.env.local

# 4. 起動
make dev
```

### ドメインを追加する場合

```bash
# 1. DB テーブル追加
# docker/db/init.sql に CREATE TABLE を追加

# 2. Go API にエンドポイント追加
# apps/go-api/internal/handler/ に新しいハンドラー作成
# apps/go-api/openapi.yaml に API 定義追加
# main.go にルート追加

# 3. 型を再生成
pnpm --filter next-app generate:types

# 4. Next.js でページ作成
# apps/next-app/src/app/ にページ追加
```

---

## API エンドポイント

| Method | Path | 説明 |
|--------|------|------|
| GET | /ping | 疎通確認 |
| POST | /auth/signup | ユーザー登録 |
| POST | /auth/login | ログイン (JWT発行) |
| POST | /auth/oauth/callback | OAuth連携 (Next.js から呼び出す) |
| GET | /auth/me | 認証ユーザー情報 |
| GET | /users | ユーザー一覧 |

詳細: `apps/go-api/openapi.yaml`

### 認証について

- **Next.js**: Auth.js (GitHub OAuth) でセッション管理
- **Go API**: JWT エンドポイントはテンプレとして残存。実際の認証フローは Next.js (BFF) 経由で `/auth/oauth/callback` を呼び出す
- **JWT_SECRET**: 将来の API 間認証やモバイル対応のために残している
