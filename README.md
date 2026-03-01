# deply

Push it, roll it, own it.

A simple and efficient remote deployment tool that enables fast application deployment, rollback, and version management via SSH.

## Features

- 🚀 **Fast Deployment** - Pack locally, extract remotely, deploy with one command
- 🔄 **Version Rollback** - Quickly rollback to any historical version
- 📜 **History Records** - View all deployed versions and their status
- 🔗 **Symlink Management** - Zero-downtime switching using symbolic links
- 🪝 **Hook Support** - Execute custom commands before and after deployment
- 📦 **Flexible Packaging** - Support for specifying include/exclude files
- 🔐 **SSH Authentication** - Support both password and key-based authentication
- ⚡ **Multi-host Deployment** - Deploy to multiple servers simultaneously

## Installation

Build from source:

```bash
git clone https://github.com/chihqiang/deply.git
cd deply
go build -o deply .
```

## Quick Start

### Basic Deployment

```bash
deply publish --hosts user@host:22 --dir ./dist --version v1.0.0
```

### Using SSH Key

```bash
deply publish --hosts user@host:22 --key ~/.ssh/id_rsa
```

### Multi-host Deployment

```bash
deply publish --hosts user@host1:22,user@host2:22 --dir ./dist
```

## Commands

### `publish` - Deploy Application

Pack local directory and deploy to remote servers.

```bash
deply publish [options]
```

**Options:**

| Option | Env Variable | Default | Description |
|--------|--------------|---------|-------------|
| `--hosts` | `DEPLY_HOSTS` | - | Remote host list, format: `user[:password]@host[:port]` (required) |
| `--key` | `DEPLY_KEY` | - | Path to SSH private key |
| `--passphrase` | `DEPLY_PASSPHRASE` | - | Passphrase for SSH private key |
| `--timeout` | `DEPLY_TIMEOUT` | 30s | SSH connection timeout |
| `--dir` | - | Current directory | Local directory to deploy |
| `--version, -V` | - | Timestamp | Deployment version number |
| `--include` | - | - | Files or directories to include when packaging (relative to --dir) |
| `--exclude` | - | - | Files or directories to exclude when packaging (relative to --dir) |
| `--remote-repo` | - | `/data/wwwroot/{project}/releases` | Remote version storage directory |
| `--current-link` | - | `/data/wwwroot/{project}/current` | Symbolic link path for current version |
| `--hook-pre-host` | `DEPLY_HOOK_PRE` | - | Remote command to run before deployment |
| `--hook-post-host` | `DEPLY_HOOK_POST` | - | Remote command to run after deployment |

**Examples:**

```bash
# Deploy dist directory to remote server
deply publish \
  --hosts deploy@192.168.1.100:22 \
  --dir ./dist \
  --version v1.0.0 \
  --remote-repo /data/app/releases \
  --current-link /data/app/current

# Use key authentication with post-deployment hook
deply publish \
  --hosts deploy@192.168.1.100:22 \
  --key ~/.ssh/id_rsa \
  --dir ./dist \
  --hook-post-host "systemctl restart nginx"
```

### `rollback` - Version Rollback

Rollback application to a specified historical version.

```bash
deply rollback [options]
```

**Options:**

| Option | Description |
|--------|-------------|
| `--hosts` | Remote host list (same as publish) |
| `--version, -V` | Version number to rollback to |
| `--remote-repo` | Remote version storage directory |
| `--current-link` | Symbolic link path for current version |

**Example:**

```bash
# Rollback to v1.0.0
deply rollback \
  --hosts deploy@192.168.1.100:22 \
  --version v1.0.0 \
  --remote-repo /data/app/releases \
  --current-link /data/app/current
```

### `history` - View History

View deployment history on remote servers.

```bash
deply history [options]
```

**Options:**

| Option | Description |
|--------|-------------|
| `--hosts` | Remote host list (same as publish) |
| `--remote-repo` | Remote version storage directory |
| `--current-link` | Symbolic link path for current version |

**Example:**

```bash
deply history \
  --hosts deploy@192.168.1.100:22 \
  --remote-repo /data/app/releases \
  --current-link /data/app/current
```

Output example:

```
Host             Path                        Version          Current  ModTime
192.168.1.100    /data/app/releases          v1.0.2           false    2026-03-01 10:30:00
192.168.1.100    /data/app/releases          v1.0.1           false    2026-02-28 15:20:00
192.168.1.100    /data/app/releases          v1.0.0           true     2026-02-27 09:10:00
```

## Deployment Flow

### Publish Flow

```apl
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Local Pack │ -> │  SSH Upload │ -> │Remote Extract│
└─────────────┘    └─────────────┘    └─────────────┘
                                            │
                                            v
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Verify     │ <- │  Update Link│ <- │  Run Hooks  │
└─────────────┘    └─────────────┘    └─────────────┘
```

1. Load remote host configuration
2. Load deployment configuration
3. Pack local directory as `tar.gz` file
4. Iterate through all hosts and execute deployment:
   - Establish SSH connection
   - Upload archive to version directory
   - Extract and delete archive
   - Execute pre-hook (if configured)
   - Update currentLink symbolic link
   - Execute post-hook (if configured)
   - Verify deployment result

### Directory Structure

Directory structure on remote server:

```
/data/wwwroot/myapp/
├── releases/                    # Version storage directory
│   ├── 20260301103000/         # Version directory (timestamp)
│   │   ├── index.html
│   │   └── assets/
│   ├── 20260228152000/
│   └── v1.0.0/                 # Version directory (custom version)
│       ├── index.html
│       └── assets/
└── current -> releases/v1.0.0  # Symbolic link to current version
```

## Environment Variables

Simplify commands using environment variables:

```bash
export DEPLY_HOSTS="deploy@192.168.1.100:22"
export DEPLY_KEY="~/.ssh/id_rsa"
export DEPLY_HOOK_POST="systemctl restart nginx"

# Then run simply
deply publish --dir ./dist --version v1.0.0
```

## Important Notes

1. **Symlink Safety**: If `currentLink` path already exists and is not a symbolic link, deployment will fail and requires manual handling
2. **Version Uniqueness**: Version numbers cannot be duplicated; existing version directories will not be overwritten
3. **Multi-host Fault Tolerance**: Deployment failure on one host will not affect other hosts
4. **Temp File Cleanup**: Local temporary packaged files will be automatically deleted after deployment

## Development

### Project Structure

```
.
├── main.go           # Program entry
├── cmdx/             # Command implementations
│   ├── publish.go    # Publish command
│   ├── rollback.go   # Rollback command
│   └── history.go    # History command
├── depx/             # Deployment core logic
│   ├── config.go     # Configuration definition
│   ├── load.go       # Configuration loading
│   ├── pack.go       # Packaging functionality
│   └── execute.go    # Execute deployment
├── sshx/             # SSH related functionality
│   ├── load.go       # Configuration loading
│   ├── ssh.go        # SSH connection
│   └── sftp.go       # SFTP file transfer
├── flagx/            # Command line flags
│   └── flag.go       # Flag definitions
└── utilx/            # Utility functions
    ├── progress.go   # Progress bar
    └── table.go      # Table output
```

## License

Apache License 2.0
