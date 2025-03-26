# OpenAPIからのコード生成手順

> [!NOTE]
> 以下の手順は旧リポジトリ向けに書かれたものです。今後の運用は検討中です。

## バックエンド

```bash
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -config oapi-codegen.yaml openapi.yaml
```

を実行すると `backend/handler/api.gen.go` が生成される。

## フロントエンド

`frontend` ディレクトリに移動して以下のコマンドを実行する。

```bash
npx openapi-typescript ../openapi.yaml -o src/api/schema.d.ts
```
