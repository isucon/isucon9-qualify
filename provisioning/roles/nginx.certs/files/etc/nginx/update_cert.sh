#!/bin/bash

# Set variables
GITHUB_API_URL="https://api.github.com/repos/KOBA789/t.isucon.pw/releases/latest"
DEST_DIR="/etc/nginx/ssl"

# Get latest release information from GitHub
echo "Fetching latest release information from GitHub..."
response=$(curl -s "${GITHUB_API_URL}")

# Parse the JSON response and download the files
echo "Downloading and placing files..."
echo $response | jq -r '.assets[] | select(.name=="fullchain.pem" or .name=="cert.pem" or .name=="key.pem") | .browser_download_url' | while read -r url; do
  file_name=$(basename "$url")
  echo "Downloading $file_name..."
  curl -L -s -o "${DEST_DIR}/${file_name}" "$url"
  chmod 0600 "${DEST_DIR}/${file_name}"
done

sudo systemctl reload nginx
