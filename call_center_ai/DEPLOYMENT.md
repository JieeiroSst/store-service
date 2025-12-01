# 🚀 Production Deployment Guide

Hướng dẫn deploy hệ thống Call Center AI lên production environment.

## 📋 Checklist trước khi deploy

- [ ] Đã test kỹ trên môi trường development
- [ ] Đã có domain name hoặc server IP tĩnh
- [ ] Đã có SSL certificate (Let's Encrypt hoặc mua)
- [ ] Đã cấu hình firewall và security groups
- [ ] Đã backup database
- [ ] Đã setup monitoring và logging
- [ ] Đã có plan cho disaster recovery

---

## 🔧 Option 1: Deploy trên VPS/Cloud Server

### 1. Chuẩn bị Server

**Minimum requirements:**
- OS: Ubuntu 22.04 LTS hoặc CentOS 8+
- RAM: 2GB+
- CPU: 2 cores+
- Storage: 20GB+ SSD
- Network: Static IP

**Recommended:**
- RAM: 4GB+
- CPU: 4 cores+
- Storage: 50GB+ SSD

### 2. Setup Server

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install required packages
sudo apt install -y \
    python3.11 \
    python3.11-venv \
    python3-pip \
    mysql-server \
    nginx \
    certbot \
    python3-certbot-nginx \
    git \
    supervisor

# Clone repository
cd /opt
sudo git clone <your-repo-url> call_center_ai
cd call_center_ai

# Set permissions
sudo chown -R www-data:www-data /opt/call_center_ai
```

### 3. Setup MySQL

```bash
# Secure MySQL
sudo mysql_secure_installation

# Create database
sudo mysql -u root -p <<EOF
CREATE DATABASE call_center_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'callcenter'@'localhost' IDENTIFIED BY 'STRONG_PASSWORD_HERE';
GRANT ALL PRIVILEGES ON call_center_db.* TO 'callcenter'@'localhost';
FLUSH PRIVILEGES;
EOF
```

### 4. Setup Application

```bash
# Create virtual environment
sudo -u www-data python3.11 -m venv venv
sudo -u www-data venv/bin/pip install --upgrade pip
sudo -u www-data venv/bin/pip install -r requirements.txt

# Configure environment
sudo cp .env.example .env
sudo nano .env  # Edit with production values

# Initialize database
sudo -u www-data venv/bin/python init_scenarios.py
```

### 5. Setup Supervisor (Process Manager)

```bash
sudo nano /etc/supervisor/conf.d/callcenter.conf
```

Nội dung file:

```ini
[program:callcenter]
command=/opt/call_center_ai/venv/bin/uvicorn main:app --host 0.0.0.0 --port 8000 --workers 4
directory=/opt/call_center_ai
user=www-data
autostart=true
autorestart=true
redirect_stderr=true
stdout_logfile=/var/log/callcenter/app.log
stdout_logfile_maxbytes=50MB
stdout_logfile_backups=10
environment=PATH="/opt/call_center_ai/venv/bin"
```

```bash
# Create log directory
sudo mkdir -p /var/log/callcenter
sudo chown www-data:www-data /var/log/callcenter

# Reload supervisor
sudo supervisorctl reread
sudo supervisorctl update
sudo supervisorctl start callcenter
```

### 6. Setup Nginx (Reverse Proxy)

```bash
sudo nano /etc/nginx/sites-available/callcenter
```

Nội dung file:

```nginx
upstream callcenter_app {
    server 127.0.0.1:8000;
}

server {
    listen 80;
    server_name your-domain.com www.your-domain.com;

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com www.your-domain.com;

    # SSL Configuration (will be added by certbot)
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Logging
    access_log /var/log/nginx/callcenter.access.log;
    error_log /var/log/nginx/callcenter.error.log;

    # Client body size limit
    client_max_body_size 10M;

    location / {
        proxy_pass http://callcenter_app;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket support (if needed)
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Static files (if any)
    location /static {
        alias /opt/call_center_ai/static;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # Health check endpoint
    location /health {
        access_log off;
        proxy_pass http://callcenter_app;
    }
}
```

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/callcenter /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Get SSL certificate
sudo certbot --nginx -d your-domain.com -d www.your-domain.com

# Reload nginx
sudo systemctl reload nginx
```

### 7. Setup Firewall

```bash
# UFW firewall
sudo ufw allow 22/tcp      # SSH
sudo ufw allow 80/tcp      # HTTP
sudo ufw allow 443/tcp     # HTTPS
sudo ufw enable

# Deny direct access to app port
sudo ufw deny 8000/tcp
```

### 8. Configure Twilio

1. Vào [Twilio Console](https://console.twilio.com)
2. Configure webhooks với production URL:
   - Voice webhook: `https://your-domain.com/voice/incoming`
   - Status callback: `https://your-domain.com/voice/status`

---

## 🐳 Option 2: Deploy với Docker

### 1. Chuẩn bị Server

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Add user to docker group
sudo usermod -aG docker $USER
```

### 2. Deploy với Docker Compose

```bash
# Clone repository
git clone <your-repo-url> call_center_ai
cd call_center_ai

# Configure .env
cp .env.example .env
nano .env  # Edit with production values

# Start services
docker-compose up -d

# Initialize scenarios
docker-compose exec api python init_scenarios.py

# View logs
docker-compose logs -f
```

### 3. Setup Nginx với Docker

Tạo `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - api
    networks:
      - call_center_network

  mysql:
    # ... existing config
    restart: always
    
  api:
    # ... existing config
    restart: always
```

---

## 🔐 Security Best Practices

### 1. Environment Variables

```bash
# Không bao giờ commit .env
# Sử dụng secrets management tools
# Ví dụ: AWS Secrets Manager, HashiCorp Vault

# .env production template
DB_PASSWORD=$(openssl rand -base64 32)
SECRET_KEY=$(openssl rand -hex 32)
```

### 2. Database Security

```sql
-- Chỉ cho phép local connections
GRANT ALL ON call_center_db.* TO 'callcenter'@'localhost';

-- Không cho phép remote root access
DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1');
FLUSH PRIVILEGES;
```

### 3. Application Security

```python
# Rate limiting (trong main.py)
from slowapi import Limiter, _rate_limit_exceeded_handler
from slowapi.util import get_remote_address

limiter = Limiter(key_func=get_remote_address)
app.state.limiter = limiter

@app.post("/voice/incoming")
@limiter.limit("100/minute")
async def handle_incoming_call(...):
    pass
```

### 4. SSL/TLS

```bash
# Auto-renew SSL certificates
sudo crontab -e

# Add line:
0 0 1 * * certbot renew --quiet
```

---

## 📊 Monitoring và Logging

### 1. Application Monitoring

**Sử dụng Prometheus + Grafana:**

```bash
# Install Prometheus
docker run -d \
  --name=prometheus \
  -p 9090:9090 \
  -v prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus

# Install Grafana
docker run -d \
  --name=grafana \
  -p 3000:3000 \
  grafana/grafana
```

### 2. Log Management

**Centralized logging với ELK Stack:**

```yaml
# docker-compose.logging.yml
services:
  elasticsearch:
    image: elasticsearch:8.5.0
    environment:
      - discovery.type=single-node
    
  logstash:
    image: logstash:8.5.0
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf
    
  kibana:
    image: kibana:8.5.0
    ports:
      - "5601:5601"
```

### 3. Health Checks

```bash
# Script kiểm tra health
#!/bin/bash
HEALTH_URL="https://your-domain.com/health"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $HEALTH_URL)

if [ $RESPONSE -ne 200 ]; then
    echo "Health check failed: $RESPONSE"
    # Send alert
    curl -X POST "https://hooks.slack.com/..." \
      -d '{"text":"Call Center API is down!"}'
fi
```

---

## 💾 Backup Strategy

### 1. Database Backup

```bash
# Daily backup script
#!/bin/bash
BACKUP_DIR="/backup/mysql"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

mysqldump -u callcenter -p call_center_db | gzip > $BACKUP_DIR/backup_$DATE.sql.gz

# Keep only last 30 days
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +30 -delete

# Upload to S3 (optional)
aws s3 cp $BACKUP_DIR/backup_$DATE.sql.gz s3://your-bucket/backups/
```

```bash
# Setup cron
sudo crontab -e

# Add daily backup at 2 AM
0 2 * * * /opt/scripts/backup_mysql.sh
```

### 2. Application Backup

```bash
# Backup code and configs
tar -czf /backup/app_$(date +%Y%m%d).tar.gz \
  /opt/call_center_ai \
  --exclude=venv \
  --exclude=__pycache__
```

---

## 📈 Performance Optimization

### 1. Database Optimization

```sql
-- Add indexes
CREATE INDEX idx_calls_from_number ON calls(from_number);
CREATE INDEX idx_calls_created_at ON calls(created_at);
CREATE INDEX idx_messages_call_id ON messages(call_id);
CREATE INDEX idx_messages_timestamp ON messages(timestamp);

-- Optimize queries
ANALYZE TABLE calls;
ANALYZE TABLE messages;
```

### 2. Application Optimization

```python
# Sử dụng connection pooling
engine = create_engine(
    DATABASE_URL,
    pool_size=20,          # Increase pool size
    max_overflow=40,       # More overflow connections
    pool_pre_ping=True,
    pool_recycle=3600
)

# Cache AI responses
from functools import lru_cache

@lru_cache(maxsize=100)
def get_cached_response(message: str):
    pass
```

### 3. CDN và Caching

```nginx
# Nginx caching
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=api_cache:10m max_size=1g;

location /api/ {
    proxy_cache api_cache;
    proxy_cache_valid 200 5m;
    proxy_cache_key "$request_uri";
}
```

---

## 🔄 CI/CD Pipeline

### GitHub Actions Example

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Deploy to server
        uses: appleboy/ssh-action@master
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: ${{ secrets.SERVER_USER }}
          key: ${{ secrets.SSH_KEY }}
          script: |
            cd /opt/call_center_ai
            git pull origin main
            source venv/bin/activate
            pip install -r requirements.txt
            sudo supervisorctl restart callcenter
```

---

## 🚨 Disaster Recovery

### 1. Backup Recovery

```bash
# Restore from backup
gunzip < backup_20240101.sql.gz | mysql -u callcenter -p call_center_db
```

### 2. Failover Plan

1. Setup standby server
2. Replicate database real-time
3. Use DNS for quick failover
4. Keep backups in multiple locations

---

## 📞 Post-Deployment Checklist

- [ ] SSL certificate đã active
- [ ] Twilio webhooks đã cấu hình
- [ ] Health checks đang hoạt động
- [ ] Logging đang ghi nhận
- [ ] Monitoring đang track metrics
- [ ] Backup đã schedule
- [ ] Firewall rules đã apply
- [ ] Load testing đã pass
- [ ] Documentation đã update
- [ ] Team đã training

---

## 🔧 Troubleshooting Production

### High CPU Usage

```bash
# Check processes
top
htop

# Check slow queries
mysqldumpslow /var/log/mysql/slow-query.log
```

### Memory Leaks

```bash
# Monitor memory
free -h
watch -n 1 free -m

# Check app memory
ps aux | grep uvicorn
```

### Connection Issues

```bash
# Check connections
netstat -an | grep 8000
ss -tuln | grep 8000

# Check MySQL connections
mysql -e "SHOW PROCESSLIST;"
```

---

**Production deployment hoàn tất! 🎉**

Hệ thống của bạn đã sẵn sàng cho traffic thực tế.
