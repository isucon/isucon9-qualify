# isucon9-qualify

## ディレクトリ構成

```
├── bench        # ベンチマーカーなどが依存するパッケージのソースコード
├── cmd          # ベンチマーカーなどのソースコード
├── docs         # 運営が用意したイベント当日や開発に利用した各種ドキュメント
├── initial-data # 初期データ作成
├── provisioning # セットアップ用ansible
└── webapp       # 各言語の参考実装
```

## ISUCARIについて

ISUCARIの使い方、紹介は[ISUCARI アプリケーション仕様書](/webapp/docs/APPLICATION_SPEC.md)を、決済と配送のために利用している外部サービスAPIの詳細は[外部サービスAPIの仕様](/webapp/docs/EXTERNAL_SERVICE_SPEC.md)を参照してください。

### ストーリー

ISUCARIは椅子を売りたい人／買いたい人をつなげるフリマアプリです。

- 日々開発が進められ、先日もBump機能がリリースされたばかり
- 世界的な椅子ブームを追い風に順調に成長を続け
- さらなる成長を見込み社長は自腹による「ｲｽｺｲﾝ還元キャンペーン」を企画
- しかし「ｲｽｺｲﾝ還元キャンペーン」の驚異的な拡散力により負荷に耐えられないことが発覚
- 社長「緊急メンテナンスをいれていいので18時までに改修しろ。18時にプロモーション開始だ」

### 商品取得APIと更新の反映について

ISUCARIには以下の商品取得APIがあります。

- 新着一覧
- カテゴリ毎新着一覧
- ユーザ毎一覧
- 取引一覧
- 商品詳細

商品が出品・編集された場合は、カテゴリ毎新着一覧、ユーザ毎一覧、取引一覧、商品詳細APIに即座に反映してください。
編集・購入された商品については、全ての商品一覧・詳細取得APIで即座に情報を更新してください。
古いデータの削除、非表示はベンチマーク上で許可されません。各商品一覧取得APIが一度に返す商品数は初期実装と同じ状態を保つ必要があります。
新着一覧については、上記の制限を満たした上でよりユーザにあわせた商品の一覧を返すことで、購入の機会を増やすことができます。

### キャンペーン機能

`POST /initialize` のレスポンスにて、ｲｽｺｲﾝ還元キャンペーンの「還元率の設定」を返すことができます。この還元率によりユーザが増減します。

`POST /initialize` のレスポンスは JSON 形式で

```json
{
  "campaign": 0,
  "language": "実装言語"
}
```

campaignが還元率の設定となります。有効な値は 0 以上 4 以下の整数で 0 の場合はキャンペーン機能が無効になります。

languageについては別の項目で説明しています。

なお、ｲｽｺｲﾝ還元の費用が下で説明するスコアから引かれることはありません。

### ベンチマーク走行

ベンチマーク走行は以下のように実施されます。

1. 初期化処理の実行 `POST /initialize`（20秒以内）
2. アプリケーション互換性チェックの走行（適宜: 数秒〜数十秒）
3. 負荷走行（60秒）
4. 負荷走行後の確認（適宜: 数秒〜数十秒）

各ステップで失敗が見付かった場合にはその時点で停止します。
ただし、負荷走行中のエラーについては、タイムアウトや500エラーを含む幾つかのエラーについては無視され、ベンチマーク走行が継続します。

また負荷走行が60秒行われた後、レスポンスが返ってきていないリクエストはすべて強制的に切断されます。
その際にnginxのアクセスログにステータスコード499が記録されることがありますが、これらのリクエストについては減点の対象外です。

### スコア計算

スコアは**取引が完了した商品（椅子）の価格の合計（ｲｽｺｲﾝ）** をベースに以下の計算式で計算されます。

```
取引が完了した商品（椅子）の価格の合計（ｲｽｺｲﾝ） - 減点 = スコア（ｲｽｺｲﾝ）
```

以下の条件のエラーが発生すると、失格・減点の対象となります。

- 致命的なエラー
  - 1回以上で失格
  - メッセージの最後に `(critical error)` が付与されます
- HTTPステータスコードやレスポンスの内容などに誤りがある
  - 1回で500ｲｽｺｲﾝ減点、10回以上で失格
- 一定時間内にレスポンスが返却されない・タイムアウト
  - 200回を超えたら100回毎に5000ｲｽｺｲﾝ減点、失格はなし
  - メッセージの最後に `（タイムアウトしました）` が付与されます

