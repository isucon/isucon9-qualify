# isucari Go application

## ビルド/実行

```sh
# build
make build

# run
make run
```
  
### build/run時のオプション
・ENABLE_TRACE：tracingを有効化するかどうか. (true/false)  
・ENABLE_PROFILE：profilingを有効化するかどうか. (true/false)  
・CAMPAIGN：キャンペーン還元率. (0~4の整数)  
  
## Trace/Profile実行前に
GCPのServiceAccountに `stackdriver@isucon-gigawatts.iam.gserviceaccount.com` が用意されているため,  
ServiceAccount Keyを作成して `/home/${USER}/serviceaccount-key.json` に配置
