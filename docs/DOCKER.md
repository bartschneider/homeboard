# Docker Deployment Guide

Complete Docker containerization for E-Paper Dashboard with production-ready configurations.

## Quick Start

### Localhost Preview
```bash
# Build and start the application
docker-compose up --build

# Access the dashboard
open http://localhost:8081
```

### Development Mode
```bash
# Start development environment with hot reload
docker-compose -f docker-compose.dev.yml up --build

# Access development tools
open http://localhost:8081  # Dashboard
open http://localhost:8025  # Mailhog (email testing)
open http://localhost:5432  # PostgreSQL
```

## Available Configurations

### Production Setup (`docker-compose.yml`)
- **Homeboard**: Main application container
- **Nginx**: Reverse proxy with SSL support
- **Redis**: Caching and session management (optional)
- **Prometheus**: Metrics collection (optional)
- **Grafana**: Metrics visualization (optional)

### Development Setup (`docker-compose.dev.yml`)
- **Hot Reload**: Source code mounted for development
- **Database**: PostgreSQL for testing
- **Mailhog**: Email testing interface
- **Python Environment**: Widget development container

## Container Profiles

### Basic Deployment
```bash
# Minimal setup - just the dashboard
docker-compose up homeboard
```

### Full Stack with Monitoring
```bash
# Include monitoring stack
docker-compose --profile monitoring up --build
```

### Development with All Tools
```bash
# Include all development tools
docker-compose -f docker-compose.dev.yml --profile dev-tools up --build
```

## Configuration

### Environment Variables
Copy and customize the environment file:
```bash
cp .env.example .env
# Edit .env with your settings
```

Key variables:
- `REDIS_PASSWORD`: Redis authentication
- `GRAFANA_PASSWORD`: Grafana admin password
- `TZ`: Timezone setting
- `CONFIG_PATH`: Configuration file path

### Volumes and Data Persistence
- `homeboard-data`: Application data
- `homeboard-logs`: Application logs
- `homeboard-backups`: Configuration backups
- `redis-data`: Redis persistence
- `prometheus-data`: Metrics storage
- `grafana-data`: Dashboard configurations

## Access Points

| Service | URL | Purpose |
|---------|-----|---------|
| Dashboard | http://localhost:8081 | Main dashboard interface |
| Admin Panel | http://localhost:8081/admin | Configuration management |
| Nginx Proxy | http://localhost | Production-like access |
| Grafana | http://localhost:3000 | Metrics visualization |
| Prometheus | http://localhost:9090 | Metrics collection |
| Mailhog | http://localhost:8025 | Email testing (dev only) |

## Development Workflow

### Hot Reload Development
```bash
# Start development environment
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f homeboard-dev

# Execute commands in development container
docker-compose -f docker-compose.dev.yml exec homeboard-dev bash
```

### Widget Development
```bash
# Access Python development environment
docker-compose -f docker-compose.dev.yml exec python-dev bash

# Test widget directly
python3 /app/widgets/clock.py '{"format": "%H:%M"}'

# Install additional Python packages
pip install package-name
```

### Database Access
```bash
# Connect to development database
docker-compose -f docker-compose.dev.yml exec postgres-dev psql -U homeboard -d homeboard_dev

# Backup development database
docker-compose -f docker-compose.dev.yml exec postgres-dev pg_dump -U homeboard homeboard_dev > backup.sql
```

## Production Deployment

### Build Optimization
```bash
# Build production image
docker build -t homeboard:latest .

# Multi-platform build
docker buildx build --platform linux/amd64,linux/arm64 -t homeboard:latest .
```

### Security Configuration
1. **SSL/TLS Setup**:
   ```bash
   # Generate self-signed certificates for testing
   mkdir -p docker/nginx/ssl
   openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
     -keyout docker/nginx/ssl/key.pem \
     -out docker/nginx/ssl/cert.pem
   ```

2. **Secrets Management**:
   ```bash
   # Use Docker secrets in production
   echo "your-redis-password" | docker secret create redis_password -
   echo "your-grafana-password" | docker secret create grafana_password -
   ```

