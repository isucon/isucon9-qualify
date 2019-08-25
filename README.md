# isucon9-qualify

**このリポジトリはtransfer repositoryの機能により、そのまま公開されます。オープンソースソフトウェア開発と同様のルールで開発してください。**

## インストール方法

```bash
$ git clone git@github.com:catatsuy/isucon9-qualify.git ~/go/src/github.com/isucon/isucon9-qualify
```

`GOPATH`を上書きしているなら、各自でいい感じにしてください。

## マージ方針

* 機能はディレクトリ毎に分かれているので、各ディレクトリの主担当者がディレクトリ内のファイルをいじる分にはmasterに直pushしてよい
  * とにかくコードがないと何も進まないし、あとで変更もできるため
  * PRを作ってもよいがレビューなしにmergeしてもよい
  * もちろんレビューを求めてもよい
* 主担当ではないものや他の人に影響を与える変更は必ずPRを作り、他の人に確認を取ること

## ディレクトリ構成

```
├── bench        # ベンチマーカーのソースコード
├── initial-data # 初期データ作成
├── provisioning # セットアップ用ansible
└── webapp       # 各言語の参考実装
```


## ベンチマーカー実行方法

```
# 初期データ作成
$ cd initial-data
$ make

$ make
$ ./bin/benchmarker
```

## 開発方法

tmuxで端末を複数立てるのがおすすめ。外部サービスを手元で立てるなら以下のようにする。

```
$ make
$ ./bin/payment
$ ./bin/shipment
```

## webapp 起動方法


### DB初期化

```shell-session
cd webapp/sql

# databaseとuserを初期化する
mysql -u root < 00_create_database.sql

# データを流し込む
./init.sh
```


### 起動方法 (go)

```shell-session
cd webapp/go
go run api.go main.go

# or
go build
./go
```

## 参考実装移植について

  * 実装は`webapp`ディレクトリ以下に各言語名でディレクトリを作って、その中で実装
    * 基本はマージ方針に従って開発して欲しいですが、PRは作ってください
  * GoがマスターなのでGoの実装に従って実装してください
  * バージョンは基本的に実装開始時の最新版を使う
  * 同じ挙動にするために無理な実装をする必要がある場合は相談してください
    * Goは異常系を全部記述する必要がありますが、他の言語だとそもそも記述しないで曖昧にすることが多いと思います
    * 他の言語への移植が厳しそうな実装は避けているつもりですが、完全に把握しているわけではないので相談してください
  * 各言語で自然な実装にすることを心がけてください
    * 今回パスワードのハッシュ化にbcryptを使用していますが、PHPではデフォルトのため `password_hash` `password_verify` に置き換えるだけで利用可能です
    * `secureRandomStr`という関数をGoでは作っていますが、Rubyなら`SecureRandom.hex`を呼ぶだけでいいはずです
    * ライブラリ選定は過去のISUCONの実装も参考にしつつ、各言語で一般的なものを極力選んでください
  * 初期実装やベンチマーカーの挙動で気になることがあれば教えてください
    * 例年実装移植中に問題が見つかります

## 使用データの取得元

- なんちゃって個人情報 http://kazina.com/dummy/
