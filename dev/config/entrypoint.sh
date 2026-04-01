#!/bin/bash
set -e

# Process dashboard JSON templates: replace ${VAR} placeholders with env var values.
# Datasource YAML and dashboard-provider YAML are static and mounted directly.

TEMPLATE_DIR="/etc/grafana/provisioning-templates/dashboards"
OUTPUT_DIR="/etc/grafana/provisioning/dashboards"

if [ -d "$TEMPLATE_DIR" ]; then
    mkdir -p "$OUTPUT_DIR"
    for tmpl in "$TEMPLATE_DIR"/*.template; do
        [ -f "$tmpl" ] || continue
        outfile="$OUTPUT_DIR/$(basename "$tmpl" .template)"
        sed \
            -e "s|\${OCI_COMPARTMENT_OCID}|${OCI_COMPARTMENT_OCID:-MISSING_COMPARTMENT_OCID}|g" \
            -e "s|\${OCI_REGION_1}|${OCI_REGION_1:-us-ashburn-1}|g" \
            -e "s|\${OCI_REGION_2}|${OCI_REGION_2:-us-phoenix-1}|g" \
            -e "s|\${OCI_NAMESPACE}|${OCI_NAMESPACE:-oci_lbaas}|g" \
            -e "s|\${OCI_METRIC}|${OCI_METRIC:-PeakBandwidth}|g" \
            "$tmpl" > "$outfile"
        echo "Processed: $(basename "$tmpl") -> $(basename "$outfile")"
    done

    # Copy static YAML files alongside processed JSON
    for f in "$TEMPLATE_DIR"/*.yaml; do
        [ -f "$f" ] || continue
        cp "$f" "$OUTPUT_DIR/$(basename "$f")"
    done
fi

# Auto-detect container-to-host hostname and construct OCI_METADATA_BASE_URL.
# Port 8169 is the SSH tunnel port set by test.sh.
if [ -z "${OCI_METADATA_BASE_URL:-}" ]; then
    metadata_host=""
    if command -v getent >/dev/null 2>&1; then
        if getent hosts host.docker.internal >/dev/null 2>&1; then
            metadata_host="host.docker.internal"
        elif getent hosts host.lima.internal >/dev/null 2>&1; then
            metadata_host="host.lima.internal"
        else
            metadata_host="172.17.0.1"
        fi
    else
        echo "IMDS host: using fallback 172.17.0.1 (getent not available)"
        metadata_host="172.17.0.1"
    fi
    export OCI_METADATA_BASE_URL="http://${metadata_host}:8169/opc/v2"
fi

exec /run.sh "$@"
