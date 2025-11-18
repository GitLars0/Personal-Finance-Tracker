# Prometheus and Grafana Testing Guide

This guide provides comprehensive instructions for testing Prometheus and Grafana in the cloud deployment.

## 🔗 Access URLs

- **Prometheus**: http://34.51.237.29:9090
- **Grafana**: http://34.51.152.243:3000
  - Username: `admin`
  - Password: `admin123`

---

## 1. Testing Prometheus

### A. Verify Prometheus is Running

Open http://34.51.237.29:9090/graph in your browser to access the Prometheus UI.

### B. Check Target Health

1. Navigate to **Status → Targets** (http://34.51.237.29:9090/targets)
2. Verify that the `finance-tracker` job is showing as **UP**
3. Check that the target `finance-app-service:80` has state **UP**

### C. Test Key Metrics Queries

In the Prometheus query interface (http://34.51.237.29:9090/graph), test the following queries:

#### HTTP Request Metrics

```promql
# Total HTTP requests
http_requests_total

# Requests by status code
http_requests_total{status="200"}
http_requests_total{status="404"}
http_requests_total{status="500"}

# Request rate per second (last 5 minutes)
rate(http_requests_total[5m])

# Request duration (P95 - 95th percentile)
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Average request duration
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])
```

#### Business Metrics

```promql
# Total transactions created
transactions_created_total

# Total accounts created
accounts_created_total

# Total budgets created
budgets_created_total

# Active users
active_users_total
```

#### Database Metrics

```promql
# Total database queries
db_queries_total

# DB query duration (average)
rate(db_query_duration_seconds_sum[5m]) / rate(db_query_duration_seconds_count[5m])

# DB queries by operation type
db_queries_total{operation="SELECT"}
db_queries_total{operation="INSERT"}
```

#### System Metrics

```promql
# Request/Response sizes
http_request_size_bytes
http_response_size_bytes

# Request duration by endpoint
http_request_duration_seconds{endpoint="/api/accounts"}
http_request_duration_seconds{endpoint="/api/transactions"}
```

### D. Generate Test Traffic

To see metrics in action, make some API calls:

```bash
# Health check (generates metrics)
curl http://34.51.237.29:9090/health

# Test your backend API endpoint
curl http://YOUR_BACKEND_IP/health
curl http://YOUR_BACKEND_IP/metrics
```

Then refresh your Prometheus queries to see the metrics update.

---

## 2. Testing Grafana

### A. Login to Grafana

1. Open http://34.51.152.243:3000
2. Enter username: `admin`
3. Enter password: `admin123`

### B. Configure Prometheus Data Source

1. Navigate to **Configuration → Data Sources** (gear icon on left sidebar)
2. Click **Add data source**
3. Select **Prometheus**
4. Configure the following settings:
   - **Name**: `Prometheus`
   - **URL**: `http://34.51.237.29:9090`
   - **Access**: `Server (default)`
5. Click **Save & Test** - you should see "Data source is working"

### C. Create Your First Dashboard

1. Click **+ → Dashboard** (or **Dashboards → New Dashboard**)
2. Click **Add a new panel**

#### Panel 1: HTTP Request Rate

- **Query**: `rate(http_requests_total[5m])`
- **Panel title**: "HTTP Requests per Second"
- **Visualization**: Time series graph

#### Panel 2: Request Success Rate

- **Query**: 
  ```promql
  sum(rate(http_requests_total{status=~"2.."}[5m])) / sum(rate(http_requests_total[5m])) * 100
  ```
- **Panel title**: "Success Rate (%)"
- **Visualization**: Stat or Gauge

#### Panel 3: Response Time (P95)

- **Query**: 
  ```promql
  histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
  ```
- **Panel title**: "95th Percentile Response Time"
- **Visualization**: Time series
- **Unit**: seconds (s)

#### Panel 4: Request Duration by Endpoint

- **Query**: 
  ```promql
  avg by(endpoint) (rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m]))
  ```
- **Panel title**: "Average Response Time by Endpoint"
- **Visualization**: Bar gauge or Time series

#### Panel 5: Business Metrics

- **Query**: 
  ```promql
  transactions_created_total
  accounts_created_total
  budgets_created_total
  ```
- **Panel title**: "Business Metrics"
- **Visualization**: Stat panels (create 3 separate panels)

#### Panel 6: Database Query Rate

- **Query**: 
  ```promql
  rate(db_queries_total[5m])
  ```
- **Panel title**: "Database Queries per Second"
- **Visualization**: Time series

### D. Import Pre-built Dashboard (Optional)

1. Navigate to **Dashboards → Import**
2. Enter dashboard ID: `1860` (Node Exporter Full) or `3662` (Prometheus 2.0 Overview)
3. Select your Prometheus data source
4. Click **Import**

---

## 3. End-to-End Testing Steps

### Step 1: Generate Traffic

```bash
# Make repeated requests to generate metrics
for i in {1..100}; do
  curl http://YOUR_BACKEND_IP/health
  curl http://YOUR_BACKEND_IP/api/accounts -H "Authorization: Bearer YOUR_TOKEN"
  sleep 0.1
done
```

### Step 2: Verify in Prometheus

1. Open http://34.51.237.29:9090/graph
2. Run query: `rate(http_requests_total[1m])`
3. Switch to **Graph** tab to see the spike in traffic

### Step 3: Verify in Grafana

1. Open your Grafana dashboard
2. Set time range to "Last 15 minutes"
3. Click refresh icon (or enable auto-refresh)
4. Observe metrics updating in real-time

---

## 4. Alert Testing (Advanced)

Create an alert in Grafana:

1. Edit any panel
2. Go to **Alert** tab
3. Create alert rule:
   - **Condition**: `WHEN avg() OF query(A, 5m, now) IS ABOVE 0.5`
   - For high response time alert: 
     ```promql
     histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5
     ```
4. Configure notification channels (email, Slack, etc.)

---

## 5. Validation Checklist

### Prometheus

- [ ] Prometheus UI accessible at http://34.51.237.29:9090
- [ ] Target `finance-tracker` showing as **UP**
- [ ] Metrics endpoint returning data (`/metrics`)
- [ ] Queries returning valid data
- [ ] Graphs showing time-series data

### Grafana

- [ ] Grafana UI accessible at http://34.51.152.243:3000
- [ ] Login successful with admin/admin123
- [ ] Prometheus data source connected and tested
- [ ] Dashboards created and displaying metrics
- [ ] Real-time data updating
- [ ] Panels showing accurate metrics

---

## 6. Available Metrics

Based on the application configuration, the following metrics are available:

### HTTP Metrics

- `http_requests_total` - Total requests by method, endpoint, and status
- `http_request_duration_seconds` - Request latency histogram
- `http_request_size_bytes` - Request payload sizes
- `http_response_size_bytes` - Response payload sizes

### Business Metrics

- `transactions_created_total` - Counter for transactions
- `accounts_created_total` - Counter for accounts
- `budgets_created_total` - Counter for budgets
- `active_users_total` - Gauge for active users

### Database Metrics

- `db_queries_total` - Database query counter
- `db_query_duration_seconds` - Query execution time

---

## 7. Troubleshooting

### Prometheus Not Showing Metrics

1. Check if Prometheus target is UP: http://34.51.237.29:9090/targets
2. Verify the backend service is exposing `/metrics` endpoint
3. Check Prometheus logs for scraping errors
4. Verify network connectivity between Prometheus and backend

### Grafana Not Displaying Data

1. Verify Prometheus data source connection: **Configuration → Data Sources → Test**
2. Check time range in dashboard (metrics might be outside selected range)
3. Verify query syntax in panel editor
4. Check Grafana logs for errors

### No Metrics After Generating Traffic

1. Wait 15-30 seconds for Prometheus to scrape new metrics
2. Refresh Prometheus query or Grafana dashboard
3. Verify the application is actually receiving requests
4. Check that middleware is properly configured to collect metrics

---

## 8. Example Use Cases

### Monitoring API Performance

Use this query to track API endpoint performance:

```promql
# P50, P95, P99 latencies
histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))
```

### Tracking Error Rate

Monitor application errors:

```promql
# Error rate (5xx responses)
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100
```

### Database Performance

Monitor database query performance:

```promql
# Slow queries (> 100ms)
histogram_quantile(0.95, rate(db_query_duration_seconds_bucket[5m])) > 0.1
```

---

## Notes

This monitoring setup demonstrates production-ready observability infrastructure with:
- ✅ Real-time metrics collection
- ✅ Custom business metrics tracking
- ✅ Performance monitoring and analysis
- ✅ Scalable monitoring architecture
- ✅ Dashboard visualization capabilities
