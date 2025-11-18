#!/bin/bash
# filepath: tests/performance/run_tests.sh

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "🚀 Starting Performance Tests..."
echo "📁 Working directory: $SCRIPT_DIR"

# Navigate to project root to run docker compose
PROJECT_ROOT="$SCRIPT_DIR/../.."
cd "$PROJECT_ROOT"
echo "📁 Starting services from: $(pwd)"

docker compose up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to start..."
sleep 10

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo "❌ k6 is not installed. Install it with: brew install k6"
    exit 1
fi

# Go back to the performance test directory
cd "$SCRIPT_DIR"
echo "📁 Running tests from: $(pwd)"

# Create results directory
mkdir -p results

# Verify test files exist in current directory
if [ ! -f "load_test.js" ]; then
    echo "❌ Error: load_test.js not found in current directory"
    ls -la
    exit 1
fi

echo "✅ All test files found in current directory"

# Run tests using relative paths (we're already in the test directory)
echo ""
echo "📊 Running Load Test..."
k6 run --out json=results/load_test_results.json load_test.js

echo ""
echo "💪 Running Stress Test..."
k6 run --out json=results/stress_test_results.json stress_test.js

echo ""
echo "⚡ Running Spike Test..."
k6 run --out json=results/spike_test_results.json spike_test.js

# Generate summary report
echo ""
echo "📄 Generating Summary Report..."
k6 run --summary-export=results/summary.json load_test.js > /dev/null 2>&1

echo ""
echo "✅ Performance tests completed!"
echo "📊 Results saved to: $SCRIPT_DIR/results/"
echo ""
echo "View results:"
echo "  cd results"
echo "  cat load_test_results.json | jq '.metrics.http_req_duration'"
echo "  cat stress_test_results.json | jq '.metrics.http_req_duration'"
echo "  cat spike_test_results.json | jq '.metrics.http_req_duration'"
echo ""
echo "📊 Monitoring:"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3001 (admin/admin)"
echo "  - Metrics: http://localhost:8080/metrics"