# isucon9 provisioning

ansible 2.8.3のみで動作確認しています。

# playbooks

- webapp.yml
  - 競技用
- bench.yml
  - ベンチマーカー用
- dev.yml
  - 競技者向け開発用の外部サービス

## 競技用サーバのセットアップ

inventory/hostsのwebappセクションに対象のホストを追加してansible-playbookコマンドを実行してください。

```
ansible-playbook webapp.yml -i inventory/hosts
```

## ベンチマーカーサーバのセットアップ

inventory/hostsのbenchセクションに対象のホストを追加してansible-playbookコマンドを実行してください。

```
ansible-playbook bench.yml -i inventory/hosts
```

## 開発用の外部サービスのセットアップ

inventory/hostsのdevセクションに対象のホストを追加してansible-playbookコマンドを実行してください。

```
ansible-playbook dev.yml -i inventory/hosts
```
