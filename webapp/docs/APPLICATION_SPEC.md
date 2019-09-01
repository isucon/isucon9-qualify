# ISUCARI アプリケーション仕様

<img src="../frontend/public/logo.png" alt="ロゴ" height="300px" />

ISUCARIは椅子を売りたい人/買いたい人をつなげるフリマアプリです。
従来のECサービスと比べて以下の特徴があります。

* 安心安全の決済基盤
* 匿名配送により住所を伝えなくても取引が可能に
* 買いたい/売りたいと思った時にすぐに使えるシンプルさ

## ISUCARIの使い方

### 椅子を売ってみよう！

1. 椅子の情報をいれよう！
    - タイムラインページの右下の出品ボタンを押すと出品画面にいくよ！
    - シンプルなフォームに情報を入力すれば即出品♪
    - ![1-1](images/1-1.png)
1. 売れるのを待とう！
    - あなたの椅子が買われるのを楽しみに待とう♪
1. 商品を送ろう！
    - 無事購入されたら椅子を送ろう！
    - 商品ページかマイページから取引画面に行こう👀

### 椅子を送ろう！

1. 発送予約をしよう！
    - 取引画面から発送予約をして椅子を送る準備をしよう😤
    - ![2-1](images/2-1.png)
1. 配達所まで椅子を届けよう
    - 発送予約をして表示されたQRコードを配達所の人に見せよう🏃‍♀️
    - 椅子を渡したら発送完了しよう♪
    - ![2-2](images/2-2.png)
1. 購入者の受け取りを待とう！
    - 椅子が届くのをまとう♪
    - 届いたかどうかは取引画面で確認できるぞ！

### 椅子を買おう！

1. ほしい椅子を探そう！
    - タイムラインかカテゴリタイムラインから好みの椅子を探そう👀
    - カテゴリタイムラインへはサイドバーからいけるよ！
    - ![3-1](images/3-1.png)
1. 椅子を買おう！
    - 運命の椅子を見つけたら購入しよう😎
    - 安心安全の決済で簡単1ステップ購入！
    - ![3-2](images/3-2.png)
1. 椅子が届くのを待とう⏱
    - 出品者が発送するのを待とう！
    - 発送ステータスは取引画面で確認できるぞ！

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
