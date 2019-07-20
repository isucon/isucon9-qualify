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

