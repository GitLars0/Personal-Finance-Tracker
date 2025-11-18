# Nginx Implementation for Personal Finance Tracker

## 📋 Overview

This implementation adds Nginx as a reverse proxy for the Personal Finance Tracker, providing:

- **Load balancing** and **reverse proxy** functionality
- **SSL/TLS termination** (ready for HTTPS)
- **Rate limiting** to prevent abuse
- **Static file serving** with caching
- **Security headers** and protection
- **Centralized logging** and monitoring

## 🏗️ Architecture

```
Internet → Nginx (Port 80/443) → Backend Services
                ├── Go Backend (app:8080)
                ├── AI Service (ai-service:5001)
                ├── Prometheus (prometheus:9090)
                └── Grafana (grafana:3000)
```

## 📁 File Structure

```
nginx/
├── nginx.conf              # Main nginx configuration
├── Dockerfile             # Nginx container build
└── conf.d/
    ├── default.conf        # Main application routing
    └── monitoring.conf     # Prometheus/Grafana routing
```

## 🔧 Configuration Details

### Main Configuration (`nginx/nginx.conf`)

- **Worker processes**: Auto-detected based on CPU cores
- **Gzip compression**: Enabled for text files
- **Rate limiting**: 10 req/s for API, 5 req/s for auth
- **Security headers**: XSS protection, content-type sniffing prevention
- **Upstream definitions**: Load balancing configuration

### Application Routing (`nginx/conf.d/default.conf`)

| Route | Destination | Purpose |
|-------|-------------|---------|
| `/api/` | Backend (app:8080) | API endpoints with rate limiting |
| `/auth/` | Backend (app:8080) | Authentication with stricter limits |
| `/health` | Backend (app:8080) | Health check endpoints |
| `/metrics` | Backend (app:8080) | Prometheus metrics |
| `/ai/` | AI Service (ai-service:5001) | AI service endpoints |
| `/` | Backend (app:8080) | Frontend SPA routing |

### Monitoring Routing (`nginx/conf.d/monitoring.conf`)

- **Prometheus**: `/prometheus/` → prometheus:9090
- **Grafana**: `/grafana/` → grafana:3000
- **Authentication**: Ready for basic auth configuration

## 🚀 Deployment

### 1. Build and Start Services

```bash
# Build all services including nginx
docker-compose build

# Start all services
docker-compose up -d

# Check nginx status
docker-compose logs nginx
```

### 2. Verify Nginx is Working

```bash
# Test main application
curl http://localhost/health

# Test API endpoint
curl http://localhost/api/accounts

# Test monitoring (if monitoring.conf is enabled)
curl http://localhost/prometheus/
curl http://localhost/grafana/
```

### 3. Access Points

- **Main Application**: http://localhost
- **API Endpoints**: http://localhost/api/*
- **Health Checks**: http://localhost/health
- **Metrics**: http://localhost/metrics
- **Prometheus**: http://monitoring.localhost/prometheus/
- **Grafana**: http://monitoring.localhost/grafana/

## ⚡ Performance Features

### Rate Limiting
- **API endpoints**: 10 requests/second with 20 burst
- **Auth endpoints**: 5 requests/second with 10 burst
- **Per client IP** tracking

### Caching
- **Static assets**: 1 year cache with immutable headers
- **Health checks**: No caching
- **API responses**: Pass-through (backend controlled)

### Compression
- **Gzip enabled** for text content
- **Minimum size**: 10KB
- **Types**: HTML, CSS, JS, JSON, XML, SVG

### Connection Management
- **Keep-alive**: Enabled with 65s timeout
- **Upstream keep-alive**: 32 connections for backend, 16 for AI service
- **Buffer optimization**: Configured for optimal performance

## 🔒 Security Features

### Headers
```
X-Frame-Options: SAMEORIGIN
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
```

### Access Controls
- **Hidden files**: Denied access to `.` files
- **Backup files**: Denied access to `~` files
- **Metrics endpoint**: Can be restricted by IP (commented out)

### Error Handling
- **Custom error pages**: 404, 50x handled gracefully
- **Graceful fallbacks**: SPA routing support

## 🔧 Customization

### Enable HTTPS

1. **Add SSL certificates** to `nginx/ssl/`
2. **Update docker-compose.yml**:
   ```yaml
   nginx:
     volumes:
       - ./nginx/ssl:/etc/nginx/ssl:ro
   ```
3. **Modify configuration** to add SSL server block

### Restrict Monitoring Access

Uncomment in `nginx/conf.d/default.conf`:
```nginx
location /metrics {
    allow 10.0.0.0/8;
    allow 172.16.0.0/12;
    allow 192.168.0.0/16;
    deny all;
    
    proxy_pass http://backend;
    # ... rest of config
}
```

### Add Basic Authentication

1. **Create password file**:
   ```bash
   htpasswd -c nginx/.htpasswd admin
   ```
2. **Add to monitoring config**:
   ```nginx
   auth_basic "Protected Area";
   auth_basic_user_file /etc/nginx/.htpasswd;
   ```

### Adjust Rate Limits

Modify in `nginx/nginx.conf`:
```nginx
limit_req_zone $binary_remote_addr zone=api:10m rate=20r/s;  # Increase rate
limit_req_zone $binary_remote_addr zone=auth:10m rate=10r/s; # Increase auth rate
```

## 📊 Monitoring & Logging

### Access Logs
- **Location**: `/var/log/nginx/access.log`
- **Format**: Includes response times and upstream metrics
- **Volume**: `nginx_logs` for persistence

### Error Logs
- **Location**: `/var/log/nginx/error.log`
- **Level**: Warning and above
- **Volume**: Included in `nginx_logs`

### Health Checks
```bash
# Check nginx container health
docker-compose ps nginx

# View nginx logs
docker-compose logs -f nginx

# Test specific endpoints
curl -I http://localhost/health
```

## 🚨 Troubleshooting

### Common Issues

1. **502 Bad Gateway**
   - Check backend services are running: `docker-compose ps`
   - Verify upstream definitions in nginx.conf

2. **404 Errors for API calls**
   - Check route definitions in `conf.d/default.conf`
   - Verify backend is exposing correct ports

3. **Rate Limiting Errors (429)**
   - Adjust rate limits in `nginx.conf`
   - Check if testing from single IP

4. **SSL/TLS Issues**
   - Verify certificate paths and permissions
   - Check SSL configuration syntax

### Debug Commands
```bash
# Test nginx configuration
docker-compose exec nginx nginx -t

# Reload nginx configuration
docker-compose exec nginx nginx -s reload

# View active connections
docker-compose exec nginx ss -tuln

# Check upstream status
curl -I http://localhost/health
```

## 📈 Next Steps

1. **SSL/HTTPS Setup**: Add certificates and HTTPS configuration
2. **WAF Integration**: Consider ModSecurity or similar
3. **CDN Integration**: Add CloudFlare or similar for static assets
4. **Advanced Monitoring**: Integrate with ELK stack or similar
5. **Auto-scaling**: Configure upstream health checks and failover

## 💡 Benefits of This Implementation

- ✅ **Single entry point** for all services
- ✅ **Improved security** with rate limiting and headers
- ✅ **Better performance** with caching and compression
- ✅ **Production ready** with health checks and monitoring
- ✅ **Scalable architecture** ready for load balancing
- ✅ **Development friendly** with easy local testing