#!/bin/bash
set -euo pipefail

# Function to check if the script is being run as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo -e "\nPlease run this script as root.\n"
        exit 1
    fi
}

# Function to remove Go installed via apt (if any)
remove_apt_go() {
    echo -e "\nChecking for apt-installed Go packages...\n"

    if dpkg -s golang-go >/dev/null 2>&1; then
        echo "Found 'golang-go' installed via apt. Removing it..."
        apt remove -y golang-go
    else
        echo "No apt-installed 'golang-go' package found."
    fi

    if dpkg -s golang >/dev/null 2>&1; then
        echo "Found 'golang' installed via apt. Removing it..."
        apt remove -y golang
    else
        echo "No apt-installed 'golang' package found."
    fi
}

# Function to install dependencies and update the OS
install_package() {
    echo -e "\nUpdating the OS...\n"
    apt update -q && apt upgrade -y
    echo -e "\nOS Updated & Upgraded.\n"

    echo -e "\nInstalling Dependencies...\n"
    apt install -y curl wget build-essential ca-certificates git
    echo -e "\nDependencies Installed.\n"
}

# Function to download and install or upgrade Go
install_go() {
    echo -e "\nInstalling or Upgrading Go...\n"
    local latest_version
    latest_version=$(curl -sL https://golang.org/VERSION?m=text | head -1)
    echo "Latest Go version available: $latest_version"
    local go_url="https://go.dev/dl/${latest_version}.linux-amd64.tar.gz"
    
    if command -v go &>/dev/null; then
        local current_version
        current_version=$(go version | awk '{print $3}')
        echo "Current installed Go version: $current_version"
        if [ "$current_version" == "$latest_version" ]; then
            echo "Go is already up-to-date ($current_version)."
            return
        else
            echo "Upgrading Go from $current_version to $latest_version."
        fi
    else
        echo "Installing Go version $latest_version."
    fi

    # Remove any existing Go installation in /usr/local/go
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

    # Prepend Go binary path to the PATH environment variable
    export PATH=/usr/local/go/bin:$PATH
    echo "export PATH=/usr/local/go/bin:\$PATH" > /etc/profile.d/go.sh
    source /etc/profile.d/go.sh

    # Verify the installation
    go version

    echo -e "\nGo installed successfully.\n"
}

# Main Execution
check_root
remove_apt_go
install_package
install_go
