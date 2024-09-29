#!/bin/bash

if [ -z "$1" ]; then
  echo "Error: IP address not provided."
  echo "Usage: $0 <IP_ADDRESS>"
  exit 1
fi

ip_address=$1

domains=(
  "bp.t.isucon.pw"
  "bs.t.isucon.pw"
  "payment.t.isucon.pw"
  "shipment.t.isucon.pw"
)

{
  echo ""
  echo "# The following entries were added automatically by a script."
  for domain in "${domains[@]}"; do
    echo "${ip_address} ${domain}"
  done
} | sudo tee -a /etc/hosts >/dev/null
