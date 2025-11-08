# Deploying a Local SQLite Database to Fly.io

This guide explains how to migrate your MySQL data locally and then deploy the resulting SQLite database to Fly.io's persistent volume.

## Overview

The strategy is:
1. Run your migration locally (MySQL → SQLite)
2. Upload the SQLite database to Fly.io's volume
3. Restart your application to use the new database

## Prerequisites

- `flyctl` CLI installed and authenticated
- Local SQLite database with migrated data
- Fly.io app deployed with persistent volume

## Option 1: Upload Database Using fly ssh sftp (Recommended)

This is the safest and most straightforward approach.

### Step 1: Prepare Your Local Database

```bash
# Ensure you have your migrated SQLite database locally
# For example: ./freelance_tracker.db
ls -lh ./freelance_tracker.db
```

### Step 2: Backup Existing Database on Fly.io (Important!)

```bash
# SSH into Fly.io and create backup
fly ssh console -C "cp /data/freelance_tracker.db /data/freelance_tracker.backup.$(date +%Y%m%d_%H%M%S).db"

# Verify backup was created
fly ssh console -C "ls -lh /data/"
```

### Step 3: Scale Down Application

This prevents database conflicts while uploading:

```bash
fly scale count 0
```

### Step 4: Upload Database Using SFTP

```bash
# Start SFTP session
fly ssh sftp shell

# Inside SFTP session:
cd /data
put ./freelance_tracker.db
ls -lh
exit
```

**Alternative one-liner:**
```bash
fly ssh sftp shell -C "cd /data && put ./freelance_tracker.db"
```

### Step 5: Verify Upload

```bash
# Check file size matches
fly ssh console -C "ls -lh /data/freelance_tracker.db"

# Verify record counts
fly ssh console -C "sqlite3 /data/freelance_tracker.db 'SELECT COUNT(*) FROM client;'"
fly ssh console -C "sqlite3 /data/freelance_tracker.db 'SELECT COUNT(*) FROM project;'"
```

### Step 6: Scale Back Up

```bash
fly scale count 1
```

### Step 7: Test Application

```bash
fly open
```

## Option 2: Use fly ssh console with Base64 Encoding

For smaller databases (<10MB), you can use base64 encoding:

### Step 1: Encode Database Locally

```bash
base64 ./freelance_tracker.db > db.b64
```

### Step 2: Upload and Decode on Fly.io

```bash
# Copy base64 content
cat db.b64 | pbcopy  # macOS
cat db.b64 | xclip -selection clipboard  # Linux

# SSH into Fly.io
fly ssh console

# Scale down first (in another terminal)
# fly scale count 0

# Paste and decode
cat > /tmp/db.b64 << 'EOF'
# Paste the base64 content here
EOF

# Decode to /data
base64 -d /tmp/db.b64 > /data/freelance_tracker.db

# Clean up
rm /tmp/db.b64

# Verify
ls -lh /data/freelance_tracker.db
sqlite3 /data/freelance_tracker.db "SELECT COUNT(*) FROM client;"

# Exit
exit

# Scale back up
fly scale count 1
```

## Option 3: Mount Volume Locally and Copy

If you have a small database and prefer Docker:

### Step 1: Create Local Volume

```bash
docker volume create flyio-data
```

### Step 2: Copy Database to Volume

```bash
docker run --rm -v flyio-data:/data -v $(pwd):/source alpine cp /source/freelance_tracker.db /data/
```

### Step 3: Update fly.toml

```toml
[mounts]
  source = "flyio-data"
  destination = "/data"
```

### Step 4: Deploy

```bash
fly deploy
```

**Note:** This approach doesn't work directly with Fly.io volumes (which are managed separately), so **Option 1 is recommended**.

## Option 4: Bundle Database in Docker Image (One-Time Deploy)

**⚠️ Use only for initial deployment or complete replacement**

This bundles the database in your Docker image and copies it to the volume on startup.

### Step 1: Create Startup Script

Create `scripts/init-db.sh`:

```bash
#!/bin/sh
set -e

# Only copy database if it's a fresh deployment
if [ "$INIT_DB" = "true" ]; then
    echo "Initializing database from bundled file..."
    cp /app/init-db/freelance_tracker.db /data/freelance_tracker.db
    echo "Database initialized successfully"
else
    echo "Using existing database on volume"
fi

# Start the application
exec ./web -addr=:8080 -dsn=/data/freelance_tracker.db
```

Make it executable:
```bash
chmod +x scripts/init-db.sh
```

### Step 2: Update Dockerfile

```dockerfile
# ... existing build stages ...

# Final stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata

# Create directory structure
RUN mkdir -p /app/init-db /data
WORKDIR /app

# Copy binary and assets
COPY --from=builder /build/web .
COPY --from=builder /build/ui ./ui
COPY --from=builder /build/migrations ./migrations

# Copy the local SQLite database
COPY ./freelance_tracker.db /app/init-db/freelance_tracker.db

# Copy startup script
COPY ./scripts/init-db.sh .
RUN chmod +x init-db.sh

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app /data

USER appuser

EXPOSE 8080
ENV ADDR=":8080"
ENV DSN="/data/freelance_tracker.db"

# Use startup script
CMD ["./init-db.sh"]
```

### Step 3: Deploy with Environment Variable

```bash
# Set INIT_DB=true to trigger database initialization
fly secrets set INIT_DB=true

# Deploy
fly deploy

# After successful deployment, unset the variable
fly secrets unset INIT_DB
```

