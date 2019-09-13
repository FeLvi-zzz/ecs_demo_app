# ecs_demo_app

ECSの勉強会で使う簡単なアプリケーション(API)

## Docker Image

[grandcolline/ecs_demo_app](https://hub.docker.com/r/grandcolline/ecs_demo_app/)

## RUN

```
$ docker run --rm -p 80:8080 -it grandcolline/ecs_demo_app
```

## USAGE

```
GET /hc                      : ヘルスチェック
GET /info                    : インスタンスID, タスクARN, コンテナIDを返す
GET /down                    : アプリケーション終了
GET /fibo?n=5                : n番目のフィボナッチ数を返す
GET /zipcode?zipcode=1000001 : 郵便番号から住所を返す
```
