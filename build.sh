#!/bin/bash
set -euo pipefail

PLUGIN_ID="oci-metrics-datasource"
VERSION=$(node -p "require('./package.json').version")
ARCHIVE="${PLUGIN_ID}-${VERSION}.zip"

if [[ ! -d ./node_modules ]]; then
  echo "dependencies not installed try running: yarn"
  exit 1
fi

rm -rf ./dist "./${PLUGIN_ID}" "./${ARCHIVE}"

yarn run build
mage -v

cp LICENSE.txt ./dist/LICENSE

for binary in \
  oci-metrics-plugin_darwin_amd64 \
  oci-metrics-plugin_darwin_arm64 \
  oci-metrics-plugin_linux_amd64 \
  oci-metrics-plugin_linux_arm \
  oci-metrics-plugin_linux_arm64 \
  oci-metrics-plugin_windows_amd64.exe; do
  if [[ ! -x "./dist/${binary}" ]]; then
    echo "missing executable backend binary: ${binary}"
    exit 1
  fi
done

if [[ ! -f ./dist/go_plugin_build_manifest ]]; then
  echo "missing Go plugin build manifest"
  exit 1
fi

if [[ -z "${1:-}" ]]; then
  echo "sign argument not specified, continuing without sign the plugin"
elif [[ "$1" == "sign" ]]; then
  yarn sign
else
  echo "Usage: ./build.sh [sign]"
  exit 1
fi

cp -R ./dist "./${PLUGIN_ID}"
zip -qr "./${ARCHIVE}" "./${PLUGIN_ID}"

echo "created ${ARCHIVE}"
