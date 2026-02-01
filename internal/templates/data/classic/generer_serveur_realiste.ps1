# ==============================================================================
# GÉNÉRATEUR DE HONEYPOT RÉALISTE (Mode Serveur Prod)
# ==============================================================================

# On recupere le dossier ou se trouve CE script (le dossier 'docker')
$ScriptDir = Split-Path $MyInvocation.MyCommand.Path

# On remonte d'un cran pour aller dans 'cowrie/fs_template'
# Cela calcule le chemin ABSOLU (ex: C:\Users\JN\...\cowrie\fs_template)
$RootFS = Join-Path (Split-Path $ScriptDir) "cowrie\fs_template"

Write-Host "🚀 Démarrage de la génération dans : $RootFS" -ForegroundColor Cyan

# Si le dossier fs_template n'existe pas, on le cree pour eviter le crash
if (-not (Test-Path $RootFS)) {
    New-Item -ItemType Directory -Force -Path $RootFS | Out-Null
    Write-Host "⚠️ Dossier créé : $RootFS" -ForegroundColor Yellow
}

# Fonction pour créer un fichier avec contenu
function New-LinuxFile ($Path, $Content) {
    # On combine pour avoir le chemin complet
    $FullPath = Join-Path $RootFS $Path
    $Dir = Split-Path $FullPath
    
    # Créer le dossier s'il n'existe pas
    if (-not (Test-Path $Dir)) { New-Item -ItemType Directory -Force -Path $Dir | Out-Null }
    
    # Conversion LF et écriture
    $ContentLF = $Content -replace "`r`n", "`n"
    [System.IO.File]::WriteAllText($FullPath, $ContentLF, [System.Text.Encoding]::UTF8)
    Write-Host "  [+] $Path" -ForegroundColor Green
}

# ... LE RESTE DU SCRIPT NE CHANGE PAS ...

# ==============================================================================
# 1. IDENTITÉ DU SERVEUR & RÉSEAU
# ==============================================================================

New-LinuxFile "etc/hostname" "srv-prod-01"

New-LinuxFile "etc/hosts" @"
127.0.0.1       localhost
127.0.1.1       srv-prod-01.localdomain srv-prod-01
192.168.1.10    db-internal.local
"@

New-LinuxFile "etc/resolv.conf" @"
nameserver 8.8.8.8
nameserver 1.1.1.1
"@

New-LinuxFile "etc/network/interfaces" @"
# The loopback network interface
auto lo
iface lo inet loopback

# The primary network interface
auto eth0
iface eth0 inet static
    address 192.168.1.50
    netmask 255.255.255.0
    gateway 192.168.1.1
    dns-nameservers 8.8.8.8
"@

# ==============================================================================
# 2. UTILISATEURS & MOTS DE PASSE (LEURRES)
# ==============================================================================

# Un faux /etc/shadow (ne donne pas vraiment accès, c'est pour la déco si on fait 'cat')
New-LinuxFile "etc/shadow" @"
root:$6$hJ8s...SALT...$eX5.:19000:0:99999:7:::
admin:$6$Fk92...SALT...$Up9.:19000:0:99999:7:::
www-data:*:19000:0:99999:7:::
"@

# ==============================================================================
# 3. HISTORIQUE DE COMMANDES (LA PARTIE LA PLUS IMPORTANTE)
# ==============================================================================

$BashHistory = @"
ping google.com
sudo apt update
sudo apt install nginx mysql-server git -y
sudo systemctl enable nginx
cd /var/www/html
ls -la
git clone https://github.com/wordpress/wordpress.git .
mv wp-config-sample.php wp-config.php
nano wp-config.php
mysql -u root -p
sudo service mysql restart
cat /var/log/nginx/error.log
htop
docker ps
docker-compose up -d
ssh-keygen -t rsa -b 4096
cat ~/.ssh/id_rsa.pub
cd /opt/backup
./backup.sh
whoami
pwd
exit
"@
New-LinuxFile "home/admin/.bash_history" $BashHistory

# ==============================================================================
# 4. FICHIERS WEB & CONFIGURATION (AVEC MOTS DE PASSE)
# ==============================================================================

New-LinuxFile "var/www/html/index.php" @"
<?php
// Silence is golden.
require( dirname( __FILE__ ) . '/wp-blog-header.php' );
"@

