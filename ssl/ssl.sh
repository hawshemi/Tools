#!/bin/bash

clear


# Function for green messages
green_msg() {
    echo -e "\033[0;32m[*] ----- $1\033[0m" # Green
}


# Function for red messages
red_msg() {
    echo -e "\033[0;31m[*] ----- $1\033[0m" # Red
}


# Function for cyan messages
cyan_msg() {
    echo -e "\033[0;36m[*] ----- $1\033[0m" # Cyan
}


# Intro
echo 
cyan_msg '================================================================================'
cyan_msg 'This script will automatically Obtain, Revoke, and Renew your SSL Certificates.'
cyan_msg 'Tested on: Ubuntu 20+, Debian 11+'
cyan_msg 'Root access is required.' 
cyan_msg 'Source is @ https://github.com/hawshemi/ssl' 
cyan_msg '================================================================================'
echo 


# Check if the script is run as root
check_root() {
    if [ "$(id -u)" != "0" ]; then
        red_msg "Error: This script must be run as root."
        echo 
        sleep 0.5
        exit 1
    else
        green_msg "Running as root, continuing..."
        sleep 0.5
    fi
}


# Validate domain format
validate_domain() {
    if [[ $1 =~ ^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$ ]]; then
        green_msg "Domain validation passed for: $1"
        echo 
        sleep 0.5
        return 0
    else
        red_msg "Validation Error: The domain '$1' is not in a valid format."
        echo 
        sleep 0.5
        return 1
    fi
}


# Install Socat if it's not already installed
install_socat() {
    if ! command -v socat &> /dev/null; then
        sudo apt update -q
        sudo apt install -y socat || red_msg "Failed to install socat."
        echo 
        sleep 0.5
    else
        green_msg "Socat is already installed."
        echo 
        sleep 0.5
    fi
}


# Allow port 80 with ufw
allow_port_80() {
    if sudo ufw status | grep -q active; then
        sudo ufw allow 80 || red_msg "Failed to allow port 80."
        echo 
        sleep 0.5
    else
        green_msg "Port 80 is already allowed."
        echo 
        sleep 0.5
    fi
}


# Install and configure ACME
install_acme() {
    curl https://get.acme.sh | sudo sh || red_msg "Failed to install ACME.sh."
    sleep 0.5
    ~/.acme.sh/acme.sh --upgrade --auto-upgrade || red_msg "Failed to set up ACME.sh auto-upgrade."
    sleep 0.5
    ~/.acme.sh/acme.sh --set-default-ca --server letsencrypt || red_msg "Failed to set default CA to Let’s Encrypt."
    sleep 0.5
}


# Function to clean up the SSL certificate directory
cleanup_ssl_dir() {
    local domain_name=$1
    local cert_dir="/etc/ssl/${domain_name}"
    if [ -d "$cert_dir" ]; then
        sudo rm -rf "$cert_dir"
        echo 
        red_msg "Removed certificate directory for $domain_name due to errors."
    fi
}


# Apply and install the SSL certificate
apply_install_ssl() {
    local domain_name=$1
    local cert_dir="/etc/ssl/${domain_name}"
    mkdir -p "${cert_dir}" || { red_msg "Failed to create certificate directory for $domain_name."; exit 1; }
    echo 
    
    ~/.acme.sh/acme.sh --issue -d "$domain_name" --standalone --keylength ec-256 || { cleanup_ssl_dir "$domain_name"; red_msg "Failed to issue certificate for $domain_name. Please check log."; echo ;  exit 1; }
    
    ~/.acme.sh/acme.sh --install-cert -d "$domain_name" --ecc \
        --fullchain-file "${cert_dir}/${domain_name}_fullchain.cer" \
        --key-file "${cert_dir}/${domain_name}_private.key" || { cleanup_ssl_dir "$domain_name"; red_msg "Failed to install certificate for $domain_name."; exit 1; }
    
    sudo chown -R nobody:nogroup "${cert_dir}" || { cleanup_ssl_dir "$domain_name"; red_msg "Failed to change owner and group of ${cert_dir}."; exit 1; }
    
    green_msg "SSL certificate obtained and installed for $domain_name."
    echo 
    green_msg "Fullchain:    ${cert_dir}/${domain_name}_fullchain.cer"
    green_msg "Private:      ${cert_dir}/${domain_name}_private.key"
    echo 
}


# Function to revoke and clean SSL certificate
revoke_ssl() {
    local domain_name=$1
    local cert_dir="/etc/ssl/${domain_name}"

    ~/.acme.sh/acme.sh --revoke -d "$domain_name" --ecc || red_msg "Failed to revoke certificate for $domain_name."
    if [ -d "$cert_dir" ]; then
        sudo rm -rf "$cert_dir" || red_msg "Failed to remove certificate directory for $domain_name."
        green_msg "Removed certificate directory for $domain_name."
        sleep 0.5
    else
        green_msg "Certificate directory for $domain_name does not exist, so there is no need to remove it."
        sleep 0.5
    fi
    ~/.acme.sh/acme.sh --remove -d "$domain_name" --ecc || red_msg "Failed to remove certificate data for $domain_name."
    green_msg "SSL certificate revoked and cleaned for $domain_name."
    echo 
    sleep 0.5
}


# Function to force renewal of SSL certificate
force_renew_ssl() {
    local domain_name=$1
    local cert_dir="/etc/ssl/${domain_name}"

    ~/.acme.sh/acme.sh --renew -d "$domain_name" --force --ecc || red_msg "Failed to renew certificate for $domain_name."
    sleep 0.5

    ~/.acme.sh/acme.sh --install-cert -d "$domain_name" --ecc \
        --fullchain-file "${cert_dir}/${domain_name}_fullchain.cer" \
        --key-file "${cert_dir}/${domain_name}_private.key" || red_msg "Failed to install renewed certificate for $domain_name."
    green_msg "SSL certificate forcefully renewed for $domain_name."
    echo 
    sleep 0.5
}


# Main function
main() {
    check_root

    while true; do
        echo 
        cyan_msg "Choose an option:"
        cyan_msg "1. Get SSL"
        cyan_msg "2. Revoke SSL"
        cyan_msg "3. Force Renew SSL"
        cyan_msg "q. Exit"
        echo 
        read -p "Enter choice: " choice

        case $choice in
            1)
                echo 
                read -p "Enter your domain name (e.g., my.example.com): " domain_name
                echo 
                if validate_domain "$domain_name"; then
                    install_socat
                    allow_port_80
                    install_acme
                    apply_install_ssl "$domain_name"
                else
                    echo 
                    red_msg "Invalid domain name. Please enter a valid domain name."
                    echo 
                fi
                ;;
            2)
                echo 
                read -p "Enter the domain name of the SSL to revoke (e.g., my.example.com): " domain_name
                echo 
                if validate_domain "$domain_name"; then
                    revoke_ssl "$domain_name"
                else
                    echo 
                    red_msg "Invalid domain name. Please enter a valid domain name."
                    echo 
                fi
                ;;
            3)
                echo 
                read -p "Enter the domain name for the SSL to force renewal (e.g., my.example.com): " domain_name
                echo 
                if validate_domain "$domain_name"; then
                    force_renew_ssl "$domain_name"
                else
                    echo 
                    red_msg "Invalid domain name. Please enter a valid domain name."
                    echo 
                fi
                ;;
            q)
                echo 
                green_msg "Script Exited."
                echo 
                exit 0
                ;;
            *)
                echo 
                red_msg "Invalid choice, please choose from the list."
                echo 
                ;;
        esac
    done
}


# Run the main function
main "$@"
