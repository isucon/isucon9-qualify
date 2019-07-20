# isucon9-qualify

**このリポジトリはtransfer repositoryの機能により、そのまま公開されます。オープンソースソフトウェア開発と同様のルールで開発してください。**

## インストール方法

```bash
$ git clone git@github.com:catatsuy/isucon9-qualify.git ~/go/src/github.com/isucon/isucon9-qualify
```

`GOPATH`を上書きしているなら、各自でいい感じにしてください。

## 方針

* 機能はディレクトリ毎に分かれているので、各ディレクトリの主担当者がディレクトリ内のファイルをいじる分にはmasterに直pushしてよい
  * とにかくコードがないと何も進まないし、あとで変更もできるため
  * PRを作ってもよいがレビューなしにmergeしてもよい
  * もちろんレビューを求めてもよい
* 主担当ではないものや他の人に影響を与える変更は必ずPRを作り、他の人に確認を取ること

## ディレクトリ構成

```
├── bench        # ベンチマーカーのソースコード
├── external     # 外部サービスのソースコード
├── provisioning # セットアップ用ansible
└── webapp       # 各言語の参考実装
```
