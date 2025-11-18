# Performance Testing & Monitoring

## Performance Tests

### Running Tests

```bash
# Install k6
brew install k6

# Run all tests
./tests/performance/run_tests.sh

# Run individual tests
k6 run tests/performance/load_test.js
k6 run tests/performance/stress_test.js
k6 run tests/performance/spike_test.js
```

### Test Types

- **Load Test** - Gradual ramp-up to 50 concurrent users
- **Stress Test** - Push system to 200 concurrent users
- **Spike Test** - Sudden spike to 1000 users

### Thresholds

- 95% of requests < 500ms
- Error rate < 1%
- 99% of requests < 1s (stress test)

## Monitoring

### Prometheus Metrics

Access metrics at: http://localhost:8080/metrics

Available metrics:
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request duration
- `http_request_size_bytes` - Request size
- `http_response_size_bytes` - Response size
- `transactions_created_total` - Business metric
- `accounts_created_total` - Business metric
- `budgets_created_total` - Business metric
- `active_users_total` - Current active users
- `db_queries_total` - Database queries
- `db_query_duration_seconds` - Query duration

### Grafana Dashboards

Access Grafana at: http://localhost:3001

Default login: `admin` / `admin`

Pre-configured dashboards show:
- HTTP request rates
- Response times (p50, p95, p99)
- Error rates
- Business metrics
- Database performance

### Prometheus

Access Prometheus at: http://localhost:9090

Query examples:
```promql
# Request rate
rate(http_requests_total[5m])

# 95th percentile latency
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Error rate
rate(http_requests_total{status=~"5.."}[5m])
```