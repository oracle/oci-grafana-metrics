# Dev Environment

Tests the OCI Metrics plugin across Grafana v7.5, v9, v10, v11, and v12.

## Quick Start

```bash
# 1. Configure (one-time)
cp .env.example .env     # edit with your OCI compartment, regions, etc.

# 2. Test
./test.sh                    # build + start + verify plugin load & health
./test.sh --tunnel           # same, with SSH tunnel for Instance Principal auth
./test.sh --tunnel --data    # also test real data queries against OCI
./test.sh --status           # check tunnel and container status
./test.sh --down             # stop everything
```

That's it. The script handles building, containers, and verification.

## What `test.sh` Does

1. Starts SSH tunnel to your OCI instance (if `--tunnel` flag is passed)
2. Builds the plugin (`yarn build` + `mage`)
3. Starts 5 Grafana containers (v7.5, v9, v10, v11, v12)
4. Verifies plugin load, health check, and optionally data queries

## .env Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `OCI_COMPARTMENT_OCID` | Yes | Your compartment OCID for test queries |
| `OCI_REGION_1` | Yes | First dashboard panel region (default: `us-ashburn-1`) |
| `OCI_REGION_2` | Yes | Second dashboard panel region (default: `us-phoenix-1`) |
| `OCI_NAMESPACE` | No | OCI namespace (default: `oci_lbaas`) |
| `OCI_METRIC` | No | Metric name (default: `PeakBandwidth`) |
| `SSH_TUNNEL_HOST` | No | OCI instance hostname for `--tunnel` |

## Grafana Instances

| Version | Port | URL |
|---------|------|-----|
| v7.5 | 3075 | http://localhost:3075 |
| v9 | 3090 | http://localhost:3090 |
| v10 | 3100 | http://localhost:3100 |
| v11 | 3110 | http://localhost:3110 |
| v12 | 3120 | http://localhost:3120 |

## How It Works

Everything is driven by `.env`. The `docker-compose.yaml` defines the Grafana containers, and the `config/` directory contains internals (entrypoint, provisioning) that you don't need to touch. Dashboard JSON uses `${VAR}` placeholders that are processed by `sed` at container startup via a custom entrypoint.

For details on the SSH tunnel approach, see [config/local-testing-with-instance-principal.md](config/local-testing-with-instance-principal.md).
