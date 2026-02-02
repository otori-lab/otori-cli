#!/bin/bash
# Backup de la base de données
mysqldump -u root -p'RootPassword123!' --all-databases > /backup/db_full.sql
echo 'Backup terminé'