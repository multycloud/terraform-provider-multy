#!/bin/bash -xe

{
date
sudo apt-get update -y

region=$(curl -s http://169.254.169.254/latest/meta-data/placement/region)

sudo apt-get -y install git mysql-client npm nodejs jq awscli curl

sudo chmod a+rwx .
git clone https://github.com/FaztTech/nodejs-mysql-links.git
cd nodejs-mysql-links
sudo git config --global --add safe.directory /nodejs-mysql-links
sudo git reset --hard d084e27cad8cdfb60167e3d7891fed7eead00a76

export DATABASE_HOST=${db_host_secret_name}
export DATABASE_USER=${db_username_secret_name}
export DATABASE_PASSWORD=${db_password_secret_name}

# both aws and az will try to run this command but only one will succeed
mysql -h ${db_host_secret_name} -P 3306 -u ${db_username_secret_name} --password=${db_password_secret_name} -e 'source database/db.sql' || true

curl https://raw.githubusercontent.com/creationix/nvm/master/install.sh | bash
source ~/.profile
source /.nvm/nvm.sh
nvm uninstall v18.0.0
nvm install 16.15.1
nvm use 16.15.1
npm i
npm run build
date
npm start

} |& tee -a logs.txt