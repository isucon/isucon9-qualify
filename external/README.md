# 外部サービス

## payment service

```
POST /card
| user | == shop id => (CORS API) | payment service |
          card number

         <=  token  ==


POST /token
| user | == token => | isucari | == token  => | payment service |
                                   api key
                                    price

                                 <= result ==
```

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