3. **Network Security**:
   - Configure firewall rules
   - Use Docker networks for service isolation
   - Enable nginx rate limiting

### Resource Limits
```yaml
# Add to docker-compose.yml services
deploy:
  resources:
    limits:
      cpus: '1.0'
      memory: 512M
    reservations:
      cpus: '0.5'
      memory: 256M
```

## Monitoring and Observability

### Health Checks
All services include health checks:
```bash
# Check service health
docker-compose ps

# View health check logs
docker inspect homeboard-app | grep -A 20 Health
```

### Metrics Collection
Prometheus collects metrics from:
- Homeboard application
- Nginx reverse proxy
- Redis cache
- Container runtime (cAdvisor)
- System metrics (node-exporter)

### Log Management
```bash
# View application logs
docker-compose logs -f homeboard

# Export logs for analysis
docker-compose logs homeboard > app.log

# Configure log rotation
# Add to docker-compose.yml:
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

## Backup and Recovery

### Configuration Backup
```bash
# Backup configuration and data
docker run --rm -v homeboard-data:/source:ro -v $(pwd):/backup alpine \
  tar -czf /backup/homeboard-backup-$(date +%Y%m%d).tar.gz /source

# Restore from backup
docker run --rm -v homeboard-data:/target -v $(pwd):/backup alpine \
  tar -xzf /backup/homeboard-backup-YYYYMMDD.tar.gz -C /target --strip-components=1
```

### Database Backup (Development)
```bash
# Automated backup script
docker-compose -f docker-compose.dev.yml exec postgres-dev \
  pg_dump -U homeboard homeboard_dev | gzip > backup-$(date +%Y%m%d).sql.gz
```

## Troubleshooting

### Common Issues

1. **Port Conflicts**:
   ```bash
   # Check port usage
   netstat -tulpn | grep :8081
   
   # Use different ports
   docker-compose up --build --scale homeboard=1 -p 8082:8081
   ```

2. **Permission Errors**:
   ```bash
   # Fix volume permissions
   docker-compose exec homeboard chown -R homeboard:homeboard /app/data
   ```

3. **Memory Issues**:
   ```bash
   # Monitor container resources
   docker stats
   
   # Increase memory limits in docker-compose.yml
   ```

4. **Widget Failures**:
   ```bash
   # Test widget directly in container
   docker-compose exec homeboard python3 /app/widgets/clock.py '{}'
   
   # Check Python dependencies
   docker-compose exec homeboard pip list
   ```

### Debug Mode
```bash
# Enable debug mode
export DEBUG=true
docker-compose up --build

# Access container shell
docker-compose exec homeboard /bin/sh

# View detailed logs
docker-compose logs --tail=100 -f homeboard
```

## Performance Tuning

### Container Optimization
- Use multi-stage builds for smaller images
- Implement proper health checks
- Configure resource limits
- Use nginx caching for static content

### Network Optimization
- Use Docker networks for service communication
- Configure nginx compression
- Implement proper rate limiting
- Use connection pooling

### Storage Optimization
- Use volume mounts for persistent data
- Implement log rotation
- Regular cleanup of unused images and volumes

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Docker Build and Deploy
on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Build Docker image
        run: docker build -t homeboard:${{ github.sha }} .
      - name: Run tests
        run: docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

### Deployment Automation
```bash
# Automated deployment script
#!/bin/bash
docker-compose pull
docker-compose up -d --remove-orphans
docker system prune -f
```

## Security Best Practices

1. **Container Security**:
   - Run containers as non-root user
   - Use minimal base images
   - Regular security updates
   - Scan images for vulnerabilities

2. **Network Security**:
   - Use Docker networks
   - Implement proper firewall rules
   - Enable SSL/TLS encryption
   - Configure rate limiting

3. **Data Security**:
   - Encrypt sensitive data
   - Use Docker secrets
   - Regular backups
   - Access control and auditing

---

For more information, see the main [README.md](README.md) and [deployment documentation](INSTALL.md).