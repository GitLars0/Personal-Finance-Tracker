#!/bin/bash

# Centralized test runner for Personal Finance Tracker

echo "🧪 Running All Tests for Personal Finance Tracker"
echo "==================================================="

# Set environment variables for testing
export JWT_SECRET="test_secret_key_for_testing_only"
export GIN_MODE="test"

# PostgreSQL test database configuration
# Uncomment and modify these if you have a PostgreSQL test database
# export TEST_DB_HOST="localhost"
# export TEST_DB_USER="postgres"
# export TEST_DB_PASSWORD="password"
# export TEST_DB_NAME="finance_tracker_test"
# export TEST_DB_PORT="5432"

# Skip database tests if no PostgreSQL test database is available
# export TEST_SKIP_DB="true"

echo ""
echo "📋 Test Configuration:"
echo "  - Tests are located in: $(pwd)"
echo "  - Backend source: $(pwd)/../backend"
echo "  - JWT Secret: $JWT_SECRET"
echo "  - Skip DB Tests: $TEST_SKIP_DB"
echo ""

if [ "$TEST_SKIP_DB" = "true" ]; then
    echo "⚠️  Database tests will be SKIPPED"
else
    echo "🐘 Database tests will run with PostgreSQL"
    echo "   Ensure your test database is configured with TEST_DB_* environment variables"
fi

echo ""

# Change to the directory where this script is located
cd "$(dirname "$0")"

echo "📋 Running all tests..."

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if CGO is available for race detection
if command -v gcc >/dev/null 2>&1 && [ "${CGO_ENABLED:-1}" != "0" ]; then
    echo "🔍 Running tests with race detection..."
    go test -v -race -coverprofile=coverage.out ./...
else
    echo "📋 Running tests without race detection (CGO not available)..."
    go test -v -coverprofile=coverage.out ./...
fi

# Check if tests passed
if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ All tests passed!${NC}"
    
    echo ""
    echo -e "${YELLOW}📊 Generating coverage report...${NC}"
    go tool cover -html=coverage.out -o coverage.html
    
    echo ""
    echo -e "${YELLOW}📈 Coverage Summary:${NC}"
    go tool cover -func=coverage.out
    
    echo ""
    echo -e "${GREEN}📄 Coverage report saved to coverage.html${NC}"
    echo -e "${GREEN}🎉 All tests completed successfully!${NC}"
    
    exit 0
else
    echo ""
    echo -e "${RED}❌ Some tests failed!${NC}"
    exit 1
fi