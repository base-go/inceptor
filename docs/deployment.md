# Deployment Guide

This guide covers deploying Inceptor to various environments.

## Deployment Options

| Method | Best For | Complexity |
|--------|----------|------------|
| BasePod | Quick deployment, managed hosting | Low |
| Docker | Self-hosted, containerized environments | Medium |
| Kubernetes | High availability, scaling | High |
| Binary | Bare metal, VMs | Medium |

---

## BasePod Deployment (Recommended)

BasePod provides the simplest deployment experience with automatic TLS, health checks, and persistent storage.

### Prerequisites

- BasePod CLI installed (`bp`)
- BasePod account at [pod.base.al](https://pod.base.al)

### Deploy

```bash
# Clone the repository
git clone https://github.com/base-go/inceptor.git
cd inceptor

# Deploy to BasePod
bp push
```

### Configuration

Edit `basepod.yaml` to customize:

```yaml
name: inceptor
port: 8080

build:
  dockerfile: Dockerfile
  context: .

env:
  INCEPTOR_SERVER_HOST: "0.0.0.0"
  INCEPTOR_SERVER_REST_PORT: "8080"
  INCEPTOR_AUTH_ENABLED: "true"
  INCEPTOR_AUTH_ADMIN_KEY: "your-secure-admin-key"
  INCEPTOR_RETENTION_DEFAULT_DAYS: "30"

volumes:
  - /app/data

health:
  path: /health
  interval: 30s

resources:
  memory: 512M
  cpu: 1
```

### Update

```bash
# Deploy new version
bp push

# View logs
bp logs inceptor

# Restart
bp restart inceptor
```

---

## Docker Deployment

### Build the Image

```bash
docker build -t inceptor:latest .
```

### Run with Docker

```bash
docker run -d \
  --name inceptor \
  -p 8080:8080 \
  -v inceptor-data:/app/data \
  -e INCEPTOR_AUTH_ADMIN_KEY=your-secure-key \
  -e INCEPTOR_AUTH_ENABLED=true \
  inceptor:latest
```

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  inceptor:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - inceptor-data:/app/data
    environment:
      - INCEPTOR_SERVER_HOST=0.0.0.0
      - INCEPTOR_SERVER_REST_PORT=8080
      - INCEPTOR_AUTH_ENABLED=true
      - INCEPTOR_AUTH_ADMIN_KEY=${ADMIN_KEY:-changeme}
      - INCEPTOR_RETENTION_DEFAULT_DAYS=30
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
    restart: unless-stopped

volumes:
  inceptor-data:
```

Run:

```bash
ADMIN_KEY=your-secure-key docker-compose up -d
```

---

## Kubernetes Deployment

### Deployment Manifest

```yaml
# inceptor-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: inceptor
  labels:
    app: inceptor
spec:
  replicas: 1  # Single instance for SQLite
  selector:
    matchLabels:
      app: inceptor
  template:
    metadata:
      labels:
        app: inceptor
    spec:
      containers:
      - name: inceptor
        image: your-registry/inceptor:latest
        ports:
        - containerPort: 8080
        env:
        - name: INCEPTOR_SERVER_HOST
          value: "0.0.0.0"
        - name: INCEPTOR_AUTH_ENABLED
          value: "true"
        - name: INCEPTOR_AUTH_ADMIN_KEY
          valueFrom:
            secretKeyRef:
              name: inceptor-secrets
              key: admin-key
        volumeMounts:
        - name: data
          mountPath: /app/data
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "1000m"
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: inceptor-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: inceptor
spec:
  selector:
    app: inceptor
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: inceptor-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
apiVersion: v1
kind: Secret
metadata:
  name: inceptor-secrets
type: Opaque
stringData:
  admin-key: "your-secure-admin-key"
```

Deploy:

```bash
kubectl apply -f inceptor-deployment.yaml
```

### Ingress with TLS

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: inceptor-ingress
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - crashes.yourdomain.com
    secretName: inceptor-tls
  rules:
  - host: crashes.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: inceptor
            port:
              number: 80
```

---

## Binary Deployment

### Build from Source

```bash
# Clone repository
git clone https://github.com/base-go/inceptor.git
cd inceptor

# Build dashboard
cd web && npm install && npm run generate && cd ..
cp -r web/.output/public internal/api/rest/static/

# Build binary
go build -o inceptor ./cmd/inceptor

# Run
./inceptor --config configs/config.yaml
```

### Systemd Service

Create `/etc/systemd/system/inceptor.service`:

```ini
[Unit]
Description=Inceptor Crash Logging Service
After=network.target

[Service]
Type=simple
User=inceptor
Group=inceptor
WorkingDirectory=/opt/inceptor
ExecStart=/opt/inceptor/inceptor --config /opt/inceptor/config.yaml
Restart=always
RestartSec=5

# Environment
Environment=INCEPTOR_AUTH_ADMIN_KEY=your-secure-key

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/inceptor/data

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable inceptor
sudo systemctl start inceptor
```

---

## Reverse Proxy Configuration

### Nginx

```nginx
upstream inceptor {
    server 127.0.0.1:8080;
}

server {
    listen 443 ssl http2;
    server_name crashes.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/crashes.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/crashes.yourdomain.com/privkey.pem;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=crash_limit:10m rate=100r/s;

    location / {
        proxy_pass http://inceptor;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /api/v1/crashes {
        limit_req zone=crash_limit burst=200 nodelay;
        proxy_pass http://inceptor;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Caddy

```
crashes.yourdomain.com {
    reverse_proxy localhost:8080

    # Rate limiting (requires rate_limit plugin)
    rate_limit {
        zone crash_zone {
            key {remote_host}
            events 100
            window 1s
        }
    }
}
```

---

## Production Checklist

### Security

- [ ] Set a strong `INCEPTOR_AUTH_ADMIN_KEY` (use `openssl rand -hex 32`)
- [ ] Enable TLS (via reverse proxy or BasePod)
- [ ] Configure firewall rules (only expose port 443)
- [ ] Set appropriate file permissions on data directory
- [ ] Regularly rotate API keys

### Reliability

- [ ] Configure health checks
- [ ] Set up monitoring/alerting
- [ ] Configure automatic restarts
- [ ] Set resource limits (memory, CPU)

### Data Management

- [ ] Set appropriate retention periods
- [ ] Configure backup strategy for `/app/data/`
- [ ] Test backup restoration
- [ ] Plan for storage growth

### Monitoring

Recommended metrics to track:
- Crash submission rate
- API response times
- Disk usage
- Memory usage
- Error rates

### Backup Strategy

The data directory (`/app/data/`) contains:
- `inceptor.db` - SQLite database
- `crashes/` - Full crash log files

Backup options:

```bash
# Simple backup
tar -czf inceptor-backup-$(date +%Y%m%d).tar.gz /app/data/

# S3 backup
aws s3 sync /app/data/ s3://your-bucket/inceptor-backup/

# Database-specific backup (while running)
sqlite3 /app/data/inceptor.db ".backup '/backup/inceptor-$(date +%Y%m%d).db'"
```

---

## Troubleshooting

### Container Won't Start

Check logs:
```bash
docker logs inceptor
# or
bp logs inceptor
```

Common issues:
- Missing environment variables
- Data directory permissions
- Port already in use

### Database Locked

SQLite doesn't support concurrent writers. If you see "database is locked":
- Ensure only one instance is running
- Check for hung processes: `fuser /app/data/inceptor.db`

### High Memory Usage

- Check retention settings (too long = more data)
- Monitor crash volume
- Consider increasing cleanup frequency

### Slow Dashboard

- Enable browser caching via reverse proxy
- Check database indexes (migrations should create them)
- Consider reducing the number of crashes displayed per page