HTTPステータスコードは、基本的に参照実装と同一のものを想定しています。またベンチマーカーのメッセージは同一のメッセージを1つにまとめます。表示されているメッセージの数とエラー数は一致しないことがあります。

また減点により0ｲｽｺｲﾝ以下になった場合は失格となります。

### `POST /initialize` での実装言語の出力

`POST /initialize` のレスポンスにて、本競技で利用した言語を出力してください。

`POST /initialize` のレスポンスは次のような JSON 形式になります。

```json
{
  "campaign": 0,
  "language": "実装言語"
}
```

languageの値が実装に利用した言語となります。languageが空の場合はベンチマーカーは失敗と見なされます。

## アプリケーションおよびベンチマーカーの起動方法

こちらのblogでも紹介しています。参考にしてください
http://isucon.net/archives/53805209.html

## 前準備

初期データを生成する。インターネットからダウンロードするので注意。

```bash
$ make init
```

## ベンチマーカー

Version: Go 1.21 or later

### build

Dockerを使う方法もある。

```bash
# ベンチマーカーbuild
$ make
$ ./bin/benchmarker
```

### 実行オプション

```bash
$ ./bin/benchmarker -help
Usage of isucon9q:
  -allowed-ips string
        allowed ips (comma separated)
  -data-dir string
        data directory (default "initial-data")
  -payment-port int
        payment service port (default 5555)
  -payment-url string
        payment url (default "http://localhost:5555")
  -shipment-port int
        shipment service port (default 7001)
  -shipment-url string
        shipment url (default "http://localhost:7001")
  -static-dir string
        static file directory (default "webapp/public/static")
  -target-host string
        target host (default "isucon9.catatsuy.org")
  -target-url string
        target url (default "http://127.0.0.1:8000")
```

  * HTTPとHTTPSに両対応
    * 証明書を検証するのでHTTPSは面倒
  * 外部サービス2つを自前で起動するので、いい感じにするならnginxを立てている必要がある
  * nginxでいい感じにするなら以下の設定が必須
    * `proxy_set_header Host $http_host;`
      * shipmentのみ必須
    * `proxy_set_header X-Forwarded-Proto "https";`
      * HTTPSでないなら不要
    * `proxy_set_header True-Client-IP $remote_addr;`
    * cf: https://github.com/isucon/isucon9-qualify/tree/master/provisioning/roles/external.nginx/files/etc/nginx


## 外部サービス

### 実行オプション

```bash
$ ./bin/shipment -help
Usage of shipment:
  -data-dir string
        data directory (default "initial-data")
  -port int
        shipment service port (default 7001)

$ ./bin/payment -help
Usage of payment:
  -port int
        payment service port (default 5555)
```

### 注意点

nginxでいい感じにするなら以下の設定が必須

  * `proxy_set_header Host $http_host;`
    * shipmentのみ必須
  * `proxy_set_header X-Forwarded-Proto "https";`
    * HTTPSでないなら不要

## webapp 起動方法

```bash
cd webapp/sql

# databaseとuserを初期化する
mysql -u root < 00_create_database.sql

# データを流し込む
cd ..
./init.sh

cd webapp/go
make
./isucari
```

## アプリケーションの動作確認

`GET /` へアクセスすることで、トップページにアクセスすることができます。
画面の「新規会員登録」から、ユーザを作成あるいは以下のテスト用ユーザが利用できます

| id       | password |
| -------- | -------- |
| isudemo1 | isudemo1 |
| isudemo2 | isudemo2 |
| isudemo3 | isudemo3 |

## Dockerを利用する

前準備を行った上で実行

### webapp

```bash
cd webapp
docker compose up
```

### benchmarker

```bash
# benchmarkerのbuild
docker build -t isucari-benchmarker -f bench/Dockerfile .

# benchmarkerの実行（Linuxは --add-host host.docker.internal:host-gateway を追加）
docker run -p 5678:5678 -p 7890:7890 -i isucari-benchmarker /opt/go/benchmarker -target-url http://host.docker.internal -data-dir /initial-data -static-dir /static -payment-url http://host.docker.internal:5678 -payment-port 5678 -shipment-url http://host.docker.internal:7890 -shipment-port 7890
```

### external service

以上だけでもベンチマークを実行することはできますが、外部サービスを起動しないと購入などのアクションを行えないため、外部サービスは別途起動する必要があります。

```bash
docker compose up
```

手元のマシンのIPアドレスが192.0.2.2の場合は以下のコマンドを実行します。ベンチマーク走行時にこの値は書き換わるので、ベンチマーク走行後に確認したい場合も都度実行する必要があります。