# LE FICHIER QUE LES ATTAQUANTS CHERCHENT : Config avec mot de passe DB
New-LinuxFile "var/www/html/wp-config.php" @"
<?php
/**
 * The base configuration for WordPress
 */
define( 'DB_NAME', 'wp_prod_db' );
define( 'DB_USER', 'wp_user' );
define( 'DB_PASSWORD', 'SuperSecurePass2024!' ); // <--- LEURRE ICI
define( 'DB_HOST', 'localhost' );
define( 'DB_CHARSET', 'utf8' );
define( 'DB_COLLATE', '' );

/** Authentication Unique Keys and Salts. */
define( 'AUTH_KEY',         'put your unique phrase here' );
define( 'SECURE_AUTH_KEY',  'put your unique phrase here' );
"@

New-LinuxFile "var/www/html/.env" @"
APP_ENV=production
APP_DEBUG=false
APP_KEY=base64:U29tZVJhbmRvbUtleUdlbmVyYXRlZA==
DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=laravel_app
DB_USERNAME=admin
DB_PASSWORD=RootPassword123!
"@

# ==============================================================================
# 5. SSH & CLÉS (RÉALISME MAXIMAL)
# ==============================================================================

New-LinuxFile "home/admin/.ssh/authorized_keys" "ssh-rsa AAAAB3NzaC1yc2E...dev-laptop@company.com"
New-LinuxFile "home/admin/.ssh/known_hosts" "|1|F1E4... ssh-rsa AAAAB..."

# Fausse clé privée (tronquée pour l'exemple)
New-LinuxFile "home/admin/.ssh/id_rsa" @"
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXk...
(CONTENU SIMULÉ - NE FONCTIONNE PAS VRAIMENT)
...X5F2a
-----END OPENSSH PRIVATE KEY-----
"@

# ==============================================================================
# 6. LOGS SYSTÈME (ACTIVITÉ SIMULÉE)
# ==============================================================================

New-LinuxFile "var/log/auth.log" @"
Jan 20 08:00:01 srv-prod-01 CRON[1122]: pam_unix(cron:session): session opened for user root
Jan 20 08:15:22 srv-prod-01 sshd[2400]: Accepted password for admin from 10.0.0.55 port 54322 ssh2
Jan 20 12:30:45 srv-prod-01 sudo:    admin : TTY=pts/0 ; PWD=/home/admin ; USER=root ; COMMAND=/usr/bin/apt update
"@

New-LinuxFile "var/log/syslog" @"
Jan 21 06:25:01 srv-prod-01 systemd[1]: Starting Daily apt upgrade and clean activities...
Jan 21 06:25:05 srv-prod-01 systemd[1]: apt-daily-upgrade.service: Succeeded.
Jan 21 06:25:05 srv-prod-01 systemd[1]: Finished Daily apt upgrade and clean activities.
"@

New-LinuxFile "var/log/nginx/access.log" @"
10.0.0.55 - - [21/Jan/2026:10:00:00 +0000] "GET / HTTP/1.1" 200 612 "-" "Mozilla/5.0"
66.249.66.1 - - [21/Jan/2026:10:05:00 +0000] "GET /robots.txt HTTP/1.1" 200 120 "-" "Googlebot/2.1"
"@

# ==============================================================================
# 7. FICHIERS DIVERS (BACKUPS, SCRIPTS, DOCKER)
# ==============================================================================

New-LinuxFile "opt/backup/backup_db.sh" @"
#!/bin/bash
# Backup de la base de données
mysqldump -u root -p'RootPassword123!' --all-databases > /backup/db_full.sql
echo 'Backup terminé'
"@

New-LinuxFile "home/admin/docker-compose.yml" @"
version: '3'
services:
  web:
    image: nginx:latest
    ports:
      - '8080:80'
  db:
    image: mysql:5.7
    environment:
      MYSQL_ROOT_PASSWORD: 'RootPassword123!'
"@

New-LinuxFile "home/admin/notes.txt" @"
TODO:
- Renouveler le certificat SSL avant le 25/02
- Vérifier pourquoi le backup a échoué hier
- Mettre à jour la clé SSH de Fabrice
"@

Write-Host "`n✅ Terminé ! Tous les faux fichiers ont été générés dans $RootFS" -ForegroundColor Green
Write-Host "👉 N'oubliez pas de lancer votre commande de mise à jour FSCTL pour que Cowrie les voie !" -ForegroundColor Yellow