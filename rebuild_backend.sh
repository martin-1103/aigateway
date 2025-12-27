#!/bin/bash

echo "=== Finding process on port 8088 ==="
PID=$(netstat -ano | grep ":8088" | grep LISTENING | awk '{print $NF}' | head -1)

if [ -n "$PID" ]; then
  echo "Found process $PID using port 8088"

  echo "Killing process $PID using taskkill..."
  taskkill /PID $PID /F 2>/dev/null || taskkill //PID $PID //F 2>/dev/null || true
  sleep 2

  # Check if still running
  STILL_RUNNING=$(netstat -ano 2>/dev/null | grep ":8088" | grep LISTENING | awk '{print $NF}' | head -1)
  if [ -n "$STILL_RUNNING" ]; then
    echo "❌ FAILED: Process still running on port 8088 (PID: $STILL_RUNNING)"
    echo "Cannot proceed. Please manually kill the process."
    exit 1
  fi
  echo "✓ Process killed successfully"
else
  echo "No process found on port 8088"
fi

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
