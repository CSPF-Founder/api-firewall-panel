#!/bin/bash

mkdir -p /app/panelbuild/
mkdir -p /app/panel/
cp -a /vagrant/code/panel/. /app/panelbuild/
cd /app/panelbuild/
make build
cp /app/panelbuild/bin/server /app/panel/server
chmod +x /app/panel/server
mkdir -p /app/panel/frontend/external/
sudo cp -p -a /app/panelbuild/frontend/external/. /app/panel/frontend/external/
mkdir -p /app/panel/frontend/external/
mkdir /app/panel/certs/
mkdir -p /app/data/temp_uploads/
mkdir -p /app/data/docker/
mkdir -p /app/data/input/
echo "Generating SSL Certificate..."
openssl req -x509 -sha256 -nodes -days 4096 -newkey rsa:2048 -keyout /app/panel/certs/panel.key -out /app/panel/certs/panel.crt -subj "/C=US/ST=CyberSecurityandPrivacyFoundation/L=CyberSecurityandPrivacyFoundation/O=Dis/CN=apiscanner"
echo "SSL Generated"
sudo chown -R vagrant:vagrant /app/panel/certs/panel.key
sudo chown -R vagrant:vagrant /app/panel/certs/panel.crt

sudo echo '[Unit]
Description=API Protector Panel

[Service]
User=vagrant
Group=vagrant
WorkingDirectory=/app/panel/
ExecStart=/app/panel/server
Restart=always

[Install]
WantedBy=multi-user.target' > /etc/systemd/system/panel.service

sudo systemctl daemon-reload
sudo systemctl enable panel.service
sudo systemctl start panel.service
sudo systemctl status panel.service
