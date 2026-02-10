# CLAUDE.md

このファイルは Claude Code (AI) およびチームメンバーへの開発ガイドです。

## プロジェクト概要

Go API + Next.js のフルスタックモノレポテンプレート。
新規プロジェクトのベースとして使用。

## よく使うコマンド

```bash
# 開発
make dev                          # Docker (API + DB) 起動
make logs                         # ログ表示
make down                         # 停止

# Next.js (apps/next-app で実行)
pnpm dev                          # 開発サーバー起動
pnpm build                        # ビルド

# テスト
make test                         # 全テスト
pnpm --filter next-app test:run   # Next.js ユニットテスト
cd apps/go-api && go test ./...   # Go テスト

# Lint/Format
make lint                         # 全 Lint
make fmt                          # 全フォーマット
pnpm biome:fix                    # Biome 修正

# コード生成
make generate                     # OpenAPI → TypeScript 型生成
```

## ディレクトリ構成

### Go API (`apps/go-api/`)

```
apps/go-api/
├── main.go                 # エントリポイント、ルーティング
├── openapi.yaml            # API 仕様 (Single Source of Truth)
├── internal/
│   ├── handler/            # HTTP ハンドラー (auth.go, user.go)
│   ├── middleware/         # ミドルウェア (auth, error, request_logger)
│   ├── model/              # データモデル (User, Request/Response 型)
│   ├── apperror/           # 共通エラー (BadRequest, NotFound 等)
│   └── logger/             # slog 設定
├── Dockerfile              # 本番用 (distroless)
└── Dockerfile.dev          # 開発用 (air hot reload)
```

**命名規則:**
- パッケージ名: 小文字、単数形 (`handler`, `model`)
- ファイル名: 小文字、スネークケース (`request_logger.go`)
- 関数名: キャメルケース、公開は大文字開始 (`NewAuthHandler`)

### Next.js (`apps/next-app/src/`)

```
src/
├── app/                    # App Router (ルーティング)
│   ├── api/               # Route Handlers
│   │   ├── auth/          # Auth.js ハンドラー
│   │   └── bff/           # BFF エンドポイント
│   ├── login/             # ログインページ
│   └── me/                # マイページ
├── components/
│   └── ui/                # shadcn/ui (テスト不要)
└── shared/
    ├── api/
    │   └── types.gen.ts   # OpenAPI 生成型 (編集禁止)
    ├── config/
    │   └── env.ts         # 環境変数
    └── lib/
        ├── auth.ts        # Auth.js 設定
        └── utils.ts       # ユーティリティ
```

**命名規則:**
- ファイル名: ケバブケース (`types.gen.ts`)
- コンポーネント: パスカルケース (`Button.tsx`)
- barrel export 禁止: `index.ts` からの再エクスポートはしない

## コーディング規約

### インポート順序

Biome が自動整理。手動で整える必要なし。

```typescript
// 1. 外部ライブラリ
import { redirect } from "next/navigation";
// 2. 内部モジュール (@/)
import type { components } from "@/shared/api/types.gen";
import { env } from "@/shared/config/env";
```

### エラーハンドリング (Go)

```go
// 共通エラーを使用
return apperror.BadRequest("invalid email")
return apperror.NotFound("user")
return apperror.Internal("database error")

// ドメイン固有エラー
return apperror.EmailAlreadyExists()
return apperror.InvalidCredentials()
```

レスポンス形式:
```json
{"code": "BAD_REQUEST", "message": "invalid email"}
```

### ロギング (Go)

```go
// slog を使用 (log.Println は使わない)
slog.Info("user created", "userID", user.ID)
slog.Error("failed to query", "error", err)
```

出力形式 (JSON):
```json
{"time":"...","level":"INFO","msg":"user created","userID":123}
```

## テスト方針

### テストすべきもの

- ユーティリティ関数 (`utils.ts`)
- ビジネスロジック
- 自作コンポーネント (`features/` 配下)

### テスト不要なもの

- shadcn/ui コンポーネント (外部ライブラリ)
- 設定ファイル (`auth.ts` 等)
- OpenAPI 生成ファイル (`types.gen.ts`)

### テストの種類

| 種類 | ツール | 対象 |
|------|--------|------|
| Unit | Vitest | ユーティリティ、ロジック |
| E2E | Playwright | ユーザーフロー |
| Go | go test | ハンドラー、ロジック |

## API 開発フロー

```
1. openapi.yaml を編集
       ↓
2. make generate (型生成)
       ↓
3. Go: handler を実装
       ↓
4. Next.js: 生成された型を使用
       ↓
5. TypeScript がエラーを検知 → 修正
```

## 認証フロー

```
[ユーザー] → [Next.js] → [Auth.js] → [GitHub OAuth]
                              ↓
                    [Go API /auth/oauth/callback]
                              ↓
                    [users + auth_identities テーブル]
```

- 内部ユーザー ID を使用 (GitHub ID ではない)
- 同じメールなら異なるプロバイダーでも同一ユーザー

## 注意事項

- `.env.local` はコミットしない
- `types.gen.ts` は編集しない (自動生成)
- Go API 変更後は air が自動再起動
- DB スキーマ変更後は `make db-reset`
