# isucari Go application

## ビルド/実行

```sh
# build
make build

# run
make run
```

 ### ビルド/実行時のオプション
・ENABLE_TRACE：tracingを有効化するかどうか. (true/false)
・ENABLE_PROFILE：profilingを有効化するかどうか. (true/false)
・CAMPAIGN=0：キャンペーン還元率. (0~4の整数)

 ## Trace/Profile実行前に
GCPのServiceAccountに `stackdriver@isucon-gigawatts.iam.gserviceaccount.com` が用意されているため,  
ServiceAccount Keyを作成して `/home/${USER}/serviceaccount-key.json` に配置
