# Deploying the Demo App to Fly.io

This guide explains how to set up and deploy the demo instance of FreelanceTrackerGo to Fly.io.

## Overview

The demo app is a separate instance with:
- **App name:** `freelance-tracker-demo`
- **Volume:** `freelance_demo_data` (512MB)
- **Configuration:** `fly.demo.toml`
- **Auto-deploy:** Pushes to `main` branch trigger deployment (both production and demo deploy together)

## Initial Setup

### Step 1: Create the Demo App

```bash
# Create the demo app (one-time setup)
flyctl apps create freelance-tracker-demo --org personal
```

### Step 2: Initial Deployment

Deploy the demo app for the first time:

```bash
# Deploy using the demo configuration
flyctl deploy --config fly.demo.toml
```

This will:
- Build the Docker image
- Deploy to `freelance-tracker-demo` app
- Create a fresh SQLite database on the volume
- Run database migrations automatically

### Step 3: Verify Deployment

```bash
# Check app status
flyctl status --app freelance-tracker-demo

# View logs
flyctl logs --app freelance-tracker-demo

# Open the demo app in browser
flyctl open --app freelance-tracker-demo
```

## Continuous Deployment

Once the initial setup is complete, the demo app auto-deploys via GitHub Actions.

### Automatic Deployment

Both the production and demo apps deploy automatically when you push to the `main` branch:

```bash
# Make changes and commit
git add .
git commit -m "Update application"

# Push to main to deploy both apps
git push origin main
```

The GitHub Actions will deploy in parallel:
- `.github/workflows/fly-deploy.yml` deploys to `freelance-tracker` (production)
- `.github/workflows/fly-demo-deploy.yml` deploys to `freelance-tracker-demo` (demo)

Both workflows:
1. Checkout code from `main` branch
2. Build Docker image
3. Deploy to their respective Fly.io apps

### Manual Deployment

You can also deploy manually at any time:

```bash
# Deploy from current directory
flyctl deploy --config fly.demo.toml
```

## Database Management

### Initialize with Fresh Data

The demo app starts with an empty database. To populate it:

```bash
# Option 1: Use the web interface
flyctl open --app freelance-tracker-demo
# Then manually create test clients, projects, etc.

# Option 2: Upload a pre-populated database (see below)
```

### Upload Database to Demo App

To upload a local SQLite database to the demo app:

```bash
# 1. Backup existing demo database (if any)
flyctl ssh console --app freelance-tracker-demo \
  -C "cp /data/freelance_tracker.db /data/freelance_tracker.backup.$(date +%Y%m%d).db"

# 2. Scale down the demo app
flyctl scale count 0 --app freelance-tracker-demo

# 3. Upload your database using SFTP
flyctl ssh sftp shell --app freelance-tracker-demo
# In SFTP session:
cd /data
put ./freelance_tracker.db
ls -lh
exit

# 4. Verify upload
flyctl ssh console --app freelance-tracker-demo \
  -C "ls -lh /data/freelance_tracker.db"

# 5. Scale back up
flyctl scale count 1 --app freelance-tracker-demo

# 6. Verify application is working
flyctl open --app freelance-tracker-demo
```

### View Demo Database

```bash
# SSH into demo app
flyctl ssh console --app freelance-tracker-demo

# Query database
sqlite3 /data/freelance_tracker.db "SELECT COUNT(*) FROM client;"
sqlite3 /data/freelance_tracker.db "SELECT COUNT(*) FROM project;"

# Exit
exit
```

### Backup Demo Database

```bash
# Create backup
flyctl ssh console --app freelance-tracker-demo \
  -C "cp /data/freelance_tracker.db /data/freelance_tracker.backup.$(date +%Y%m%d).db"

# Download backup to local machine
flyctl ssh sftp shell --app freelance-tracker-demo
cd /data
get freelance_tracker.db ./demo_backup.db
exit
```

### Reset Demo Database

To start fresh:

