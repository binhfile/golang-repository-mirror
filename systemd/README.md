# systemd Service File

This directory contains a systemd service file for running `go-mod-clone` in server mode as a background service.

## Features

- Runs the Go Module Proxy server as a system service
- Automatic restart on failure
- Journal logging integration
- Security hardening (sandboxed environment)
- Resource limits configured
- User isolation (dedicated `go-mod-clone` user)

## Installation

### Prerequisites

- Linux system with systemd
- `go-mod-clone` binary built and installed
- Root access to manage systemd services

### Step 1: Create storage directory

```bash
sudo mkdir -p /var/lib/go-mod-clone/modules
sudo chmod 0755 /var/lib/go-mod-clone
sudo chmod 0755 /var/lib/go-mod-clone/modules
```

### Step 2: Install the binary

```bash
# Download or build the binary
wget https://github.com/binhfile/golang-repository-mirror/releases/download/v1.2.0/go-mod-clone-linux-amd64

# Install to system location
sudo mv go-mod-clone-linux-amd64 /usr/local/bin/go-mod-clone
sudo chmod +x /usr/local/bin/go-mod-clone
```

### Step 3: Install the service file

```bash
sudo cp go-mod-clone.service /etc/systemd/system/
sudo systemctl daemon-reload
```

### Step 4: Enable and start the service

```bash
# Enable the service to start on boot
sudo systemctl enable go-mod-clone

# Start the service
sudo systemctl start go-mod-clone

# Check the status
sudo systemctl status go-mod-clone
```

## Usage

### View service logs

```bash
# View recent logs
sudo journalctl -u go-mod-clone -n 50

# Follow logs in real-time
sudo journalctl -u go-mod-clone -f

# View logs from the last hour
sudo journalctl -u go-mod-clone --since "1 hour ago"
```

### Manage the service

```bash
# Start the service
sudo systemctl start go-mod-clone

# Stop the service
sudo systemctl stop go-mod-clone

# Restart the service
sudo systemctl restart go-mod-clone

# Check service status
sudo systemctl status go-mod-clone

# View service details
sudo systemctl cat go-mod-clone
```

## Configuration

### Modify service settings

To change the listening port, storage location, or other settings:

```bash
# Edit the service file
sudo systemctl edit --full go-mod-clone
```

Common configuration options in the `[Service]` section:

- `ExecStart`: The command to run the service
  - `--host`: Listen address (default: 0.0.0.0)
  - `--port`: Listen port (default: 3000)
  - `--storage-root`: Module storage location
  - `--log-level`: Log level (debug, info, warn, error)

Example: Change to listen only on localhost and port 8080

```ini
ExecStart=/usr/local/bin/go-mod-clone server \
  --storage-root /var/lib/go-mod-clone/modules \
  --host localhost \
  --port 8080 \
  --log-level info
```

Then reload and restart:

```bash
sudo systemctl daemon-reload
sudo systemctl restart go-mod-clone
```

## Pre-population with modules

Before starting the service, you might want to pre-populate the module cache:

```bash
# Create modules.txt with desired modules
cat > /tmp/modules.txt <<EOF
golang.org/x/crypto@v0.45.0
golang.org/x/crypto@v0.46.0
github.com/gin-gonic/gin@v1.9.1
EOF

# Run the prefill mode
sudo /usr/local/bin/go-mod-clone \
  --modules /tmp/modules.txt \
  --storage-root /var/lib/go-mod-clone/modules \
  --concurrency 4

# Start the service
sudo systemctl start go-mod-clone
```

## Usage with Go

Once the service is running, configure your Go environment:

```bash
export GOPROXY=http://localhost:3000
go get github.com/user/module@version
```

For remote access, replace `localhost` with the server's IP address:

```bash
export GOPROXY=http://192.168.1.100:3000
go get github.com/user/module@version
```

## Troubleshooting

### Service fails to start

Check the logs:
```bash
sudo journalctl -u go-mod-clone -n 100
```

### Permission denied errors

Ensure the storage directory has proper permissions:
```bash
sudo chmod -R 0755 /var/lib/go-mod-clone
```

### Port already in use

If port 3000 is already in use, edit the service:
```bash
sudo systemctl edit --full go-mod-clone
```

Change the `--port` value to an available port (e.g., 3001).

### Out of file descriptors

If you get "too many open files" errors, the service can handle up to 65536:
```bash
# Check current limits
sudo systemctl show -p LimitNOFILE go-mod-clone
```

This is already configured in the service file.

## Security Considerations

The service file includes several security hardening options:

- **NoNewPrivileges**: Prevents privilege escalation
- **PrivateTmp**: Isolated temporary filesystem
- **ProtectSystem**: Read-only access to system directories
- **ProtectHome**: Home directories not accessible
- **ReadWritePaths**: Only the storage directory is writable

To modify these settings:
```bash
sudo systemctl edit --full go-mod-clone
```

## Uninstallation

To completely remove the service:

```bash
# Stop the service
sudo systemctl stop go-mod-clone

# Disable autostart
sudo systemctl disable go-mod-clone

# Remove service file
sudo rm /etc/systemd/system/go-mod-clone.service
sudo systemctl daemon-reload

# Remove storage directory (optional)
sudo rm -rf /var/lib/go-mod-clone
```

## Additional Resources

- [systemd Manual](https://www.freedesktop.org/software/systemd/man/systemd.service.html)
- [go-mod-clone Documentation](https://github.com/binhfile/golang-repository-mirror)
- [Go GOPROXY Documentation](https://golang.org/cmd/go/#hdr-Module_authentication_using_go_sum)
