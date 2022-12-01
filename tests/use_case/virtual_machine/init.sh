#!/bin/bash -xe

{
date
sudo apt-get update -y

sudo apt install apache2 -y

cd /var/www/html
sudo echo "Hello from ${cloud}!" | sudo tee index.html

} |& tee -a logs.txt