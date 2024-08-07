#!/bin/bash

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print messages in color
print_message() {
    COLOR=$1
    MESSAGE=$2
    echo -e "${COLOR}${MESSAGE}${NC}"
}

# Function to check if the user is root
check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        echo
        print_message $RED "This script must be run as root"
        echo
        sleep 1
        exit 1
    fi
}

# Function to install required dependencies
install_dependencies() {
    print_message $YELLOW "Updating package lists..."
    echo
    sleep 1
    apt update
    echo
    print_message $YELLOW "Installing required dependencies..."
    echo
    apt install sudo wget curl gnupg2 ca-certificates lsb-release -yqq
}

# Function to install Nginx on Ubuntu
install_nginx_ubuntu() {
    echo
    print_message $GREEN "Detected Ubuntu. Installing Nginx..."
    echo
    sleep 1
    install_dependencies
    apt install ubuntu-keyring -yqq

    curl https://nginx.org/keys/nginx_signing.key | gpg --dearmor | tee /usr/share/keyrings/nginx-archive-keyring.gpg >/dev/null

    gpg --dry-run --quiet --no-keyring --import --import-options import-show /usr/share/keyrings/nginx-archive-keyring.gpg

    echo "deb [signed-by=/usr/share/keyrings/nginx-archive-keyring.gpg] http://nginx.org/packages/ubuntu `lsb_release -cs` nginx" | tee /etc/apt/sources.list.d/nginx.list

    echo -e "Package: *\nPin: origin nginx.org\nPin: release o=nginx\nPin-Priority: 900\n" | tee /etc/apt/preferences.d/99nginx

    apt update
    apt install nginx -yqq
    sleep 0.5
    echo
    nginx_version=$(nginx -v 2>&1)
    print_message $GREEN "$nginx_version"
    echo
    print_message $GREEN "Nginx installed successfully."
    echo
    sleep 1
}

# Function to install Nginx on Debian
install_nginx_debian() {
    echo
    print_message $GREEN "Detected Debian. Installing Nginx..."
    sleep 1
    echo
    install_dependencies
    apt install debian-archive-keyring -yqq

    curl https://nginx.org/keys/nginx_signing.key | gpg --dearmor | tee /usr/share/keyrings/nginx-archive-keyring.gpg >/dev/null

    gpg --dry-run --quiet --no-keyring --import --import-options import-show /usr/share/keyrings/nginx-archive-keyring.gpg

    echo "deb [signed-by=/usr/share/keyrings/nginx-archive-keyring.gpg] http://nginx.org/packages/debian `lsb_release -cs` nginx" | tee /etc/apt/sources.list.d/nginx.list

    echo -e "Package: *\nPin: origin nginx.org\nPin: release o=nginx\nPin-Priority: 900\n" | tee /etc/apt/preferences.d/99nginx

    apt update
    apt install nginx -yqq
    sleep 0.5
    echo
    nginx_version=$(nginx -v 2>&1)
    print_message $GREEN "$nginx_version"
    echo
    print_message $GREEN "Nginx installed successfully."
    echo
    sleep 1
}

# Function to check if Nginx is installed
check_nginx_installed() {
    if command -v nginx >/dev/null 2>&1; then
        VERSION=$(nginx -v 2>&1)
        echo
        print_message $GREEN "Nginx is already installed: ${VERSION}"
        echo
        exit 0
    fi
}

# Check if user is root
check_root

# Check if Nginx is already installed
check_nginx_installed

# Determine the OS and call the appropriate function
if [ -f /etc/lsb-release ]; then
    . /etc/lsb-release
    if [ "$DISTRIB_ID" == "Ubuntu" ]; then
        install_nginx_ubuntu
    else
        echo
        print_message $RED "Unsupported Ubuntu variant: $DISTRIB_ID"
        echo
    fi
elif [ -f /etc/debian_version ]; then
    install_nginx_debian
else
    echo
    print_message $RED "Unsupported operating system."
    echo
    exit 1
fi
