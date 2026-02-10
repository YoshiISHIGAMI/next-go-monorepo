# CLAUDE.md

このファイルはClaude Code（AI）およびチームメンバーへのガイドです。

## プロジェクト概要

Go API + Next.js のフルスタックモノレポ。将来のプロジェクトのテンプレートとして使用。

### 技術スタック

- **Frontend**: Next.js 16.1.6 LTS, React 19, Tailwind CSS v4, shadcn/ui
- **Backend**: Go (Echo), PostgreSQL
- **Auth**: Auth.js v5 (GitHub OAuth)
- **Tooling**: pnpm, Biome, Vitest, Playwright

## よく使うコマンド

```bash
# 開発
pnpm dev                          # Next.js 起動 (apps/next-app内で)
docker compose up -d --build      # Go API + DB 起動

# テスト
pnpm --filter next-app test:run   # Vitest ユニットテスト
pnpm --filter next-app test:e2e   # Playwright E2E

# Lint/Format
pnpm biome:check                  # チェックのみ
pnpm biome:fix                    # 自動修正
```

## ディレクトリ構成 (Next.js - FSD)

```
apps/next-app/src/
├── app/                 # App Router (ルーティング)
│   ├── api/            # Route Handlers
│   ├── login/          # ログインページ
│   └── me/             # マイページ
├── components/
│   └── ui/             # shadcn/ui コンポーネント (テスト不要)
├── features/           # 機能単位のコード (将来用)
└── shared/
    ├── config/         # 環境変数など
    ├── lib/            # ユーティリティ (auth, utils)
    └── api/            # APIクライアント
```

## コーディング規約

### インポート

- **barrel export 禁止**: `index.ts` からの再エクスポートはしない
- **直接インポート**: `import { auth } from "@/shared/lib/auth"`
- **Biome**: インポート順序は Biome が自動整理 (`pnpm biome:fix`)

### ファイル配置

- shadcn/ui コンポーネント → `src/components/ui/`
- 自作の共有コンポーネント → `src/shared/ui/` (将来用)
- 機能固有のコンポーネント → `src/features/<feature>/ui/`

### スタイル

- Tailwind CSS v4 (CSS-first、`@theme` で設定)
- OKLCH カラー (shadcn/ui デフォルト)

## テスト方針

### テストすべきもの

- ユーティリティ関数 (`utils.ts` など)
- ビジネスロジックを含むカスタムフック
- 自作コンポーネント (features/ 配下)

### テスト不要なもの

- shadcn/ui コンポーネント (外部ライブラリ)
- 設定ファイル (`auth.ts` など)
- 単純なラッパーコンポーネント

### テストの種類

| 種類 | ツール | 対象 |
|------|--------|------|
| Unit | Vitest | ユーティリティ、ロジック |
| Component | Vitest + RTL | 自作コンポーネント |
| E2E | Playwright | ユーザーフロー全体 |

## 認証フロー

```
[ユーザー] → [Next.js] → [Auth.js] → [GitHub OAuth]
                ↓
         [Go API /auth/oauth/callback]
                ↓
         [users + auth_identities テーブル]
```

- 内部ユーザーID を使用 (GitHub ID ではない)
- 同じメールなら異なるプロバイダーでも同一ユーザー

## 注意事項

- `.env.local` はコミットしない
- Go API の変更後は `docker compose up -d --build` で再ビルド
- AUTH_SECRET は `npx auth secret` で生成
