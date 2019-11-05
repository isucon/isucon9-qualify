# 外部サービスAPIの仕様

予選マニュアルに利用可能なURLなどの情報があるので確認してください。

## payment service

決済サービスAPI。クレジットカード情報の非保持化にも対応しているので安心して利用できます。


### `POST /card`

* 加盟店ID(`shop_id`)とカード番号(`card_number`)を送ると、その加盟店IDで5分間だけ使えるトークンを発行できる
* JavaScript経由で決済サービスにアクセスするので[Cross-Origin Resource Sharing (CORS)](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)に対応
* カード番号の形式は`[0-9A-F]{8}`

#### API仕様

- request: application/json
  - shop_id
  - card_number
- response: application/json
  - http status code: 200
    - token
  - http status code: 400
    - error: json decode error
    - error: wrong shop id
    - error: card number is wrong

```
example:

# request
{
  "shop_id": "11",
  "card_number": "AAAAAAAA"
}

# response
{
  "token": "abcdefg"
}

{
  "error": "card number is wrong"
}
```

### `POST /token`

* 加盟店IDに紐付くAPIキー・加盟店IDに紐付くトークン・値段を送ると実際に決済が行われる
* 残高不足などの理由で正当なカード番号でも決済に失敗するケースがある

#### API仕様

- request: application/json
  - shop_id
  - token
  - api_key
  - price
- response: application/json
  - http status code: 200
    - status: ok
      - 決済成功
    - status: fail
      - 決済失敗
    - status: invalid
      - 無効なトークン
  - http status code: 400
    - error: json decode error
    - error: wrong shop id
    - error: wrong api key

```
example:

# request
{
  "shop_id": "11",
  "token": "abcd",
  "api_key": "itisapikey",
  "price": 10000
}

# response
{
  "status": "ok"
}

{
  "error": "wrong shop id"
}
```

## shipment service

配送サービスAPI。配送会社が直接住所を扱うことで、お客様同士は住所を教え合うことなく利用できます。

### Authorization

AuthorizationヘッダのBearerトークンに、予め払い出されているユニークなappidを使用して認証を行う。

```
Authorization: Bearer <APP_ID>
```

### `POST /create`

* 集荷予約の作成
* 配送先の住所・配送元の住所・名前を送ると数字10桁のid（集荷予約ID）が送られてくる

#### API仕様

- request: application/json
  - to_address
  - to_name
  - from_address
  - from_name
- response: application/json
  - http status code: 200
    - reserve_id: 0000000000
    - reserve_time: 1570000000
      - (epoch time)
  - http status code: 400
    - error: json decode error
    - error: required parameter was not passed
  - http status code: 401
    - （Authorization失敗）

### `POST /request`

* 集荷リクエスト
* 集荷予約IDを送ると集荷リクエストができる
* レスポンスは `/accept` へのURLが書かれているQRコードの画像ファイル
* サービス側は画像を保存して、自前で配信する必要がある

#### API仕様

- request: application/json
  - reserve_id
- response: image/png (error時はapplication/json)
  - http status code: 200
    - png binary
  - http status code: 400
    - error: json decode error
    - error: required parameter was not passed
    - error: empty
  - http status code: 401
    - （Authorization失敗）

### `GET /accept`

* 発送
* 集荷予約IDとトークンを一緒にリクエストをする必要がある
  * QRコードのURLに正しいトークンが付与されている
* 配達員が開く想定なので、配達員しか知らない認証を入れるべきだが、ISUCON的に厳しそうなので現在は認証なし

#### API仕様

- request: application/json
  - reserve_id
- response: application/json
  - http status code: 200
    - accept: ok
  - http status code: 400
    - error: wrong parameters
    - error: empty

### `GET /status`

* 配送ステータス
* 集荷予約IDを送ると `initial`, `wait_pickup`, `shipping`, `done` のどれかのステータスが返ってくる

#### API仕様

- request: application/json
  - reserve_id
- response: application/json
  - http status code: 200
    - reserve_time: 1570000000
      - (epoch time)
    - status: done
      - 下記参照
  - http status code: 400
    - error: json decode error
    - error: required parameter was not passed
    - error: empty
  - http status code: 401
    - （Authorization失敗）

### ステータスの仕様

ステータスは以下

1. `initial`: 集荷予約作成
  * `/create` を呼ばれた後はこの状態
2. `wait_pickup`: 集荷待ち
  * `/request` を呼ばれた後はこの状態
3. `shipping`: 配送中
  * `/accept` を呼ばれた後はこの状態
4. `done`: 配送完了
  * 配送が終了するとこのステータスになる
