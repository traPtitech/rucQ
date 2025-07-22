# rucQ

旧：https://github.com/traP-jp/rucQ

- フロントエンド（ユーザー用）：https://github.com/traPtitech/rucQ-UI
- フロントエンド（管理者用）：https://github.com/traPtitech/rucQ-Admin

## 開発環境の立ち上げ

```sh
docker compose watch
go run dev/create_bot.go >> .env # 初回のみ

# 終了
docker compose down
```

rucQ（localhost:8080）の他に

- [Adminer](http://localhost:8082/?server=mariadb&username=root&db=rucq)（パスワード：`password`）
- [Swagger UI](http://localhost:8081)
- [traQ](http://localhost:3000)

などが立ち上がります。詳しくは[compose.yaml](./compose.yaml)を参照してください。

## コード生成

API、モックは次のコマンドで生成できます。

```sh
go generate ./...
```

## テスト

テストは次のコマンドで実行できます。

```sh
go test ./...
```
