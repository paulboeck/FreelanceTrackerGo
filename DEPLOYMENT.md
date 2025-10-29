# Fly.io Deployment Guide

This guide walks you through deploying FreelanceTrackerGo to Fly.io's free tier.

## Prerequisites

1. **Install flyctl** (Fly.io CLI):
   ```bash
   # macOS
   brew install flyctl

   # Linux/macOS (alternative)
   curl -L https://fly.io/install.sh | sh

   # Windows (PowerShell)
   powershell -Command "iwr https://fly.io/install.ps1 -useb | iex"
   ```

2. **Create Fly.io account** (if you don't have one):
   ```bash
   fly auth signup
   ```

   Or login if you already have an account:
   ```bash
   fly auth login
   ```

## Initial Deployment

### Step 1: Launch the App

From your project root directory, run:

```bash
fly launch --no-deploy
```

When prompted:
- **App name**: Accept the suggested name or choose your own (must be unique across Fly.io)
- **Region**: Choose the region closest to you or your users (e.g., `iad` for US East)
- **Database**: Select "No" - we're using SQLite with a persistent volume
- **Overwrite fly.toml**: Select "No" - we already have a configured one

### Step 2: Create Persistent Volume for SQLite

The SQLite database needs a persistent volume to survive restarts:

```bash
fly volumes create freelance_data --region iad --size 1
```

**Important**: Replace `iad` with whatever region you chose in Step 1.

### Step 3: Deploy the Application

```bash
fly deploy
```

This will:
- Build your Docker image
- Push it to Fly.io
- Create a machine with the persistent volume mounted
- Start your application

### Step 4: Verify Deployment

```bash
# Check app status
fly status

# View recent logs
fly logs

# Open app in browser
fly open
```

## Free Tier Details

Your deployment includes:
- **Compute**: 3 shared-cpu-1x VMs with 256MB RAM each
- **Storage**: 3GB total persistent volume storage (we're using 1GB)
- **Bandwidth**: 160GB outbound per month
- **Cost**: $0/month within free tier limits

## Useful Commands

### Viewing Logs
```bash
# Live tail logs
fly logs

# Filter logs
fly logs --region iad
```

### SSH into Machine
```bash
fly ssh console
```

### Scaling
```bash
# Scale memory (costs money above free tier)
fly scale memory 512

# Add more machines (costs money above free tier)
fly scale count 2
```

### Managing Volumes
```bash
# List volumes
fly volumes list

# Create snapshot (backup)
fly volumes snapshots create freelance_data

# List snapshots
fly volumes snapshots list
```

### App Management
```bash
# Stop app (keeps volume, no charges)
fly apps stop

# Restart app
fly apps restart

# Destroy app completely (WARNING: deletes everything)
fly apps destroy freelance-tracker-go
```

## Configuration Details

### fly.toml Overview

Key settings in your `fly.toml`:

- **auto_stop_machines/auto_start_machines**: App sleeps when idle, wakes on request (free tier friendly)
- **min_machines_running = 0**: Allows app to fully stop when idle (saves resources)
- **memory = '256mb'**: Fits within free tier limits
- **mounts**: Persistent volume for SQLite database at `/data`

### Environment Variables

If you need to set environment variables (e.g., for email configuration):

```bash
fly secrets set SMTP_PASSWORD="your-password"
fly secrets set API_KEY="your-key"
```

List secrets:
```bash
fly secrets list
```

## Troubleshooting

### App Won't Start

Check logs:
```bash
fly logs
```

Common issues:
- **Volume not mounted**: Verify volume exists with `fly volumes list`
- **Port mismatch**: Ensure app listens on port 8080
- **Database permissions**: Check logs for SQLite permission errors

### Volume Issues

If volume isn't working:
```bash
# List volumes
fly volumes list

# Delete and recreate (WARNING: loses data)
fly volumes delete freelance_data
fly volumes create freelance_data --region iad --size 1
fly deploy
```

### Slow Cold Starts

Free tier machines sleep after inactivity. First request after sleep takes ~5-10 seconds. This is normal for free tier.

To keep app always running (uses more free tier resources):
```bash
fly scale count 1 --max-per-region 1
```

Edit `fly.toml` and change:
```toml
min_machines_running = 1
```

Then redeploy:
```bash
fly deploy
```

## Updating Your App

After making code changes:

```bash
# Deploy updates
fly deploy

# Or force rebuild
fly deploy --no-cache
```

## Monitoring

View app dashboard:
```bash
fly dashboard
```

Or visit: https://fly.io/apps/your-app-name

## Cost Management

Monitor usage:
```bash
fly dashboard
```

The free tier includes:
- Up to 3 shared-cpu-1x machines (256MB RAM each)
- 3GB persistent volumes
- 160GB bandwidth/month

**Staying within free tier**:
- Use `auto_stop_machines = true` (already configured)
- Keep volume ≤ 3GB total
- Monitor bandwidth usage in dashboard

## Backup Strategy

### Manual Backups

1. Create volume snapshot:
   ```bash
   fly volumes snapshots create freelance_data
   ```

2. Or download database file:
   ```bash
   fly ssh console
   cat /data/freelance_tracker.db > /tmp/backup.db
   exit

   fly ssh sftp get /tmp/backup.db ./local-backup.db
   ```

### Restore from Backup

```bash
# Restore from snapshot when creating new volume
fly volumes create freelance_data --snapshot-id snap_xxx --region iad
```

## Custom Domain (Optional)

Add your own domain:

```bash
fly certs create yourdomain.com
fly certs create www.yourdomain.com
```

Then add DNS records as instructed by Fly.io.

## Security Notes

- HTTPS is enforced by default (handled by Fly.io proxy)
- App runs as non-root user inside container
- Database is isolated on persistent volume
- Use `fly secrets` for sensitive configuration (never commit secrets to git)

## Support

- Fly.io docs: https://fly.io/docs/
- Community forum: https://community.fly.io/
- Status page: https://status.flyio.net/

## Quick Reference

```bash
# Deploy app
fly deploy

# Check status
fly status

# View logs
fly logs

# Open in browser
fly open

# SSH into machine
fly ssh console

# List volumes
fly volumes list

# Stop app
fly apps stop

# Restart app
fly apps restart
```
