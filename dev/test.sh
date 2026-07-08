#!/usr/bin/env bash
#
# OCI Metrics Plugin — compatibility test runner
#
# Tests the plugin across Grafana v7.5, v9, v10, v11, and v12.
#
# Usage:
#   ./test.sh                  # build, start, verify plugin load + health
#   ./test.sh --tunnel         # start SSH tunnel + build + start + verify
#   ./test.sh --tunnel --data  # tunnel + build + start + verify with data queries
#   ./test.sh --data           # build + start + verify with data (no tunnel)
#   ./test.sh --status         # show tunnel and container status
#   ./test.sh --down           # stop everything (containers + SSH tunnel)

set -euo pipefail
DEV_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$DEV_DIR/.." && pwd)"
CONFIG_DIR="$DEV_DIR/config"
TUNNEL_PID_FILE="$DEV_DIR/.tunnel.pid"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BOLD='\033[1m'
NC='\033[0m'

stop_tunnel() {
    if [ -f "$TUNNEL_PID_FILE" ]; then
        pid=$(cat "$TUNNEL_PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid" 2>/dev/null
            echo "SSH tunnel stopped (PID $pid)"
        fi
        rm -f "$TUNNEL_PID_FILE"
    fi
    # Fallback: kill any orphaned tunnel processes for port 8169
    pkill -f "ssh.*-L 8169:169.254.169.254" 2>/dev/null || true
}

# --- Parse flags ---
FLAG_TUNNEL=false
FLAG_DATA=false
FLAG_DOWN=false
FLAG_STATUS=false

for arg in "$@"; do
    case "$arg" in
        --help|-h)
            echo "Usage: ./test.sh [OPTIONS]"
            echo ""
            echo "Tests the OCI Metrics plugin across Grafana v7.5, v9, v10, v11, v12."
            echo ""
            echo "Options:"
            echo "  (none)     Build, start, and verify plugin load + health check"
            echo "  --tunnel   Start SSH tunnel for Instance Principal auth"
            echo "  --data     Also test data queries against OCI"
            echo "  --status   Show tunnel and container status, then exit"
            echo "  --down     Stop all containers and SSH tunnel"
            echo "  --help     Show this help"
            echo ""
            echo "Examples:"
            echo "  ./test.sh                  # build + start + verify"
            echo "  ./test.sh --tunnel         # tunnel + build + start + verify"
            echo "  ./test.sh --tunnel --data  # tunnel + build + start + verify with data"
            echo ""
            echo "First-time setup:"
            echo "  cp .env.example .env   # then edit with your values"
            echo "  docker compose pull    # optional: cache images before using a restricted network"
            echo ""
            echo "SSH tunnel:"
            echo "  Set SSH_TUNNEL_HOST in .env and pass --tunnel to start"
            echo "  an IMDS tunnel for Instance Principal auth."
            exit 0
            ;;
        --tunnel)  FLAG_TUNNEL=true ;;
        --data)    FLAG_DATA=true ;;
        --down)    FLAG_DOWN=true ;;
        --status)  FLAG_STATUS=true ;;
        *)
            echo -e "${RED}Unknown option: $arg${NC}"
            echo "Run ./test.sh --help for usage."
            exit 1
            ;;
    esac
done

# --- Handle --down ---
if [ "$FLAG_DOWN" = true ]; then
    echo "Stopping containers..."
    cd "$DEV_DIR" && docker compose down 2>&1 | tail -3
    stop_tunnel
    exit 0
fi

# --- Handle --status ---
if [ "$FLAG_STATUS" = true ]; then
    echo -e "${BOLD}=== SSH Tunnel ===${NC}"
    if [ -f "$TUNNEL_PID_FILE" ] && kill -0 "$(cat "$TUNNEL_PID_FILE")" 2>/dev/null; then
        echo -e "  Process:     ${GREEN}running${NC} (PID $(cat "$TUNNEL_PID_FILE"))"
        if curl -sf -m 3 -H "Authorization: Bearer Oracle" \
            http://localhost:8169/opc/v2/instance/region >/dev/null 2>&1; then
            echo -e "  IMDS check:  ${GREEN}OK${NC}"
        else
            echo -e "  IMDS check:  ${RED}FAIL${NC} (process alive but IMDS not responding)"
        fi
    else
        echo -e "  Process:     ${YELLOW}not running${NC}"
    fi
    echo ""
    echo -e "${BOLD}=== Containers ===${NC}"
    cd "$DEV_DIR" && docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "  No containers running"
    exit 0
fi

# --- Check .env ---
if [ ! -f "$DEV_DIR/.env" ]; then
    echo -e "${RED}No .env file found.${NC}"
    echo ""
    echo "  cp $DEV_DIR/.env.example $DEV_DIR/.env"
    echo "  # Then edit .env with your values"
    exit 1
fi

# Source .env
set -a
source "$DEV_DIR/.env"
set +a

# --- Check Docker images before starting the tunnel or build ---
if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}Docker is not running.${NC} Start Docker and try again."
    exit 1
fi

missing_images=()
while IFS= read -r image; do
    if ! docker image inspect "$image" >/dev/null 2>&1; then
        missing_images+=("$image")
    fi
