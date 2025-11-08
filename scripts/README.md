# Deployment Scripts

## deploy-db-to-fly.sh

Automated script to deploy a local SQLite database to Fly.io's persistent volume.

### Usage

```bash
# Deploy default database (./freelance_tracker.db)
./scripts/deploy-db-to-fly.sh

# Deploy specific database file
./scripts/deploy-db-to-fly.sh /path/to/my-database.db
```

### What It Does

1. ✅ Validates local database file exists
2. ✅ Shows database size and record counts
3. ✅ Creates backup on Fly.io
4. ✅ Scales down application (prevents conflicts)
5. ✅ Uploads database via SFTP
6. ✅ Verifies upload succeeded
7. ✅ Scales application back up
8. ✅ Provides rollback instructions

### Complete Workflow Example

```bash
# 1. Run your MySQL migration locally (using your migration tool)
# Result: ./freelance_tracker.db

# 2. Verify local database
sqlite3 ./freelance_tracker.db "SELECT COUNT(*) FROM client;"
sqlite3 ./freelance_tracker.db "SELECT COUNT(*) FROM project;"

# 3. Deploy to Fly.io
./scripts/deploy-db-to-fly.sh

# 4. Test application
fly open

# 5. Monitor logs
fly logs
```

### Prerequisites

- `flyctl` CLI installed and authenticated
- `sqlite3` command-line tool (for verification)
- Fly.io app deployed with persistent volume
- Local SQLite database file ready to upload

### Safety Features

- **Automatic backup** - Creates timestamped backup before upload
- **User confirmation** - Asks for confirmation before proceeding
- **Graceful scaling** - Scales down to prevent database conflicts
- **Verification** - Checks upload succeeded
- **Rollback instructions** - Provides commands to restore backup

### Rollback

If something goes wrong, restore the backup:

```bash
fly scale count 0
fly ssh console -C "cp /data/freelance_tracker.backup.YYYYMMDD_HHMMSS.db /data/freelance_tracker.db"
fly scale count 1
```

The script prints the exact rollback command when it completes.

## Manual Upload (Alternative)

If you prefer manual control:

```bash
# 1. Create backup
fly ssh console -C "cp /data/freelance_tracker.db /data/freelance_tracker.backup.$(date +%Y%m%d).db"

# 2. Scale down
fly scale count 0

# 3. Upload
fly ssh sftp shell
cd /data
put ./freelance_tracker.db
exit

# 4. Verify
fly ssh console -C "ls -lh /data/freelance_tracker.db"

# 5. Scale up
fly scale count 1
```

See [../.github/DEPLOY_SQLITE_TO_FLY.md](../.github/DEPLOY_SQLITE_TO_FLY.md) for detailed documentation.
