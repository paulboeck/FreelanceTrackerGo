# Deployment Guide

This guide covers deploying the FreelanceTracker API to Fly.io and accessing the API documentation.

---

## API Documentation

### Viewing the OpenAPI Specification

The complete OpenAPI 3.0 specification is available at `docs/openapi.yaml`.

### Using Swagger UI (Local)

1. **Online Swagger Editor**:
   ```bash
   # Open the Swagger Editor
   open https://editor.swagger.io/

   # Then paste the contents of docs/openapi.yaml
   ```

2. **Local Swagger UI with Docker**:
   ```bash
   docker run -p 8081:8080 \
     -e SWAGGER_JSON=/openapi.yaml \
     -v $(pwd)/docs/openapi.yaml:/openapi.yaml \
     swaggerapi/swagger-ui

   # Open http://localhost:8081
   ```

3. **Using Redoc** (Alternative):
   ```bash
   npx @redocly/cli preview-docs docs/openapi.yaml
   ```

### API Endpoints Summary

**Base URL**: `http://localhost:8080/api/v1`

**Authentication**: All endpoints (except `/auth/login`) require:
```
Authorization: Bearer ftk_your_api_key_here
```

**35 Total Endpoints** across 7 categories:
- Authentication (4 endpoints)
- Clients (6 endpoints)
- Projects (8 endpoints)
- Timesheets (4 endpoints)
- Invoices (7 endpoints)
- Reports (1 endpoint)
- Settings (4 endpoints)

---

## Deployment to Fly.io

### Prerequisites

1. **Install Fly.io CLI**:
   ```bash
   # macOS
   brew install flyctl

   # Linux
   curl -L https://fly.io/install.sh | sh

   # Windows
   powershell -Command "iwr https://fly.io/install.ps1 -useb | iex"
   ```

2. **Login to Fly.io**:
   ```bash
   flyctl auth login
   ```

3. **Verify Installation**:
   ```bash
   flyctl version
   ```

### Step 1: Create Fly.io Application

```bash
# Initialize Fly.io app (run in project root)
flyctl launch --name freelancetracker-api --region sjc

# This will:
# - Create a fly.toml configuration file
# - Set up a Fly.io app
# - Ask about PostgreSQL (say NO - we use SQLite)
```

### Step 2: Create fly.toml Configuration

Create `fly.toml` in the project root:

```toml
app = "freelancetracker-api"
primary_region = "sjc"

[build]
  [build.args]
    GO_VERSION = "1.21"

[env]
  PORT = "8080"
  APP_ENV = "production"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 0
  processes = ["app"]

[[http_service.checks]]
  grace_period = "10s"
  interval = "30s"
  method = "GET"
  timeout = "5s"
  path = "/health"

[mounts]
  source = "data"
  destination = "/data"

[[vm]]
  cpu_kind = "shared"
  cpus = 1
  memory_mb = 256
```

### Step 3: Create Dockerfile

Create `Dockerfile` in the project root:

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/web

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Copy UI assets
COPY --from=builder /app/ui ./ui

# Create data directory for SQLite
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main", "-addr=:8080", "-dsn=/data/freelancetracker.db"]
```

### Step 4: Create Health Check Endpoint

Add to `cmd/web/routes.go`:

```go
// Health check endpoint for Fly.io
router.GET("/health", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
        "version": "1.0.0",
    })
})
```

### Step 5: Create Persistent Volume

```bash
# Create a volume for SQLite database
flyctl volumes create data --region sjc --size 1

# This ensures your database persists across deployments
```

### Step 6: Set Environment Variables (Secrets)

```bash
# Set production secrets
flyctl secrets set SESSION_SECRET=$(openssl rand -base64 32)
flyctl secrets set ENCRYPTION_KEY=$(openssl rand -base64 32)

# Optional: Set SMTP credentials for email
flyctl secrets set SMTP_HOST=smtp.gmail.com
flyctl secrets set SMTP_PORT=587
flyctl secrets set SMTP_USERNAME=your_email@gmail.com
flyctl secrets set SMTP_PASSWORD=your_app_password
```

### Step 7: Deploy

```bash
# Deploy to Fly.io
flyctl deploy

# Monitor deployment
flyctl logs

# Check status
flyctl status
```

### Step 8: Create First User

```bash
# SSH into the machine
flyctl ssh console

