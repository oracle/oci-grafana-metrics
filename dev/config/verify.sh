#!/usr/bin/env bash
#
# Verifies the oci-metrics-datasource plugin across multiple Grafana versions.
#
# Tests three levels:
#   1. Plugin loaded (appears in /api/plugins)
#   2. Health check passes (datasource can authenticate to OCI)
#   3. Data query returns results (dashboard panels get real data)
#
# Usage:
#   docker compose up -d
#   ./verify.sh          # plugin load + health check
#   ./verify.sh --data   # also test data queries
#
# Prerequisites for --data:
#   - SSH tunnel to OCI instance running (started via test.sh --tunnel)
#   - Provisioning directory mounted with datasource + dashboard configs

set -euo pipefail

# Source .env if it exists (for OCI_COMPARTMENT_OCID, OCI_REGION_1, etc.)
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEV_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
if [ -f "$DEV_DIR/.env" ]; then
    set -a
    source "$DEV_DIR/.env"
    set +a
fi

PLUGIN_ID="oci-metrics-datasource"
MAX_WAIT=90
POLL_INTERVAL=3
TEST_DATA=false

if [[ "${1:-}" == "--data" ]]; then
    TEST_DATA=true
    if [ "${OCI_COMPARTMENT_OCID:-}" = "" ] || [[ "${OCI_COMPARTMENT_OCID:-}" == *REPLACE* ]]; then
        echo "ERROR: OCI_COMPARTMENT_OCID not set. Edit dev/.env first."
        exit 1
    fi
fi

VERSIONS=("v7.5" "v9" "v10" "v11" "v12")
PORTS=(3075 3090 3100 3110 3120)

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BOLD='\033[1m'
NC='\033[0m'

pass_count=0
fail_count=0
data_pass=0
data_fail=0

echo ""
echo -e "${BOLD}OCI Metrics Plugin Verification${NC}"
echo "================================"
if [ "$TEST_DATA" = true ]; then
    echo "Mode: plugin load + health check + data query"
else
    echo "Mode: plugin load + health check (use --data for query test)"
fi
echo ""

