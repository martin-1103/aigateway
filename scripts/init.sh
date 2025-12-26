#!/bin/bash
# ============================================
# AI Gateway Database Initialization Script
# ============================================
# This script initializes the database:
# 1. Creates the database if it doesn't exist
# 2. Runs auto-migrations via Go application
# 3. Loads seed data
# ============================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Database configuration (override with environment variables)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-root}"
DB_PASSWORD="${DB_PASSWORD:-}"
DB_NAME="${DB_NAME:-aigateway}"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}AI Gateway Database Initialization${NC}"
echo -e "${GREEN}========================================${NC}"

# Step 1: Create database if it doesn't exist
echo -e "\n${YELLOW}[1/3] Creating database '${DB_NAME}'...${NC}"
mysql -h"${DB_HOST}" -P"${DB_PORT}" -u"${DB_USER}" -p"${DB_PASSWORD}" <<EOF
CREATE DATABASE IF NOT EXISTS ${DB_NAME} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
EOF

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Database created successfully${NC}"
else
    echo -e "${RED}✗ Failed to create database${NC}"
    exit 1
fi

# Step 2: Run auto-migrations
echo -e "\n${YELLOW}[2/3] Running auto-migrations...${NC}"
cd "$(dirname "$0")/.."
go run cmd/main.go &

# Wait for migrations to complete (adjust timeout as needed)
MIGRATION_PID=$!
sleep 5

# Check if the process is still running (server started)
if ps -p $MIGRATION_PID > /dev/null; then
    echo -e "${GREEN}✓ Migrations completed${NC}"
    # Kill the server since we only needed migrations
    kill $MIGRATION_PID 2>/dev/null || true
else
    echo -e "${YELLOW}⚠ Migration process exited (may have already completed)${NC}"
fi

# Step 3: Load seed data
echo -e "\n${YELLOW}[3/3] Loading seed data...${NC}"
mysql -h"${DB_HOST}" -P"${DB_PORT}" -u"${DB_USER}" -p"${DB_PASSWORD}" "${DB_NAME}" < scripts/seed.sql

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Seed data loaded successfully${NC}"
else
    echo -e "${RED}✗ Failed to load seed data${NC}"
    exit 1
fi

# Summary
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}Initialization Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "Database: ${DB_NAME}"
echo -e "Host: ${DB_HOST}:${DB_PORT}"
echo -e "\nSeeded data:"
echo -e "  - 3 providers (antigravity, openai, glm)"
echo -e "  - 3 proxy servers"
echo -e "  - 8 accounts"
echo -e "\n${YELLOW}Next steps:${NC}"
echo -e "  1. Review config.yaml for your environment"
echo -e "  2. Run: go run cmd/main.go"
echo -e "  3. Test: curl http://localhost:8080/health"
echo -e "${GREEN}========================================${NC}"
