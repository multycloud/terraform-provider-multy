#!/bin/bash -xe

{
date
sudo apt-get update -y

# az cli
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
az login --identity --allow-no-subscriptions

sudo apt-get -y install git npm mysql-client curl jq
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
sudo chmod a+rwx .
git clone https://github.com/FaztTech/nodejs-mysql-links.git
cd nodejs-mysql-links

#export DATABASE_HOST=$(az keyvault secret show --vault-name '${vault_name}' -n '${db_host_secret_name}' | jq ".value" -r)
#export DATABASE_USER=$(az keyvault secret show --vault-name '${vault_name}' -n '${db_username_secret_name}' | jq ".value" -r)
#export DATABASE_PASSWORD=$(az keyvault secret show --vault-name '${vault_name}' -n '${db_password_secret_name}' | jq ".value" -r)

# both aws and az will try to run this command but only one will succeed
mysql -h ${db_host_secret_name} -P 3306 -u ${db_username_secret_name} --password=${db_password_secret_name} -e 'source database/db.sql' || true

curl https://raw.githubusercontent.com/creationix/nvm/master/install.sh | bash
source ~/.profile
npm i
npm run build
date
npm start

} |& tee -a logs.txt