# Inside the container, create a user (you'll need to add a CLI command for this)
# Or use the web UI at https://your-app.fly.dev
```

---

## Post-Deployment Configuration

### 1. Custom Domain (Optional)

```bash
# Add custom domain
flyctl certs add api.yourdomain.com

# Get DNS instructions
flyctl certs show api.yourdomain.com

# Add the provided CNAME or A record to your DNS provider
```

### 2. Scaling

```bash
# Scale to multiple regions
flyctl regions add lax
flyctl regions add ord

# Scale machines
flyctl scale count 2

# Scale memory
flyctl scale memory 512

# Scale CPU
flyctl scale vm dedicated-cpu-1x
```

### 3. Monitoring

```bash
# View logs
flyctl logs

# View metrics
flyctl dashboard

# Open Grafana dashboard
flyctl dashboard metrics
```

### 4. Backup Strategy

```bash
# Create volume snapshot
flyctl volumes snapshots create data

# List snapshots
flyctl volumes snapshots list data

# Restore from snapshot (if needed)
flyctl volumes create data_restore --snapshot-id <snapshot-id>
```

---

## Production Checklist

### Security

- [ ] Change default admin password
- [ ] Rotate API keys regularly
- [ ] Enable CORS for specific domains only
- [ ] Set up rate limiting (already configured)
- [ ] Review and limit API key scopes
- [ ] Enable HTTPS only (Fly.io does this automatically)

### Monitoring

- [ ] Set up error tracking (Sentry, Rollbar, etc.)
- [ ] Configure log aggregation
- [ ] Set up uptime monitoring (UptimeRobot, Pingdom, etc.)
- [ ] Create alerts for critical errors
- [ ] Monitor database size

### Performance

- [ ] Enable database connection pooling
- [ ] Monitor API response times
- [ ] Set up caching where appropriate
- [ ] Optimize database queries if needed
- [ ] Consider CDN for static assets

### Backup & Recovery

- [ ] Schedule regular volume snapshots
- [ ] Test backup restoration process
- [ ] Document recovery procedures
- [ ] Set up off-site backups
- [ ] Create disaster recovery plan

---

## Environment Variables

### Required

```bash
SESSION_SECRET=<random-string>    # For session management
ENCRYPTION_KEY=<random-string>    # For sensitive data encryption
PORT=8080                         # Application port
```

### Optional (Email)

```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM_NAME="FreelanceTracker"
```

### Optional (Application)

```bash
APP_ENV=production
LOG_LEVEL=info
RATE_LIMIT=100                    # Requests per 2 minutes
CORS_ORIGINS=https://yourdomain.com
```

---

## Troubleshooting

### Application Won't Start

```bash
# Check logs
flyctl logs

# SSH into machine
flyctl ssh console

# Check disk space
df -h

# Check database file
ls -lah /data/
```

### Database Issues

```bash
# SSH into machine
flyctl ssh console

# Check database file
cd /data
sqlite3 freelancetracker.db

# Run SQL commands
.tables
SELECT COUNT(*) FROM client;
.quit
```

### Performance Issues

```bash
# View metrics
flyctl dashboard metrics

# Scale up memory
flyctl scale memory 512

# Add more regions
flyctl regions add lax
```

### Connection Errors

```bash
# Check if app is running
flyctl status

# Check health endpoint
curl https://your-app.fly.dev/health

# Restart app
flyctl apps restart
```

---

## Cost Estimation

### Fly.io Free Tier

- 3 shared-cpu-1x VMs with 256MB RAM (enough for API)
- 3GB persistent volume storage
- 160GB outbound data transfer

### Estimated Monthly Cost (Beyond Free Tier)

- **Small** (1 VM, 256MB): $0 (within free tier)
- **Medium** (1 VM, 512MB): ~$5-10/month
- **Production** (2 VMs, 512MB each, 3 regions): ~$20-30/month

---

## Support & Resources

- **Fly.io Documentation**: https://fly.io/docs/
- **Fly.io Community**: https://community.fly.io/
- **Project Repository**: https://github.com/your-username/freelancetracker
- **API Documentation**: https://your-app.fly.dev/api/docs (after deployment)

---

## Next Steps After Deployment

1. **Test all endpoints** using the OpenAPI spec
2. **Create your first client** via the API
3. **Set up monitoring** and alerts
4. **Configure backups** schedule
5. **Update DNS** if using custom domain
6. **Share API documentation** with your team

---

**Last Updated**: 2025-11-09
