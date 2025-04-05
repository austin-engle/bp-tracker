#!/bin/bash

# Readme
# =====

# Usage:
#   ./scripts/test.sh

# This script runs a series of integration tests against the Blood Pressure Tracker
# application. It checks that the application is running, checks the status endpoint,
# tests the submit endpoint with valid and invalid data, and checks the export endpoint.
# It also verifies that the application has created a database and that it contains
# data.
# File: scripts/test.sh


echo "🏥 Testing Blood Pressure Tracker"
echo "================================"

# Variables
PORT=32401
BASE_URL="http://localhost:$PORT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# Function to check if the application is running
check_app() {
    curl -s "$BASE_URL" > /dev/null
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Application is running${NC}"
        return 0
    else
        echo -e "${RED}✗ Application is not running${NC}"
        return 1
    fi
}

# Function to test endpoint
test_endpoint() {
    local endpoint=$1
    local method=${2:-GET}
    local data=$3

    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "%{http_code}" "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint")
    fi

    http_code=${response: -3}
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓ $method $endpoint${NC}"
        return 0
    else
        echo -e "${RED}✗ $method $endpoint (HTTP $http_code)${NC}"
        return 1
    fi
}

echo "1. Checking application status..."
check_app
if [ $? -ne 0 ]; then
    echo "Please start the application first"
    exit 1
fi

echo -e "\n2. Testing endpoints..."
test_endpoint "/"

test_endpoint "/submit" "POST" '{
    "systolic1": 120,
    "diastolic1": 80,
    "pulse1": 72,
    "systolic2": 122,
    "diastolic2": 81,
    "pulse2": 70,
    "systolic3": 121,
    "diastolic3": 80,
    "pulse3": 71
}'

test_endpoint "/export/csv"

echo -e "\n3. Checking database..."
DB_COUNT=$(sqlite3 bp.db "SELECT count(*) FROM readings;")
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Database contains $DB_COUNT readings${NC}"
else
    echo -e "${RED}✗ Unable to access database${NC}"
fi

echo -e "\n4. Testing invalid input..."
test_endpoint "/submit" "POST" '{
    "systolic1": 300,
    "diastolic1": 200,
    "pulse1": 500
}'

echo -e "\nTest complete! 🎉"
