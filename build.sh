#!/bin/bash

#Do grunt work
# nvm install 12.20
# nvm use 12.20

# rm -rf node_modules

# yarn 


if [[ ! -d ./node_modules ]]; then
  echo "dependencies not installed try running: yarn"
  exit 1
fi
rm -rf ./oci-metrics-datasource
rm ./oci-metrics-datasource.zip 
rm ./plugin.tar
./node_modules/.bin/grunt

# build go

POST=''
GOOS=''

OS="`uname`"
case $OS in
  'Linux')
      POST='_linux_amd64'
      GOOS="linux"
    ;;
  'Darwin')
      POST='_darwin_amd64'
      GOOS="darwin"
    ;;
  'AIX') ;;
  *) ;;
esac

go mod vendor

if [[ ! -d ./vendor ]]; then
  echo "dependencies not installed try running | go mod vendor didn't work as expected"
  exit 1
fi

echo "building go binary"
GOOS=$GOOS go build -o ./dist/oci-metrics-plugin$POST

# For debugger
#  GOOS=$GOOS go build -o ./dist/oci-metrics-plugin$POST -gcflags="all=-N -l"

# For release
 GOOS=linux go build -o ./dist/oci-metrics-plugin_linux_amd64
 GOOS=windows GOARCH=amd64 go build -o ./dist/oci-metrics-plugin_windows_amd64.exe

grafana-toolkit plugin:sign

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

# Please make sure if you have the api keys installed in bash profile in name,  GRAFANA_API_KEY
# Note : Please make sure that you are running the commands in a non-proxy env and without vpn, else grafana signing might fail"
# yarn  global add @grafana/toolkit
# grafana-toolkit plugin:sign