done < <(cd "$DEV_DIR" && docker compose config --images)

if [ "${#missing_images[@]}" -gt 0 ]; then
    echo -e "${BOLD}Downloading required Grafana images from Docker Hub...${NC}"
    printf '  - %s\n' "${missing_images[@]}"
    echo ""
    if ! (cd "$DEV_DIR" && docker compose pull); then
        echo ""
        echo -e "${RED}Could not download the required Grafana images.${NC}"
        echo "Make sure Docker Hub is reachable from your network."
        echo "If your network or VPN blocks external registries, switch networks and run:"
        echo "  cd $DEV_DIR"
        echo "  docker compose pull"
        exit 1
    fi
    echo ""
fi

# --- Hint if SSH_TUNNEL_HOST is set but --tunnel not passed ---
if [ "$FLAG_TUNNEL" = false ] && [ -n "${SSH_TUNNEL_HOST:-}" ]; then
    echo -e "${YELLOW}Hint:${NC} SSH_TUNNEL_HOST is set. Pass ${BOLD}--tunnel${NC} to start the SSH tunnel."
    echo ""
fi

# --- SSH tunnel (only with --tunnel) ---
if [ "$FLAG_TUNNEL" = true ]; then
    if [ -z "${SSH_TUNNEL_HOST:-}" ]; then
        echo -e "${RED}--tunnel requires SSH_TUNNEL_HOST in .env${NC}"
        exit 1
    fi

    # Check for stale PID file and validate the process is actually our tunnel
    if [ -f "$TUNNEL_PID_FILE" ]; then
        pid=$(cat "$TUNNEL_PID_FILE")
        if kill -0 "$pid" 2>/dev/null && \
           ps -p "$pid" -o command= 2>/dev/null | grep -q "ssh.*8169.*${SSH_TUNNEL_HOST}"; then
            echo -e "SSH tunnel:  ${GREEN}already running${NC} (PID $pid)"
        else
            rm -f "$TUNNEL_PID_FILE"
            # Fall through to start a new tunnel below
        fi
    fi

    # Start a new tunnel if no valid PID file exists
    if [ ! -f "$TUNNEL_PID_FILE" ]; then
        # Check if port 8169 is already in use
        if lsof -i :8169 -sTCP:LISTEN >/dev/null 2>&1; then
            echo -e "${RED}Port 8169 already in use.${NC} Kill the existing process or run ./test.sh --down first."
            exit 1
        fi

        echo -n "SSH tunnel:  starting to ${SSH_TUNNEL_HOST}... "
        ssh -L 8169:169.254.169.254:80 -N -f \
            -o ServerAliveInterval=30 \
            -o ServerAliveCountMax=3 \
            -o ExitOnForwardFailure=yes \
            "$SSH_TUNNEL_HOST" 2>/dev/null
        tunnel_pid=$(pgrep -f "ssh.*-L 8169:169.254.169.254.*${SSH_TUNNEL_HOST}" | tail -1)
        if [ -n "$tunnel_pid" ]; then
            echo "$tunnel_pid" > "$TUNNEL_PID_FILE"
            echo -e "${GREEN}OK${NC} (PID $tunnel_pid)"
        else
            echo -e "${RED}FAIL${NC}"
            echo "Could not start SSH tunnel. Check your SSH config for ${SSH_TUNNEL_HOST}."
            exit 1
        fi

        echo -n "IMDS check:  "
        sleep 1
        if curl -sf -H "Authorization: Bearer Oracle" http://localhost:8169/opc/v2/instance/region >/dev/null 2>&1; then
            echo -e "${GREEN}OK${NC}"
        else
            echo -e "${YELLOW}WARN${NC} (IMDS not responding yet — may need a moment)"
        fi
    fi
    echo ""
fi

# --- Build plugin ---
cd "$REPO_ROOT"
echo -e "${BOLD}Building frontend...${NC}"
yarn build
echo ""
echo -e "${BOLD}Building backend binaries for all platforms (this may take a few minutes)...${NC}"
mage -v
echo ""

# --- Start containers ---
echo -e "${BOLD}Starting Grafana containers...${NC}"
cd "$DEV_DIR"
docker compose up -d
echo ""

# --- Run verification ---
echo -e "${BOLD}Running verification...${NC}"
echo ""

# Pass --data to verify.sh if requested
if [ "$FLAG_DATA" = true ]; then
    "$CONFIG_DIR/verify.sh" "--data"
else
    "$CONFIG_DIR/verify.sh"
fi
rc=$?

# Remind user that containers (and possibly tunnel) are still running
if [ -f "$TUNNEL_PID_FILE" ] && kill -0 "$(cat "$TUNNEL_PID_FILE")" 2>/dev/null; then
    echo -e "\n${YELLOW}Note:${NC} Containers and SSH tunnel (PID $(cat "$TUNNEL_PID_FILE")) are still running."
else
    echo -e "\n${YELLOW}Note:${NC} Containers are still running."
fi
echo -e "      Run ${BOLD}./test.sh --down${NC} to stop everything."
exit $rc
