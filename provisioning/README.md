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

ansible-playbook webapp.yml -i inventory/hosts -e ansible_python_interpreter=/usr/bin/python3 -e ansible_ssh_host=**** -e ansible_ssh_user=ubuntu -e ansible_ssh_private_key_file=****
```

## ベンチマーカーサーバのセットアップ

inventory/hostsのbenchセクションに対象のホストを追加してansible-playbookコマンドを実行してください。

```
ansible-playbook bench.yml -i inventory/hosts

ansible-playbook bench.yml -i inventory/hosts -e ansible_python_interpreter=/usr/bin/python3 -e ansible_ssh_host=**** -e ansible_ssh_user=ubuntu -e ansible_ssh_private_key_file=****
```

## 開発用の外部サービスのセットアップ

inventory/hostsのdevセクションに対象のホストを追加してansible-playbookコマンドを実行してください。

```
ansible-playbook dev.yml -i inventory/hosts
```
