FROM nginx:1.25-alpine

# Copy nginx configuration
COPY nginx/nginx.conf /etc/nginx/nginx.conf
COPY nginx/conf.d/ /etc/nginx/conf.d/

# Create directories for logs and static files
RUN mkdir -p /var/log/nginx /var/www/html

# Copy static files if needed (optional)
# COPY --from=frontend-builder /app/frontend/build /var/www/html

# Create nginx user and set permissions
RUN chown -R nginx:nginx /var/log/nginx /var/cache/nginx /etc/nginx/conf.d
RUN touch /var/run/nginx.pid && chown nginx:nginx /var/run/nginx.pid

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:80/health || exit 1

# Install curl for health checks
RUN apk add --no-cache curl

# Use non-root user
USER nginx

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]