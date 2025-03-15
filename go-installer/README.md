# Go Installer/Upgrader Script

This bash script installs or upgrades the latest version of the Go programming language from the official source. It removes any apt-installed Go packages to prevent conflicts and ensures that the new Go binary is prioritized in your system's PATH.

## Features

- **Root Privilege Check:** Ensures the script is run as root.
- **Apt Package Removal:** Detects and removes any existing Go installations installed via `apt` (e.g., `golang-go` or `golang`).
- **OS Update & Dependency Installation:** Updates the system and installs necessary dependencies.
- **Go Installation/Upgrade:** Downloads and installs the latest Go release from `go.dev`.
- **PATH Configuration:** Prepends `/usr/local/go/bin` to your PATH for immediate use of the new version.

## Prerequisites

- A Debian/Ubuntu-based system.
- Root privileges (or using `sudo`).

## Usage

1. **Make the script executable:**

   ```bash
   chmod +x go-installer.sh

2. **Run:**

   ```bash
   sudo ./go-installer.sh

