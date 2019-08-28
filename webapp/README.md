# webapp

## Campaign 機能について

TODO

## 各テーブルのステータス遷移

### items

```
items
.
↓ /sell （sellerによる出品）
on_sale
↓ /buy （buyerによる購入）
trading
↓ /complete （buyerによる受け取り完了）
sold_out
```

### transaction_evidences

```
transaction_evidences
.
↓ /buy （buyerによる購入）
wait_shipping
↓ /ship_done （sellerによる配送完了）
wait_done
↓ /complete （buyerによる受け取り完了）
done
```

### shippings

```
shippings
.
↓ /buy （buyerによる購入）
initial
↓ /ship （sellerによる配送）
wait_pickup
↓ （shipment serviceのstatusから）
shipping
↓ （shipment serviceのstatusから）
done
```

### status 遷移表

|                       | WHO    | items    | transaction_evidences | shippings            |
|-----------------------|--------|----------|-----------------------|----------------------|
| postSell              | seller  | on_sale  | -                    | -                    |
| postBuy  (購入)      | buyer  | trading  | wait_shipping         | initial              |
| postShip (集荷予約)   | seller | ↓        | ↓                     | wait_pickup          |
| postShipDone (発送完了)|  seller | ↓        | wait_done             | shipping または done |
| postComplete (取引完了)| buyer  | sold_out | done                  | done                 |
