#!/bin/sh

run_backup() {
    echo "Starting backup at $(date)"
    backhub /app/config.yaml
    echo "Backup completed at $(date)"
}

run_backup

# Schedule backup every 3 days
while true; do
    sleep 259200
    run_backup
done
