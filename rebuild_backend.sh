#!/bin/bash

echo "=== Killing all aigateway-backend processes ==="
powershell -Command "Get-Process aigateway-backend -ErrorAction SilentlyContinue | Stop-Process -Force" 2>/dev/null || true
sleep 2

# Verify all killed
REMAINING=$(tasklist 2>/dev/null | grep -i "aigateway-backend" | wc -l)
if [ "$REMAINING" -gt 0 ]; then
  echo "❌ FAILED: $REMAINING aigateway-backend processes still running"
  exit 1
fi
echo "✓ All aigateway-backend processes killed"

echo ""
echo "=== Building backend ==="
cd D:/temp/aigateway/backend
go build -o aigateway.exe . || exit 1
echo "✓ Build complete"

echo ""
echo "=== Starting backend (with SKIP_MIGRATION=true for faster startup) ==="
SKIP_MIGRATION=true ./aigateway.exe &
sleep 3

RUNNING=$(netstat -ano 2>/dev/null | grep ":8088" | grep LISTENING)
if [ -n "$RUNNING" ]; then
  echo "✓ Backend running on port 8088"
else
  echo "❌ Backend failed to start"
  exit 1
fi
