#!/bin/bash
# Installation script for go-mod-clone systemd service
# Usage: sudo ./install.sh [--simple] [--port PORT] [--storage-root PATH]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
SERVICE_FILE="go-mod-clone.service"
PORT=3000
STORAGE_ROOT="/var/lib/go-mod-clone/modules"
INSTALL_DIR="/usr/local/bin"
WORK_DIR="/opt/go-mod-clone"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --simple)
            SERVICE_FILE="go-mod-clone-simple.service"
            shift
            ;;
        --port)
            PORT="$2"
            shift 2
            ;;
        --storage-root)
            STORAGE_ROOT="$2"
            shift 2
            ;;
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --help)
            echo "Usage: sudo ./install.sh [options]"
            echo ""
            echo "Options:"
            echo "  --simple              Use simple service file (no security hardening)"
            echo "  --port PORT           Set server port (default: 3000)"
            echo "  --storage-root PATH   Set module storage path (default: /var/lib/go-mod-clone/modules)"
            echo "  --install-dir PATH    Set binary install directory (default: /usr/local/bin)"
            echo "  --help                Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo -e "${YELLOW}Go Module Proxy Server - systemd Installation${NC}"
echo ""

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    echo -e "${RED}Error: This script must be run as root (use sudo)${NC}"
    exit 1
fi

# Check if service file exists
if [[ ! -f "$SERVICE_FILE" ]]; then
    echo -e "${RED}Error: Service file '$SERVICE_FILE' not found${NC}"
    exit 1
fi

echo "Installation Configuration:"
echo "  Service file: $SERVICE_FILE"
echo "  Port: $PORT"
echo "  Storage root: $STORAGE_ROOT"
echo "  Install directory: $INSTALL_DIR"
echo ""

# Step 1: Create user and group
echo -e "${YELLOW}Step 1: Creating system user and group...${NC}"
if ! id -u go-mod-clone &>/dev/null; then
    useradd --system --home /var/lib/go-mod-clone --shell /bin/false go-mod-clone
    echo -e "${GREEN}✓ Created user 'go-mod-clone'${NC}"
else
    echo -e "${GREEN}✓ User 'go-mod-clone' already exists${NC}"
fi

# Step 2: Create directories
echo -e "${YELLOW}Step 2: Creating directories...${NC}"
mkdir -p "$(dirname "$STORAGE_ROOT")"
mkdir -p "$STORAGE_ROOT"
mkdir -p "$WORK_DIR"
echo -e "${GREEN}✓ Created directories${NC}"

# Step 3: Set permissions
echo -e "${YELLOW}Step 3: Setting permissions...${NC}"
chown -R go-mod-clone:go-mod-clone "$(dirname "$STORAGE_ROOT")"
chown -R go-mod-clone:go-mod-clone "$WORK_DIR"
chmod 0755 "$(dirname "$STORAGE_ROOT")"
chmod 0750 "$STORAGE_ROOT"
echo -e "${GREEN}✓ Set permissions${NC}"

# Step 4: Check if binary exists
echo -e "${YELLOW}Step 4: Checking for go-mod-clone binary...${NC}"
if [[ ! -f "$INSTALL_DIR/go-mod-clone" ]]; then
    echo -e "${YELLOW}Note: go-mod-clone binary not found at $INSTALL_DIR/go-mod-clone${NC}"
    echo "Please download and install the binary manually:"
    echo "  wget https://github.com/binhfile/golang-repository-mirror/releases/download/v1.2.0/go-mod-clone-linux-amd64"
    echo "  sudo mv go-mod-clone-linux-amd64 $INSTALL_DIR/go-mod-clone"
    echo "  sudo chmod +x $INSTALL_DIR/go-mod-clone"
else
    echo -e "${GREEN}✓ Binary found${NC}"
fi

# Step 5: Install service file
echo -e "${YELLOW}Step 5: Installing service file...${NC}"
cp "$SERVICE_FILE" /etc/systemd/system/go-mod-clone.service

# Update paths in service file if non-default
if [[ "$PORT" != "3000" || "$STORAGE_ROOT" != "/var/lib/go-mod-clone/modules" ]]; then
    sed -i "s|--port 3000|--port $PORT|g" /etc/systemd/system/go-mod-clone.service
    sed -i "s|/var/lib/go-mod-clone/modules|$STORAGE_ROOT|g" /etc/systemd/system/go-mod-clone.service
fi

systemctl daemon-reload
echo -e "${GREEN}✓ Service file installed${NC}"

# Step 6: Summary
echo ""
echo -e "${GREEN}Installation completed!${NC}"
echo ""
echo "Next steps:"
echo "  1. Enable the service on boot:"
echo "     sudo systemctl enable go-mod-clone"
echo ""
echo "  2. Start the service:"
echo "     sudo systemctl start go-mod-clone"
echo ""
echo "  3. Check service status:"
echo "     sudo systemctl status go-mod-clone"
echo ""
echo "  4. View logs:"
echo "     sudo journalctl -u go-mod-clone -f"
echo ""
echo "Configuration:"
echo "  Service file: /etc/systemd/system/go-mod-clone.service"
echo "  Storage: $STORAGE_ROOT"
echo "  Port: $PORT"
echo ""
echo "Usage with Go:"
echo "  export GOPROXY=http://localhost:$PORT"
echo "  go get github.com/user/module@version"
echo ""
