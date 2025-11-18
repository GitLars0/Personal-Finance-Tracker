@echo off
REM Centralized test runner for Personal Finance Tracker

echo  Running All Tests for Personal Finance Tracker
echo ===================================================

REM Set environment variables for testing
set JWT_SECRET=test_secret_key_for_testing_only
set GIN_MODE=test

REM PostgreSQL test database configuration
REM Uncomment and modify these if you have a PostgreSQL test database
REM set TEST_DB_HOST=localhost
REM set TEST_DB_USER=postgres
REM set TEST_DB_PASSWORD=password
REM set TEST_DB_NAME=finance_tracker_test
REM set TEST_DB_PORT=5432

REM Skip database tests if no PostgreSQL test database is available
REM set TEST_SKIP_DB=true

echo.
echo  Test Configuration:
echo   - Tests are located in: %~dp0
echo   - Backend source: %~dp0..\backend
echo   - JWT Secret: %JWT_SECRET%
echo   - Skip DB Tests: %TEST_SKIP_DB%
echo.

if defined TEST_SKIP_DB (
    echo   Database tests will be SKIPPED
) else (
    echo  Database tests will run with PostgreSQL
    echo    Ensure your test database is configured with TEST_DB_* environment variables
)

echo.

REM Change to the tests directory
cd /d "%~dp0"

echo 📋 Running all tests...

REM Run tests with coverage (removed -race flag for Windows CGO compatibility)
go test -v -coverprofile=coverage.out ./...

if %errorlevel% equ 0 (
    echo.
    echo  All tests passed!
    
    echo.
    echo  Generating coverage report...
    go tool cover -html=coverage.out -o coverage.html
    
    echo.
    echo  Coverage Summary:
    go tool cover -func=coverage.out
    
    echo.
    echo  Coverage report saved to coverage.html
    echo  All tests completed successfully!
    
    exit /b 0
) else (
    echo.
    echo  Some tests failed!
    exit /b 1
)