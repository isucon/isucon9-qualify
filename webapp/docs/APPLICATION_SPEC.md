# ISUCARI アプリケーション仕様

<img src="../frontend/public/logo.png" alt="ロゴ" style="height: 300px" />

ISUCARIは椅子を売りたい人/買いたい人をつなげるフリマアプリです。
従来のECサービスと比べて以下の特徴があります。

* 安心安全の決済基盤
* 匿名配送により住所を伝えなくても取引が可能に
* 買いたい/売りたいと思った時にすぐに使えるシンプルさ

## ISUCARIの使い方

### 会員登録/ログイン

### 椅子を売る

#### 出品する

#### 商品を発送する

### 椅子を買う

#### 買う

#### 商品を受け取る

## Campaign 機能について

TODO

##  外部サービスの仕様

[外部サービス仕様書](EXTERNAL_SERVICE_SPEC.md) を参照

## ステータス遷移表

|                       | WHO    | items    | transaction_evidences | shippings            |
|-----------------------|--------|----------|-----------------------|----------------------|
| postSell              | 出品者  | on_sale  | -                    | -                    |
| postBuy  (購入)      | 購入者  | trading  | wait_shipping         | initial              |
| postShip (集荷予約)   | 出品者 | ↓        | ↓                     | wait_pickup          |
| postShipDone (発送完了)|  出品者 | ↓        | wait_done             | shipping または done |
| postComplete (取引完了)| 購入者  | sold_out | done                  | done                 |


## 各テーブルごとのURLとステータス遷移

#### items テーブル

```
↓ /sell （出品者による出品）
on_sale
↓ /buy （購入者による購入）
trading
↓ /complete （購入者による取引完了）
sold_out
```

#### transaction_evidences テーブル

```
↓ /buy （購入者による購入）
wait_shipping
↓ /ship_done （出品者による発送完了）
wait_done
↓ /complete （購入者による受け取り完了）
done
```

### shippings テーブル

```
↓ /buy （購入者による購入）
initial
↓ /ship （出品者による配送）
wait_pickup
↓ /ship_done（shipment serviceへ問い合わせた結果のstatusから）
shipping
↓ /ship_done（shipment serviceへ問い合わせた結果のstatusから）
done
```
