#!/bin/bash

curl -L -s -o /etc/nginx/ssl/fullchain.pem https://github.com/KOBA789/t.isucon.pw/releases/latest/download/fullchain.pem
curl -L -s -o /etc/nginx/ssl/key.pem https://github.com/KOBA789/t.isucon.pw/releases/latest/download/key.pem

chmod 0600 /etc/nginx/ssl/fullchain.pem
chmod 0600 /etc/nginx/ssl/key.pem

sudo systemctl reload nginx
