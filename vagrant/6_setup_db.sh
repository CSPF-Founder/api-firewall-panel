#!/bin/bash
mkdir -p /app/dbinfra/
cp -a /vagrant/code/dbdocker/. /app/dbinfra/
#Random DB Password
rootpass=`(sudo head /dev/urandom | tr -dc 'A-Za-z0-9' | head -c 20)`
normalpass=`(sudo head /dev/urandom | tr -dc 'A-Za-z0-9' | head -c 20)`
sed -i "s/\[ROOT_PASS_TO_REPLACE\]/$rootpass/g" /app/dbinfra/docker-compose.yml
sed -i "s/\[PASSWORD_TO_REPLACE\]/$normalpass/g" /app/dbinfra/docker-compose.yml
cd /app/dbinfra/
make up
echo "Writing .env file"
echo "PRODUCT_TITLE=API Protector">>/app/panel/.env
echo "SERVER_ADDRESS=0.0.0.0:8443">>/app/panel/.env
echo "DATABASE_URI=root:$rootpass@(127.0.0.1:3306)/api_protector?charset=utf8&parseTime=True&loc=Local">>/app/panel/.env
echo "DBMS_TYPE=mysql">>/app/panel/.env
echo "COPYRIGHT_FOOTER_COMPANY= Cyber Security and Privacy Foundation ">>/app/panel/.env
echo "CLI_BIN_PATH=/app/bin/apcli">>/app/panel/.env
echo "WORK_DIR=/app/data/">>/app/panel/.env
echo "# The temp_uploads should be in the same device as the WORK_DIR(or docker gives invalid-cross)">>/app/panel/.env
echo "TEMP_UPLOADS_DIR=/app/data/">>/app/panel/.env
echo "MIGRATIONS_PREFIX=db">>/app/panel/.env
echo "LOG_LEVEL=info">>/app/panel/.env
echo "# TLS Configuration">> /app/panel/.env
echo "USE_TLS=true">>/app/panel/.env
echo "CERT_PATH=/app/panel/certs/panel.crt">>/app/panel/.env
echo "KEY_PATH=/app/panel/certs/panel.key">>/app/panel/.env