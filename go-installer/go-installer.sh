#!/bin/bash
set -euo pipefail

# Function to check if the script is being run as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo -e "\nPlease run this script as root.\n"
        exit 1
    fi
}

# Install Dependencies
install_package() {
    echo -e "\nUpdating the OS...\n"
    apt update -q && apt upgrade -y
    echo -e "\nOS Updated & Upgraded.\n"

    echo -e "\nInstalling Dependencies...\n"
    apt install -y curl wget build-essential ca-certificates git
    echo -e "\nDependencies Installed.\n"
}

# Function to download and install Go
install_go() {
    echo -e "\nInstalling Go...\n"
    local go_version
    go_version=$(curl -sL https://golang.org/VERSION?m=text | head -1)
    local go_url="https://go.dev/dl/${go_version}.linux-amd64.tar.gz"

    # Remove any existing Go installation
    rm -rf /usr/local/go

    # Download the Go archive with error handling
    if ! curl -sLo go.tar.gz "$go_url"; then
        echo "Failed to download Go from $go_url"
        exit 1
    fi

    # Extract the Go archive and check for errors
    if ! tar -C /usr/local/ -xzf go.tar.gz; then
        echo "Failed to extract Go archive"
        rm go.tar.gz
        exit 1
    fi
    rm go.tar.gz

    # Add Go binary path to system PATH
    export PATH=$PATH:/usr/local/go/bin
    echo "export PATH=\$PATH:/usr/local/go/bin" > /etc/profile.d/go.sh

    # Source the updated PATH
    source /etc/profile.d/go.sh

    # Check installed Go version
    go version

    echo -e "\nGo installed successfully.\n"
}

# Main Execution
check_root
install_package
install_go