```
$ cat initialize.json
{
  "payment_service_url": "http://192.0.2.2:5556",
  "shipment_service_url": "http://192.0.2.2:7002"
}

$ curl -XPOST http://127.0.0.1:8000/initialize \
-H 'Content-Type: application/json' \
-d @initialize.json
```

なお外部サービス（の決済サービスAPI）はアプリケーションとブラウザ両方から同じURLでアクセスできる必要があります。

## Ansibleを利用する場合

provisioningを参照

### TLS証明書について

以下のrepoを利用して `*.t.isucon.pw` の証明書を利用している。

https://github.com/KOBA789/t.isucon.pw

証明書が古い場合、更新するスクリプトを用意している。

```bash
sudo /etc/nginx/update_cert.sh
```

### ホスト名

それぞれのアプリケーションのホスト名のprefixは以下になっている。

| host prefix | 用途                            |
| ----------- | ------------------------------- |
| isucari.    | isucariアプリケーション         |
| payment     | payment service                 |
| shipment    | shipment service                |
| bp          | benchmarker用のpayment service  |
| bs          | benchmarker用のshipment service |

以下の点に気をつけること。

* payment-serviceはブラウザとisucariアプリケーション両方からアクセスするため、ローカルとisucariアプリケーションの両方から同じホスト名でアクセスできる必要がある
  * shipment serviceはisucariアプリケーションだけでもよいが、今回は全部同じ設定をする方法を紹介する
* shipment serviceはAnsibleではHTTPSで提供されているが、HTTPに変更したい場合はnginx上の`proxy_set_header X-Forwarded-Proto "https";`を削除する必要がある

なのでisucariアプリケーションを203.0.113.1、ベンチマーカーのIPアドレスが192.0.2.1だった場合、以下のように `/etc/hosts` を指定する。

共通

```
192.0.2.1 bp.t.isucon.pw
192.0.2.1 bs.t.isucon.pw
192.0.2.1 payment.t.isucon.pw
192.0.2.1 shipment.t.isucon.pw
```

ローカル

```
203.0.113.1 isucari.t.isucon.pw
```

もちろんDNSの設定をすれば問題ない。その場合は証明書を自分で用意するか、HTTPで提供すること。ISUCON9予選本番では外部サービスはDNSを設定、競技者が利用するisucariアプリケーションはDNSを通さずに利用した。これは競技者は同じ証明書・host名を使い回していたためである。

### initialize

```
$ cat initialize.json
{
  "payment_service_url": "https://payment.t.isucon.pw",
  "shipment_service_url": "https://shipment.t.isucon.pw"
}

$ curl -XPOST https://isucari.t.isucon.pw/initialize \
-H 'Content-Type: application/json' \
-d @initialize.json
```

### ベンチマーカー

isucariアプリケーションのIPアドレスが203.0.113.1なら以下のように実行する。

```bash
/home/isucon/isucari/bin/benchmarker -target-url https://203.0.113.1 -target-host isucari.t.isucon.pw -data-dir /home/isucon/isucari/initial-data/ -static-dir /home/isucon/isucari/webapp/public/static/ -payment-url https://bp.t.isucon.pw -shipment-url https://bs.t.isucon.pw
```

なお、hostsを指定しているか、DNSが設定されているなら`-target-host`は指定する必要はなく、`-target-url https://isucari.t.isucon.pw`という指定でよい。

## 運営側のブログ

技術情報などについても記載されているので参考にしてください。

  * ISUCON9予選の出題と外部サービス・ベンチマーカーについて - catatsuy - Medium https://medium.com/@catatsuy/isucon9-qualify-969c3abdf011
  * ISUCONのベンチマーカーとGo https://gist.github.com/catatsuy/74cd66e9ff69d7da0ff3311e9dcd81fa
  * ISUCON9予選でフロントエンド周りの実装を担当した話 - はらへり日記 https://sota1235.hatenablog.com/entry/2019/10/07/110500

## サポートするMySQLのバージョン

MySQL 8.0にて動作確認しています。

ただし、nodejsでアプケーションを起動する場合、MySQL 8.0の認証方式によっては動作しないことがあります。
詳しくは、 https://github.com/isucon/isucon9-qualify/pull/316 を参考にしてください


## 使用データの取得元

- なんちゃって個人情報 http://kazina.com/dummy/
- 椅子画像提供 941-san https://twitter.com/941/status/1157193422127505412
