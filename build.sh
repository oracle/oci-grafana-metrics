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
yarn create @grafana/plugin

mage --debug -v

cp LICENSE.txt ./dist/LICENSE

if [ -z $1 ]; then
  echo "sign argument not specified, continuing without sign the plugin"
else
  if [ $1 = "sign" ]; then
    npx @grafana/sign-plugin
  else
    echo "Usage: ./build.sh <sign>"
  fi  
fi

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

