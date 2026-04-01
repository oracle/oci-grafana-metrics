# Local Testing with Instance Principal Auth

How to test the OCI Metrics Grafana plugin locally in Docker using Instance Principal auth via SSH tunnel.

## Prerequisites

- An OCI compute instance with SSH access
- The instance must be in a dynamic group with IAM policies for monitoring/identity API access
- SSH configured for the instance
- Docker or Rancher Desktop running on your machine

## How It Works

The OCI Go SDK supports an environment variable `OCI_METADATA_BASE_URL` that overrides the Instance Metadata Service (IMDS) endpoint. By SSH-tunneling IMDS from an OCI instance to your machine, the plugin can authenticate as that instance. The `entrypoint.sh` script automatically detects the container-to-host hostname and constructs `OCI_METADATA_BASE_URL` at container startup.

```
YOUR MACHINE                                        OCI COMPUTE INSTANCE
┌──────────────────────┐     SSH Tunnel             ┌─────────────────────┐
│ Docker container     │  port 8169 ← 169.254...    │                     │
│ ├─ Grafana           │ ◄────────────────────────► │  IMDS               │
│ └─ OCI plugin        │                            │  (169.254.169.254)  │
│      │               │                            └─────────────────────┘
│      └───────────────┼── direct ──► telemetry.<region>.oraclecloud.com
└──────────────────────┘              identity.<region>.oraclecloud.com
```

Only IMDS traffic goes through the tunnel. OCI API calls go directly from your machine.

## Steps

### 1. Start the SSH tunnel

```bash
ssh -L 8169:169.254.169.254:80 -N <your-oci-instance-hostname>
```

This forwards your local port 8169 to the IMDS service on the OCI instance.

### 2. Verify the tunnel works

```bash
curl -H "Authorization: Bearer Oracle" http://localhost:8169/opc/v2/instance/region
```

You should see the instance's region (e.g., `us-ashburn-1`).

### 3. Start containers

Use `test.sh` with the `--tunnel` flag to handle the tunnel, containers, and verification in one step:

```bash
cd dev
./test.sh --tunnel        # handles tunnel, containers, and verification
```

Or start manually with Docker Compose:

```bash
cd dev
docker compose up -d
```

### 4. Verify

The datasource and dashboard are auto-provisioned. Open Grafana at http://localhost:3075 (v7.5), http://localhost:3090 (v9), etc.

Or run the verification script:

```bash
./verify.sh --data
```

## Token Refresh

The security token lasts ~1 hour. The SDK automatically refreshes it by re-fetching certs from IMDS through the tunnel. **Keep the SSH tunnel running** for as long as you're testing. If the tunnel drops, auth will fail on the next refresh.

## Troubleshooting

### Tunnel dropped / auth stopped working

Check the current status:

```bash
./test.sh --status
```

Restart the tunnel and containers:

```bash
./test.sh --tunnel
```

Or restart manually:

```bash
ssh -L 8169:169.254.169.254:80 -N <your-oci-instance-hostname>
docker compose restart
```

### Health check takes a long time then fails

The plugin is likely trying to reach `169.254.169.254` directly. Verify the SSH tunnel is running (`./test.sh --status`) and restart if needed (`./test.sh --tunnel`). The container-to-host hostname is auto-detected by `entrypoint.sh` — if auto-detection fails, you can set `OCI_METADATA_BASE_URL` directly in your `.env` file (e.g., `OCI_METADATA_BASE_URL=http://host.docker.internal:8169/opc/v2`).

## Reference

- `OCI_METADATA_BASE_URL` is defined in the OCI Go SDK at `common/auth/instance_principal_key_provider.go`
- The env var overrides all 4 IMDS endpoints: region, cert.pem, key.pem, intermediate.pem
- Security tokens are JWT-based and not IP-bound, so API calls from your machine are accepted
