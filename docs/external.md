# 外部サービス

## payment service

```
POST /card
| user | == shop id => (CORS API) | payment service |
          card number

         <=  token  ==


POST /token
| user | == token => | isucari | == token  => | payment service |
                                   shop id
                                   api key
                                    price

                                 <= result ==
```

### URL

* `POST /card`
  * カード番号を外部サービスから投げてもらうための口
  * 外部サービスから叩く前提なのでCORSに対応
  * 加盟店IDとカード番号を送ると、その加盟店IDで5分間だけ使えるトークンを発行できる
  * カード番号の形式は`^[0-9A-F]{8}`
  * カード番号に`FA10`が含まれる場合は必ず失敗する隠し機能がある
    * これは参加者には非公開想定、理由はFAILに近いから
* `POST /token`
  * 実際に決済をさせる
  * APIキーとトークンと値段を送ると実際に決済が行われる


## shipment service

```
# POST /create
| isucari | == address => | shipment service |
                name
            <=   id    ==
 (initial)

# POST /request
| isucari | == id => | shipment service |
        <= url (/accept) ==
(wait_pickup)

# GET /accept
| operator | == GET => | shipment service |

([sync] shipping)

# GET /status
| isucari | == id =>  | shipment service |
          <= status ==

(shipping, done)
```

### URL

* `POST /create`
  * 配送予約の作成
  * 配送先・配送元の住所・名前を送ると数字10桁のid（配送予約ID）が送られてくる
* `POST /request`
  * 配送リクエスト
  * 配送予約IDを送ると配送リクエストができる
  * レスポンスは `/accept` へのURLが書かれているQRコード
  * サービス側は画像を保存して、自前で配信する必要がある
* `GET /accept`
  * 配送受付
  * 配送予約IDとシードをSHA1したトークンと一緒にリクエストをする必要がある（QRコードに含まれる）
  * オペレータが叩く想定なので、オペレータしか知らない認証を入れるべきだが、ISUCON的に厳しそうなので、一旦認証はなし
* `GET /status`
  * 配送ステータス
  * 配送予約IDを送ると `initial`, `wait_pickup`, `shipping`, `done` のどれかのステータスが返ってくる

### ステータスの仕様

ステータスは以下

  * `initial`: 配送予約作成
    * `/create` を呼ばれた後はこの状態
  * `wait_pickup`: 配送待ち
    * `/request` を呼ばれた後はこの状態
  * `shipping`: 配送中
    * `/accept` を呼ばれた後はこの状態
  * `done`: 配送済み
    * 配送が終了するとこのステータスになる
