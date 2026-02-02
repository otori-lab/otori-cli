import os

# --- CONFIGURATION ---
# On calcule le chemin vers 'fs_template' (relatif à ce script)
# Structure: classic/docker/generer_serveur.py  ->  classic/cowrie/fs_template
script_dir = os.path.dirname(os.path.abspath(__file__))
root_fs = os.path.join(script_dir, "..", "cowrie", "fs_template")
root_fs = os.path.abspath(root_fs)

print(f"🚀 Démarrage de la génération dans : {root_fs}")

if not os.path.exists(root_fs):
    os.makedirs(root_fs)
    print(f"⚠️ Dossier créé : {root_fs}")

# Fonction pour créer un fichier format Linux
def create_file(path_rel, content):
    # Chemin complet
    full_path = os.path.join(root_fs, path_rel)
    # Créer le dossier parent si besoin
    os.makedirs(os.path.dirname(full_path), exist_ok=True)
    
    # Écriture en mode texte avec encodage UTF-8
    # newline='\n' force le format Linux (LF) même sur Windows
    with open(full_path, "w", encoding="utf-8", newline='\n') as f:
        f.write(content.strip())  # strip() enlève les espaces superflus au début/fin
    
    print(f"  [+] {path_rel}")

# ==============================================================================
# 1. IDENTITÉ DU SERVEUR & RÉSEAU
# ==============================================================================

create_file("etc/hostname", "srv-prod-01")

create_file("etc/hosts", """
127.0.0.1       localhost
127.0.1.1       srv-prod-01.localdomain srv-prod-01
192.168.1.10    db-internal.local
""")

create_file("etc/resolv.conf", """
nameserver 8.8.8.8
nameserver 1.1.1.1
""")

create_file("etc/network/interfaces", """
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
""")

# ==============================================================================
# 2. UTILISATEURS & MOTS DE PASSE (LEURRES)
# ==============================================================================

create_file("etc/shadow", """
root:$6$hJ8s...SALT...$eX5.:19000:0:99999:7:::
admin:$6$Fk92...SALT...$Up9.:19000:0:99999:7:::
www-data:*:19000:0:99999:7:::
""")

# ==============================================================================
# 3. HISTORIQUE DE COMMANDES
# ==============================================================================

create_file("home/admin/.bash_history", """
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
""")

# ==============================================================================
# 4. WEB & CONFIGURATION
# ==============================================================================

create_file("var/www/html/index.php", """
<?php
// Silence is golden.
require( dirname( __FILE__ ) . '/wp-blog-header.php' );
""")

create_file("var/www/html/wp-config.php", """
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
""")

create_file("var/www/html/.env", """
APP_ENV=production
APP_DEBUG=false
APP_KEY=base64:U29tZVJhbmRvbUtleUdlbmVyYXRlZA==
DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=laravel_app
DB_USERNAME=admin
DB_PASSWORD=RootPassword123!
""")

# ==============================================================================
# 5. SSH & LOGS
# ==============================================================================

create_file("home/admin/.ssh/authorized_keys", "ssh-rsa AAAAB3NzaC1yc2E...dev-laptop@company.com")
create_file("home/admin/.ssh/known_hosts", "|1|F1E4... ssh-rsa AAAAB...")

create_file("home/admin/.ssh/id_rsa", """
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXk...
(CONTENU SIMULÉ - NE FONCTIONNE PAS VRAIMENT)
...X5F2a
-----END OPENSSH PRIVATE KEY-----
""")

create_file("var/log/auth.log", """
Jan 20 08:00:01 srv-prod-01 CRON[1122]: pam_unix(cron:session): session opened for user root
Jan 20 08:15:22 srv-prod-01 sshd[2400]: Accepted password for admin from 10.0.0.55 port 54322 ssh2
Jan 20 12:30:45 srv-prod-01 sudo:    admin : TTY=pts/0 ; PWD=/home/admin ; USER=root ; COMMAND=/usr/bin/apt update
""")

create_file("var/log/syslog", """
Jan 21 06:25:01 srv-prod-01 systemd[1]: Starting Daily apt upgrade and clean activities...
Jan 21 06:25:05 srv-prod-01 systemd[1]: apt-daily-upgrade.service: Succeeded.
Jan 21 06:25:05 srv-prod-01 systemd[1]: Finished Daily apt upgrade and clean activities.
""")

create_file("var/log/nginx/access.log", """
10.0.0.55 - - [21/Jan/2026:10:00:00 +0000] "GET / HTTP/1.1" 200 612 "-" "Mozilla/5.0"
66.249.66.1 - - [21/Jan/2026:10:05:00 +0000] "GET /robots.txt HTTP/1.1" 200 120 "-" "Googlebot/2.1"
""")

# ==============================================================================
# 6. SCRIPTS & NOTES
# ==============================================================================

create_file("opt/backup/backup_db.sh", """
#!/bin/bash
# Backup de la base de données
mysqldump -u root -p'RootPassword123!' --all-databases > /backup/db_full.sql
echo 'Backup terminé'
""")

create_file("home/admin/docker-compose.yml", """
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
""")

create_file("home/admin/notes.txt", """
TODO:
- Renouveler le certificat SSL avant le 25/02
- Vérifier pourquoi le backup a échoué hier
- Mettre à jour la clé SSH de Fabrice
""")

print("\n✅ Terminé ! Lancez maintenant votre commande de fusion (FSCTL) !")