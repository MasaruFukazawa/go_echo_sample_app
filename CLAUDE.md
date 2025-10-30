# CLAUDE.md

このファイルは、Claude Code (claude.ai/code) がこのリポジトリで作業する際のガイダンスを提供します。

## プロジェクト概要

nginx リバースプロキシを備えた Go Echo サンプルアプリケーション。AWS ECS へのデプロイを想定して設計されています。nginx と Go アプリケーションが同じネットワーク名前空間を共有するマルチコンテナ構成（ECS タスクの動作を模倣）を採用しています。

## 開発コマンド

### ホットリロード付きローカル開発
```bash
# air（ホットリロード）を使った開発環境の起動
docker-compose up

# アプリケーションへのアクセス
# - http://localhost (nginx リバースプロキシ経由)
# - http://localhost:8080 (Go アプリに直接アクセス)

# 環境の停止
docker-compose down

# app コンテナに入る
docker-compose exec app sh
```

### 本番ビルド
```bash
# ECR 向け app イメージのビルド
docker build -f app/Dockerfile -t go-echo-app:latest app

# ECR 向け nginx イメージのビルド
docker build -f nginx/Dockerfile -t nginx-proxy:latest nginx

# 本番ビルドをローカルで実行
docker run -d --name go-app -p 8080:8080 -p 80:80 go-echo-app:latest
docker run -d --name nginx-proxy --network container:go-app nginx-proxy:latest
```

### Go コマンド（コンテナ内）
```bash
# Go モジュールの初期化
go mod init <module-name>

# 依存関係の追加
go mod tidy

# アプリケーションの直接実行
go run main.go
```

## アーキテクチャ

### マルチコンテナのネットワーク構成

このアプリケーションは AWS ECS 同一タスク内コンテナの動作に一致する**ネットワーク名前空間共有**パターンを使用しています：

- **開発環境 (docker-compose)**: nginx が `network_mode: "service:app"` を使って app コンテナのネットワーク名前空間を共有
- **本番環境 (ECS)**: 両方のコンテナが同じ ECS タスク定義内で実行され、localhost ネットワークを共有
- **nginx 設定**: `localhost:8080` を指定（開発・本番の両方で動作）

このアーキテクチャにより以下が保証されます：
1. 開発環境が本番のネットワーク動作と一致
2. nginx と app が localhost 経由で通信（Docker DNS 不要）
3. app コンテナのみがホストにポートを公開

### ディレクトリ構造

- `app/` - Go Echo アプリケーション
  - `Dockerfile` - 本番ビルド（マルチステージ Alpine ベース）
  - `Dockerfile.dev` - air ホットリロード付き開発ビルド
  - `main.go` - Echo Web サーバー（ポート 8080）
  - `.air.toml` - ホットリロード用 Air 設定
- `nginx/` - Nginx リバースプロキシ
  - `Dockerfile` - 本番 nginx イメージ
  - `nginx.conf` - プロキシ設定（localhost:8080 → ポート 80）

### CI/CD

GitHub Actions ワークフロー（`.github/workflows/build.yml`）が main ブランチへの push 時に両方のイメージを自動ビルドして ECR にプッシュします：

- **ECR リポジトリ**:
  - `821975976774.dkr.ecr.ap-northeast-1.amazonaws.com/go_echo_sample_app/app`
  - `821975976774.dkr.ecr.ap-northeast-1.amazonaws.com/go_echo_sample_app/nginx`
- **タグ**: `latest` とコミット SHA の両方のタグでプッシュ
- **認証**: GitHub Secrets の AWS アクセスキー認証情報を使用

### 重要な設定ポイント

1. **ポートマッピング**: docker-compose では app コンテナが 80（nginx）と 8080（アプリ直接アクセス）の両方を公開
2. **ホットリロード**: 開発環境ではファイル変更時の自動再コンパイルに `air` を使用
3. **本番バイナリ**: 静的リンク用に `CGO_ENABLED=0` でビルド（依存関係なし）
4. **ネットワーク名前空間の共有**: nginx と app 間の localhost 通信に必須
