これまで返答された質問をまとめていきます

Q. いくつかのページにブラウザで直接アクセスするとHTTP Status 404と一緒にJSONが表示されます  
A. Node.jsの実装で以下のページのみ、このバグが確認されています

- `/items/${item_id}`
- `/transactions/${transaction_id}`
- `/users/${user_id}`
- `/users/setting`

※ `${}`には任意の変数が入ります

これらのページに関しては初期実装の時点でAPIサーバが404のJSONを返すバグが確認されています(フロントエンドで遷移する際は正常にページが表示されます)
Node.jsの実装に限り、これらのページに関してはマニュアルにある下記の制約事項の対象外とします

- アプリケーションはブラウザ上での表示を初期状態と同様に保つ必要があります

なお、これを修正する場合は初期実装の284行目のあとに下記コードを追加することで解消します

```
fastify.get("/items/:item_id", getIndex);
fastify.get("/transactions/:transaction_id", getIndex);
fastify.get("/users/:user_id", getIndex);
fastify.get("/users/setting", getIndex);
```

Q. serversタブでサーバ名を追加できません  
A. ポータルの仕様でサーバ名にunique制約がかかっているため、他のチームと被る可能性の低いサーバ名の設定をお願いします

Q. 価格モデルの設定をサブスクリプションにしたため、isucon-instance-checker実行時に価格モデルが従量課金(Pay-As-You-Go)でないためにエラーが出るが失格対象となるか  
A. 価格モデルの設定に関してはスコアに影響しないため、従量課金(Pay-As-You-Go)でなくても失格対象とはなりません

Q. Ruby実装に誤りがあります  
A. `ruby/lib/isucari/web.rb`に以下の差分を当ててください

```
@@ -336,7 +336,7 @@ module Isucari
           end

           item_detail['buyer_id'] = item['buyer_id']
-          item_detail['buyey'] = buyer
+          item_detail['buyer'] = buyer
         end
```

Q. Ruby実装でCSRFトークンのエラーが出てブラウザ上で使えません  
A. 運営が確認していない問題です。ベンチマーカーが通るなら失格対象にはしません。
