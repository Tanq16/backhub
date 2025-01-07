#!/bin/sh

run_backup() {
    echo "Starting backup at $(date)"
    backhub
    echo "Backup completed at $(date)"
}

run_backup

# Schedule backup every 3 days
while true; do
    sleep 259200  # 3 days in seconds
    run_backup
done
