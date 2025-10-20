# rucQ

旧：https://github.com/traP-jp/rucQ

- フロントエンド（ユーザー用）：https://github.com/traPtitech/rucQ-UI
- フロントエンド（管理者用）：https://github.com/traPtitech/rucQ-Admin

## 開発環境の立ち上げ

```sh
docker compose up --build # またはdocker compose watch
go run dev/create_bot.go >> .env # 初回のみ

# 終了
docker compose down
```

rucQ Server（localhost:8080）の他に

- [rucQ UI](http://localhost:3002)
- [rucQ Admin](http://localhost:3003)
- [Adminer](http://localhost:8082/?server=mariadb&username=root&db=rucq)（パスワード：`password`）
- [Swagger UI](http://localhost:8081)
- [traQ](http://localhost:3000)

などが立ち上がります。
rucQのユーザーはデフォルトでは`traq`ですが、traQでユーザーを作成して.envに`RUCQ_USER`を設定すると切り替えることができます。
詳しくは[compose.yaml](./compose.yaml)を参照してください。

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