### Step 4: Cleanup Dockerfile

After successful deployment, remove the database bundling from Dockerfile:

```dockerfile
# Remove these lines:
# COPY ./freelance_tracker.db /app/init-db/freelance_tracker.db
# COPY ./scripts/init-db.sh .

# Restore original CMD:
CMD ["./web", "-addr=:8080", "-dsn=/data/freelance_tracker.db"]
```

Then deploy again to clean up:
```bash
fly deploy
```

## Recommended Workflow

Here's the complete recommended workflow:

```bash
# 1. Run your MySQL migration locally
# (using whatever migration tool you have)
# Result: ./freelance_tracker.db

# 2. Verify local database
sqlite3 ./freelance_tracker.db "SELECT COUNT(*) FROM client;"
sqlite3 ./freelance_tracker.db "SELECT COUNT(*) FROM project;"
sqlite3 ./freelance_tracker.db "SELECT COUNT(*) FROM timesheet;"

# 3. Backup existing Fly.io database
fly ssh console -C "cp /data/freelance_tracker.db /data/freelance_tracker.backup.$(date +%Y%m%d).db"

# 4. Scale down to prevent conflicts
fly scale count 0

# 5. Upload new database
fly ssh sftp shell
cd /data
put ./freelance_tracker.db
ls -lh
exit

# 6. Verify upload
fly ssh console -C "sqlite3 /data/freelance_tracker.db 'SELECT COUNT(*) FROM client;'"

# 7. Scale back up
fly scale count 1

# 8. Test application
fly open

# 9. Monitor logs
fly logs

# 10. If everything works, clean up old backup (optional)
fly ssh console -C "rm /data/freelance_tracker.backup.*.db"
```

## Troubleshooting

### Issue: "Permission Denied" When Uploading

**Solution:** Ensure the app is scaled down and you have proper permissions:

```bash
fly scale count 0
fly ssh sftp shell
cd /data
put ./freelance_tracker.db
```

### Issue: Database File Too Large for SFTP

**Solution:** Use compression:

```bash
# Compress locally
gzip -c ./freelance_tracker.db > freelance_tracker.db.gz

# Upload compressed file
fly ssh sftp shell
cd /data
put ./freelance_tracker.db.gz
exit

# Decompress on Fly.io
fly ssh console -C "gunzip -f /data/freelance_tracker.db.gz"
```

### Issue: Application Still Using Old Database

**Solution:** Force restart:

```bash
fly apps restart
```

### Issue: Need to Roll Back

**Solution:** Restore from backup:

```bash
fly scale count 0
fly ssh console
cp /data/freelance_tracker.backup.YYYYMMDD.db /data/freelance_tracker.db
exit
fly scale count 1
```

## Post-Upload Checklist

After uploading the database:

- [ ] Verify record counts match source
- [ ] Test login functionality
- [ ] Check all clients appear
- [ ] Verify projects are linked correctly
- [ ] Test creating new records
- [ ] Check timesheets and invoices
- [ ] Monitor application logs: `fly logs`
- [ ] Keep backup for 30 days

## Security Considerations

### Never Commit Database Files

Add to `.gitignore`:

```gitignore
*.db
*.db-shm
*.db-wal
freelance_tracker*.db
```

### Encrypt Database During Transfer (Optional)

For sensitive data:

```bash
# Encrypt locally
openssl enc -aes-256-cbc -salt -in freelance_tracker.db -out freelance_tracker.db.enc

# Upload encrypted file
fly ssh sftp shell
cd /data
put ./freelance_tracker.db.enc
exit

# Decrypt on Fly.io
fly ssh console
openssl enc -d -aes-256-cbc -in /data/freelance_tracker.db.enc -out /data/freelance_tracker.db
rm /data/freelance_tracker.db.enc
exit
```

## Performance Tips

### Large Databases (>100MB)

For large databases:

1. **Use compression:**
   ```bash
   gzip -9 ./freelance_tracker.db
   ```

2. **Upload during low-traffic hours**

3. **Monitor upload progress:**
   ```bash
   # Use rsync through SSH for progress reporting
   fly ssh console -C "exit" && rsync -avz --progress -e "fly ssh console --command" ./freelance_tracker.db :/data/
   ```

4. **Verify integrity after upload:**
   ```bash
   # Check file size
   fly ssh console -C "ls -lh /data/freelance_tracker.db"

   # Run integrity check
   fly ssh console -C "sqlite3 /data/freelance_tracker.db 'PRAGMA integrity_check;'"
   ```

## Alternative: Direct Volume Access

If you need direct access to Fly.io volumes (advanced):

```bash
# List volumes
fly volumes list

# Create a temporary machine with volume attached
fly machine run --volume <volume-id>:/data alpine sleep 3600

# Use the temporary machine to upload
fly ssh console -a <temp-machine-name>
# Upload your database
exit

# Destroy temporary machine
fly machine destroy <machine-id>
```

## Summary

The **recommended approach** is:

1. ✅ **Run migration locally** - safest, you can test thoroughly
2. ✅ **Upload via SFTP** - simple, reliable, direct to volume
3. ✅ **Scale down during upload** - prevents database conflicts
4. ✅ **Verify before scaling up** - ensure data integrity
5. ✅ **Keep backups** - safety net for rollback

This gives you full control over the migration process and avoids complications with bundling databases in Docker images.
