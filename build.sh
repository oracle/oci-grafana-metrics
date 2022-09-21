#!/bin/bash

#Do grunt work
# nvm install 12.20
# nvm use 12.20

if [[ ! -d ./node_modules ]]; then
  echo "dependencies not installed try running: yarn"
  exit 1
fi
rm -rf ./oci-metrics-datasource
rm ./oci-metrics-datasource.zip 
rm ./plugin.tar
./node_modules/.bin/grunt

mage --debug -v

#grafana-toolkit plugin:sign

mv ./dist ./oci-metrics-datasource
tar cvf plugin.tar ./oci-metrics-datasource
zip -r oci-metrics-datasource ./oci-metrics-datasource

# Instructions for signing
# Please make sure
# nvm install 12.20

# nvm use 12.20

# yarn
# For grafana publishing
# yarn install --pure-lockfile && yarn build
#
# Please make sure if you have the api keys installed in bash profile in name,  GRAFANA_API_KEY
# Note : Please make sure that you are running the commands in a non-proxy env and without vpn, else grafana signing might fail"
# yarn  global add @grafana/toolkit
# grafana-toolkit plugin:sign

