#!/bin/bash

set -e

DIR="/www/wwwroot/aigateway"
NODE_PATH="/www/server/nodejs/v22.19.0/bin"

echo "=== AIGateway Deploy Script ==="

cd "$DIR"

echo "[1/5] Git pull..."
git fetch origin
git reset --hard origin/master
git clean -fd

echo "[2/5] Building backend..."
cd "$DIR/backend"
GOOS=linux GOARCH=amd64 go build -o aigateway .

echo "[3/5] Building frontend..."
cd "$DIR/frontend"
export PATH="$NODE_PATH:$PATH"
npm install --legacy-peer-deps
npm run build

echo "[4/5] Restarting services..."
sudo systemctl restart aigateway
sudo systemctl restart aigateway-frontend

echo "[5/5] Checking status..."
sudo systemctl status aigateway --no-pager -l
sudo systemctl status aigateway-frontend --no-pager -l

echo "=== Deploy complete ==="