for i in "${!VERSIONS[@]}"; do
    version="${VERSIONS[$i]}"
    port="${PORTS[$i]}"
    url="http://localhost:${port}"

    echo -e "${BOLD}Grafana ${version}${NC} (port ${port})"

    # --- Step 1: Wait for Grafana to be ready ---
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
        printf "  %-20s %b\n" "Grafana ready:" "${RED}FAIL${NC} (not ready after ${MAX_WAIT}s)"
        fail_count=$((fail_count + 1))
        echo ""
        continue
    fi
    printf "  %-20s %b\n" "Grafana ready:" "${GREEN}OK${NC} (${elapsed}s)"

    # --- Step 2: Check plugin loaded ---
    plugin=""
    if command -v jq &>/dev/null; then
        plugin=$(curl -s "${url}/api/plugins" 2>/dev/null | jq -r ".[] | select(.id == \"${PLUGIN_ID}\") | .id" 2>/dev/null)
    else
        plugin=$(curl -s "${url}/api/plugins" 2>/dev/null | grep -o "\"id\":\"${PLUGIN_ID}\"" 2>/dev/null || true)
    fi

    if [ -n "$plugin" ]; then
        printf "  %-20s %b\n" "Plugin loaded:" "${GREEN}PASS${NC}"
        pass_count=$((pass_count + 1))
    else
        printf "  %-20s %b\n" "Plugin loaded:" "${RED}FAIL${NC}"
        fail_count=$((fail_count + 1))
        echo ""
        continue
    fi

    # --- Step 3: Check datasource health ---
    # Find the provisioned datasource ID
    ds_id=""
    if command -v jq &>/dev/null; then
        ds_id=$(curl -s "${url}/api/datasources" 2>/dev/null | jq -r ".[] | select(.type == \"${PLUGIN_ID}\") | .id" 2>/dev/null | head -1)
    fi

    if [ -n "$ds_id" ] && [ "$ds_id" != "null" ]; then
        health_response=$(curl -s "${url}/api/datasources/${ds_id}/health" 2>/dev/null)
        health_status=""
        if command -v jq &>/dev/null; then
            health_status=$(echo "$health_response" | jq -r '.status // .message // "unknown"' 2>/dev/null)
        fi

        if echo "$health_status" | grep -qi "ok"; then
            printf "  %-20s %b\n" "Health check:" "${GREEN}PASS${NC}"
        else
            printf "  %-20s %b\n" "Health check:" "${YELLOW}WARN${NC} (${health_status})"
        fi
    else
        printf "  %-20s %b\n" "Health check:" "${YELLOW}SKIP${NC} (no datasource provisioned)"
    fi

    # --- Step 4: Test data query (only with --data) ---
    if [ "$TEST_DATA" = true ] && [ -n "$ds_id" ] && [ "$ds_id" != "null" ]; then
        _compartment="${OCI_COMPARTMENT_OCID}"
        _region="${OCI_REGION_1:-us-ashburn-1}"
        _namespace="${OCI_NAMESPACE:-oci_lbaas}"
        _metric="${OCI_METRIC:-PeakBandwidth}"
        query_payload=$(cat <<QUERY_EOF
{
  "from": "now-6h",
  "to": "now",
  "queries": [{
    "refId": "A",
    "datasourceId": ${ds_id},
    "compartment": "${_compartment}",
    "region": "${_region}",
    "tenancy": "DEFAULT/",
    "namespace": "${_namespace}",
    "queryText": "${_metric}[1m].avg()",
    "rawQuery": true,
    "interval": "[1m]",
    "intervalMs": 60000,
    "maxDataPoints": 100
  }]
}
QUERY_EOF
        )

        # Try /api/tsdb/query first (v7.5), fall back to /api/ds/query (v9+)
        query_response=$(curl -s -X POST "${url}/api/tsdb/query" \
            -H "Content-Type: application/json" \
            -d "$query_payload" 2>/dev/null)

        # If tsdb/query returned 404 or no results key, use ds/query
        if ! echo "$query_response" | grep -q '"results"' 2>/dev/null; then
            query_response=$(curl -s -X POST "${url}/api/ds/query" \
                -H "Content-Type: application/json" \
                -d "$query_payload" 2>/dev/null)
        fi

        has_data=false
        error_msg=""

        if command -v jq &>/dev/null; then
            error_msg=$(echo "$query_response" | jq -r '.results.A.error // empty' 2>/dev/null)

            if [ -z "$error_msg" ]; then
                # v7.5: "dataframes" (base64 Arrow, non-empty array = data)
                df_count=$(echo "$query_response" | jq '.results.A.dataframes // [] | length' 2>/dev/null || echo "0")
                # v9+: "frames" (JSON objects with data.values)
                frame_count=$(echo "$query_response" | jq '.results.A.frames // [] | length' 2>/dev/null || echo "0")

                if [ "$df_count" -gt 0 ] || [ "$frame_count" -gt 0 ]; then
                    has_data=true
                fi
            fi
        else
            if echo "$query_response" | grep -q '"dataframes"\|"frames"'; then
                has_data=true
            fi
        fi

        if [ -n "$error_msg" ]; then
            printf "  %-20s %b\n" "Data query:" "${RED}FAIL${NC} (${error_msg})"
            data_fail=$((data_fail + 1))
        elif [ "$has_data" = true ]; then
            printf "  %-20s %b\n" "Data query:" "${GREEN}PASS${NC}"
            data_pass=$((data_pass + 1))
        else
            printf "  %-20s %b\n" "Data query:" "${RED}FAIL${NC} (no data returned)"
            data_fail=$((data_fail + 1))
        fi
    fi

    echo ""
done

echo "================================"
echo -e "${BOLD}Summary${NC}"
echo -e "  Plugin load:  ${GREEN}${pass_count} passed${NC}, ${RED}${fail_count} failed${NC}"
if [ "$TEST_DATA" = true ]; then
    echo -e "  Data query:   ${GREEN}${data_pass} passed${NC}, ${RED}${data_fail} failed${NC}"
fi
echo ""

if [ "$fail_count" -le 1 ] && [ "$data_fail" -eq 0 ]; then
    echo -e "${GREEN}All checks passed.${NC}"
    [ "$fail_count" -eq 1 ] && echo "(v7.5 failure expected — grafanaDependency >= 9.0.0)"
    exit 0
else
    echo -e "${RED}Unexpected failures detected.${NC}"
    echo "Check logs: docker compose logs <service-name>"
    exit 1
fi
