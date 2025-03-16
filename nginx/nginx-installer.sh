#!/bin/bash
set -euo pipefail

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print messages in color
print_message() {
    local color="$1"
    local message="$2"
    echo -e "${color}${message}${NC}"
}

# Function to check if the user is root
check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        echo
        print_message "$RED" "This script must be run as root."
        echo
        exit 1
    fi
}

# Function to install required dependencies
install_dependencies() {
    print_message "$YELLOW" "Updating package lists..."
    apt update -qq
    print_message "$YELLOW" "Installing required dependencies..."
    apt install -yqq sudo wget curl gnupg2 ca-certificates lsb-release
}

# Function to install Nginx on Ubuntu
install_nginx_ubuntu() {
    print_message "$GREEN" "Detected Ubuntu. Installing Nginx..."
    install_dependencies
    apt install -yqq ubuntu-keyring

    sleep 0.5

    curl -fsSL https://nginx.org/keys/nginx_signing.key | gpg --dearmor | tee /usr/share/keyrings/nginx-archive-keyring.gpg >/dev/null

    sleep 0.5

    gpg --dry-run --quiet --no-keyring --import --import-options import-show /usr/share/keyrings/nginx-archive-keyring.gpg

    sleep 0.5

    echo "deb [signed-by=/usr/share/keyrings/nginx-archive-keyring.gpg] http://nginx.org/packages/ubuntu $(lsb_release -cs) nginx" \
      | tee /etc/apt/sources.list.d/nginx.list

    echo -e "Package: *\nPin: origin nginx.org\nPin: release o=nginx\nPin-Priority: 900\n" \
      | tee /etc/apt/preferences.d/99nginx

    sleep 0.5

    apt update -qq
    apt install -yqq nginx

    sleep 0.5

    nginx_version=$(nginx -v 2>&1)
    print_message "$GREEN" "$nginx_version"
    print_message "$GREEN" "Nginx installed successfully."
}

# Function to install Nginx on Debian
install_nginx_debian() {
    print_message "$GREEN" "Detected Debian. Installing Nginx..."
    install_dependencies
    apt install -yqq debian-archive-keyring

    sleep 0.5

    curl -fsSL https://nginx.org/keys/nginx_signing.key | gpg --dearmor | tee /usr/share/keyrings/nginx-archive-keyring.gpg >/dev/null

    sleep 0.5

    gpg --dry-run --quiet --no-keyring --import --import-options import-show /usr/share/keyrings/nginx-archive-keyring.gpg

    sleep 0.5

    echo "deb [signed-by=/usr/share/keyrings/nginx-archive-keyring.gpg] http://nginx.org/packages/debian $(lsb_release -cs) nginx" \
      | tee /etc/apt/sources.list.d/nginx.list

    echo -e "Package: *\nPin: origin nginx.org\nPin: release o=nginx\nPin-Priority: 900\n" \
      | tee /etc/apt/preferences.d/99nginx

    sleep 0.5

    apt update -qq
    apt install -yqq nginx

    sleep 0.5
    
    nginx_version=$(nginx -v 2>&1)
    print_message "$GREEN" "$nginx_version"
    print_message "$GREEN" "Nginx installed successfully."
}

# Function to check if Nginx is already installed
check_nginx_installed() {
    if command -v nginx >/dev/null 2>&1; then
        nginx_version=$(nginx -v 2>&1)
        print_message "$GREEN" "Nginx is already installed: ${nginx_version}"
        exit 0
    fi
}

# Main Execution

check_root
check_nginx_installed

# Determine OS using /etc/os-release
if [ -f /etc/os-release ]; then
    . /etc/os-release
    case "$ID" in
        ubuntu)
            install_nginx_ubuntu
            ;;
        debian)
            install_nginx_debian
            ;;
        *)
            print_message "$RED" "Unsupported operating system: $ID"
            exit 1
            ;;
    esac
else
    print_message "$RED" "Cannot determine the operating system. /etc/os-release not found."
    exit 1
fi
