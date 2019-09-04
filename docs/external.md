# 外部サービス 開発者用ドキュメント

[外部サービスAPIの仕様](../webapp/docs/EXTERNAL_SERVICE_SPEC.md)を参照のこと

## payment service

### 補足

* `POST /card`
  * カード番号に`FA10`が含まれる場合は必ず失敗する隠し機能がある
    * これは参加者には非公開想定、理由はFAILに近いから

## shipment service
