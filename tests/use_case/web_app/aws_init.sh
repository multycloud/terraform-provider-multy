#!/bin/bash -xe

{
date
sudo apt-get update -y

region=$(curl -s http://169.254.169.254/latest/meta-data/placement/region)

sudo apt-get -y install git mysql-client npm nodejs jq awscli curl

sudo chmod a+rwx .
git clone https://github.com/FaztTech/nodejs-mysql-links.git
cd nodejs-mysql-links

#export DATABASE_HOST=$(aws ssm get-parameter --with-decryption --name "/${vault_name}/${db_host_secret_name}" --region "$region" | jq ".Parameter.Value" -r)
#export DATABASE_USER=$(aws ssm get-parameter --with-decryption --name "/${vault_name}/${db_username_secret_name}" --region "$region" | jq ".Parameter.Value" -r)
#export DATABASE_PASSWORD=$(aws ssm get-parameter --with-decryption --name "/${vault_name}/${db_password_secret_name}" --region "$region" | jq ".Parameter.Value" -r)

# both aws and az will try to run this command but only one will succeed
mysql -h ${db_host_secret_name} -P 3306 -u ${db_username_secret_name} --password=${db_password_secret_name} -e 'source database/db.sql' || true

curl https://raw.githubusercontent.com/creationix/nvm/master/install.sh | bash
source ~/.profile
nvm install v10.13.0
npm i
npm run build
date
npm start

} |& tee -a logs.txt