```bash
# Scale down
flyctl scale count 0 --app freelance-tracker-demo

# Delete database
flyctl ssh console --app freelance-tracker-demo \
  -C "rm /data/freelance_tracker.db"

# Scale back up (migrations will create fresh database)
flyctl scale count 1 --app freelance-tracker-demo
```

## Monitoring

### View Logs

```bash
# Real-time logs
flyctl logs --app freelance-tracker-demo

# Filter logs
flyctl logs --app freelance-tracker-demo | grep ERROR
```

### Check Status

```bash
# App status
flyctl status --app freelance-tracker-demo

# Volume info
flyctl volumes list --app freelance-tracker-demo

# Machine info
flyctl machine list --app freelance-tracker-demo
```

### Check Resource Usage

```bash
# View app metrics
flyctl dashboard --app freelance-tracker-demo
```

## Configuration Differences

The demo app has identical configuration to production except:

| Setting | Production | Demo |
|---------|------------|------|
| App name | `freelance-tracker` | `freelance-tracker-demo` |
| Volume name | `freelance_data` | `freelance_demo_data` |
| Volume size | 1GB | 512MB |
| Auto-deploy branch | `main` | `main` |
| Config file | `fly.toml` | `fly.demo.toml` |

Both apps use:
- Same Dockerfile
- 256MB RAM
- Auto-stop/auto-start enabled
- Same build process
- Same application code

## Cost Management

Both apps stay within Fly.io free tier:
- ✅ 2 apps × 256MB RAM = 512MB (under 768MB limit for 3 VMs)
- ✅ 1GB + 512MB = 1.5GB volume storage (under 3GB limit)
- ✅ Auto-stop/start enabled (only charges when active)

**Tips to minimize costs:**
- Demo app auto-stops when inactive (included by default)
- Use `flyctl scale count 0` when not needed for extended periods
- Monitor usage: `flyctl dashboard`

## Troubleshooting

### Demo App Won't Start

Check logs for errors:
```bash
flyctl logs --app freelance-tracker-demo
```

Common issues:
- Volume not attached: Verify volume name matches `fly.demo.toml`
- Out of memory: Check memory settings (should be 256mb)
- Database locked: Scale down and restart

### GitHub Action Fails

1. Check that `FLY_API_TOKEN` secret is set in GitHub repo settings
2. Verify the token has access to `freelance-tracker-demo` app
3. Check workflow logs in GitHub Actions tab

### Volume Full

The 512MB volume should be sufficient for demo purposes. If needed:

```bash
# Check volume usage
flyctl ssh console --app freelance-tracker-demo \
  -C "df -h /data"

# Check database size
flyctl ssh console --app freelance-tracker-demo \
  -C "ls -lh /data/freelance_tracker.db"
```

If volume is full, you can:
1. Clean up old backups
2. Reset the database
3. Resize the volume (will incur additional charges if over free tier)

## Deleting Demo App

If you need to remove the demo app:

```bash
# Delete the app (this also deletes the volume)
flyctl apps destroy freelance-tracker-demo

# Confirm deletion when prompted
```

**Warning:** This permanently deletes all data. Backup first if needed.

## Reference

- **Production app:** `freelance-tracker`
- **Demo app:** `freelance-tracker-demo`
- **Production config:** `fly.toml`
- **Demo config:** `fly.demo.toml`
- **Production workflow:** `.github/workflows/fly-deploy.yml`
- **Demo workflow:** `.github/workflows/fly-demo-deploy.yml`
- **Deployment guide:** `.github/DEPLOY_SQLITE_TO_FLY.md`

## Next Steps

After initial setup:

1. ✅ Create demo app and volume on Fly.io
2. ✅ Deploy demo app manually once
3. ✅ Push to `main` branch to test auto-deployment (both apps deploy)
4. ✅ Populate demo database with test data
5. ✅ Share demo URL with stakeholders
6. ✅ Set up regular demo resets (optional)

For more details on database operations, see [DEPLOY_SQLITE_TO_FLY.md](DEPLOY_SQLITE_TO_FLY.md).
