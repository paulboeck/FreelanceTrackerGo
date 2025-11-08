#!/bin/bash
set -e

# Deploy SQLite Database to Fly.io Volume
# This script uploads a local SQLite database to Fly.io's persistent volume

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
LOCAL_DB="${1:-./freelance_tracker.db}"
REMOTE_PATH="/data/freelance_tracker.db"

echo -e "${GREEN}Fly.io Database Deployment Script${NC}"
echo "=================================="
echo ""

# Check if local database exists
if [ ! -f "$LOCAL_DB" ]; then
    echo -e "${RED}Error: Local database not found: $LOCAL_DB${NC}"
    echo "Usage: $0 [path-to-local-database.db]"
    exit 1
fi

echo "Local database: $LOCAL_DB"
echo "Remote path: $REMOTE_PATH"
echo ""

# Show database size
DB_SIZE=$(ls -lh "$LOCAL_DB" | awk '{print $5}')
echo "Database size: $DB_SIZE"
echo ""

# Verify record counts
echo -e "${YELLOW}Local database record counts:${NC}"
sqlite3 "$LOCAL_DB" "SELECT COUNT(*) as clients FROM client;" 2>/dev/null && echo "  ✓ Clients counted" || echo "  ⚠ Could not count clients"
sqlite3 "$LOCAL_DB" "SELECT COUNT(*) as projects FROM project;" 2>/dev/null && echo "  ✓ Projects counted" || echo "  ⚠ Could not count projects"
sqlite3 "$LOCAL_DB" "SELECT COUNT(*) as timesheets FROM timesheet;" 2>/dev/null && echo "  ✓ Timesheets counted" || echo "  ⚠ Could not count timesheets"
echo ""

# Confirm with user
echo -e "${YELLOW}This will:${NC}"
echo "  1. Create a backup of the existing database on Fly.io"
echo "  2. Scale down your application (to prevent database conflicts)"
echo "  3. Upload $LOCAL_DB to Fly.io"
echo "  4. Verify the upload"
echo "  5. Scale your application back up"
echo ""
read -p "Continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Deployment cancelled."
    exit 0
fi

echo ""
echo -e "${GREEN}Step 1: Creating backup on Fly.io...${NC}"
BACKUP_NAME="freelance_tracker.backup.$(date +%Y%m%d_%H%M%S).db"
fly ssh console -C "cp $REMOTE_PATH /data/$BACKUP_NAME" 2>/dev/null && \
    echo "  ✓ Backup created: $BACKUP_NAME" || \
    echo "  ⚠ Could not create backup (database may not exist yet)"

echo ""
echo -e "${GREEN}Step 2: Scaling down application...${NC}"
fly scale count 0
echo "  ✓ Application scaled to 0 instances"

echo ""
echo -e "${GREEN}Step 3: Uploading database to Fly.io...${NC}"
echo "  Starting upload..."

# Create a temporary SFTP batch file
SFTP_BATCH=$(mktemp)
cat > "$SFTP_BATCH" << EOF
cd /data
put $LOCAL_DB freelance_tracker.db
ls -lh freelance_tracker.db
exit
EOF

# Upload using SFTP
fly ssh sftp shell < "$SFTP_BATCH"
rm "$SFTP_BATCH"

echo "  ✓ Upload complete"

echo ""
echo -e "${GREEN}Step 4: Verifying upload...${NC}"
REMOTE_SIZE=$(fly ssh console -C "ls -lh $REMOTE_PATH | awk '{print \$5}'")
echo "  Remote database size: $REMOTE_SIZE"

# Verify record counts on Fly.io
echo "  Checking record counts on Fly.io..."
fly ssh console -C "sqlite3 $REMOTE_PATH 'SELECT COUNT(*) FROM client;'" >/dev/null 2>&1 && \
    echo "    ✓ Database accessible on Fly.io" || \
    echo "    ⚠ Warning: Could not query database"

echo ""
echo -e "${GREEN}Step 5: Scaling application back up...${NC}"
fly scale count 1
echo "  ✓ Application scaled to 1 instance"

echo ""
echo -e "${GREEN}Deployment complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Test your application: fly open"
echo "  2. Monitor logs: fly logs"
echo "  3. Verify data in the UI"
echo ""
echo "Backup available at: /data/$BACKUP_NAME"
echo ""
echo "To rollback if needed:"
echo "  fly scale count 0"
echo "  fly ssh console -C \"cp /data/$BACKUP_NAME $REMOTE_PATH\""
echo "  fly scale count 1"
