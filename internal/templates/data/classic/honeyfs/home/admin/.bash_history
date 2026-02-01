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