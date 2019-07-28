# webapp

## 各テーブルのステータス遷移

### items

```
items
.
↓ /sell （sellerによる出品）
on_sale
↓ /buy （buyerによる購入）
trading
↓ /ship_done （sellerによる配送完了）
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
