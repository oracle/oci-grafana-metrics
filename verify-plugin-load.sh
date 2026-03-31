#!/usr/bin/env bash
#
# Verifies the oci-metrics-datasource plugin loaded in each Grafana instance.
#
# Usage:
#   docker compose up -d
#   ./verify-plugin-load.sh
#
# Expected: v7.5 fails (grafanaDependency >= 9.0.0), v9-v12 pass.

set -euo pipefail

PLUGIN_ID="oci-metrics-datasource"
MAX_WAIT=60
POLL_INTERVAL=3

VERSIONS=("v7.5" "v9" "v10" "v11" "v12")
PORTS=(3075 3090 3100 3110 3120)

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BOLD='\033[1m'
NC='\033[0m'

pass_count=0
fail_count=0

echo ""
echo -e "${BOLD}OCI Metrics Plugin Load Verification${NC}"
echo "======================================"
echo ""

for i in "${!VERSIONS[@]}"; do
    version="${VERSIONS[$i]}"
    port="${PORTS[$i]}"
    url="http://localhost:${port}"

    printf "%-12s (port %s) ... " "Grafana ${version}" "${port}"

    # Wait for Grafana to be ready
    elapsed=0
    ready=false
    while [ "$elapsed" -lt "$MAX_WAIT" ]; do
        code=$(curl -s -o /dev/null -w "%{http_code}" "${url}/api/health" 2>/dev/null || echo "000")
        if [ "$code" = "200" ]; then
            ready=true
            break
        fi
        sleep "$POLL_INTERVAL"
        elapsed=$((elapsed + POLL_INTERVAL))
    done

    if [ "$ready" = false ]; then
        echo -e "${RED}FAIL${NC} (not ready after ${MAX_WAIT}s)"
        fail_count=$((fail_count + 1))
        continue
    fi

    # Check if plugin is listed
    if command -v jq &>/dev/null; then
        plugin=$(curl -s "${url}/api/plugins" 2>/dev/null | jq -r ".[] | select(.id == \"${PLUGIN_ID}\") | .id" 2>/dev/null)
    else
        plugin=$(curl -s "${url}/api/plugins" 2>/dev/null | grep -o "\"id\":\"${PLUGIN_ID}\"" 2>/dev/null || true)
    fi

    if [ -n "$plugin" ]; then
        echo -e "${GREEN}PASS${NC} (plugin loaded)"
        pass_count=$((pass_count + 1))
    else
        echo -e "${RED}FAIL${NC} (plugin not found)"
        fail_count=$((fail_count + 1))
    fi
done

echo ""
echo "--------------------------------------"
echo -e "Results: ${GREEN}${pass_count} passed${NC}, ${RED}${fail_count} failed${NC}"
echo ""

if [ "$fail_count" -gt 1 ]; then
    echo -e "${YELLOW}NOTE:${NC} v7.5 failure is expected (grafanaDependency >= 9.0.0)."
    echo "      Check logs for unexpected failures: docker compose logs <service>"
fi

# v7.5 failure is expected, so exit 0 if only 1 failure
[ "$fail_count" -le 1 ]
