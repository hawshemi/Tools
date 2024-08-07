#!/bin/bash


clear


set -e  # Exit immediately if a command exits with a non-zero status.


# Green, Yellow & Red Messages.
green_msg() {
    tput setaf 2
    echo "[*] ----- $1"
    tput sgr0
}

yellow_msg() {
    tput setaf 3
    echo "[*] ----- $1"
    tput sgr0
}

red_msg() {
    tput setaf 1
    echo "[*] ----- $1"
    tput sgr0
}

# Handle errors and exit
handle_error() {
    red_msg "$1"
    exit 1
}


# Clean up unnecessary files
cleanup() {
    rm -f prometheus-*.tar.gz node_exporter-*.tar.gz grafana-enterprise_*.deb
}

# Retrieve the server's IP address
get_server_ip() {
    local ip_sources=("https://ipv4.icanhazip.com" "https://api.ipify.org" "https://ipv4.ident.me/")
    local server_ip

    for source in "${ip_sources[@]}"; do
        server_ip=$(curl -s --max-time 10 --retry 3 "$source")
        if [ -n "$server_ip" ]; then
            echo "$server_ip"
            return 0
        fi
    done

    red_msg "Unable to retrieve a valid IP address from any source."
    exit 1
}

# Determine the system architecture
get_architecture() {
    case "$(uname -m)" in
        x86_64)
            echo "amd64"
            ;;
        aarch64)
            echo "arm64"
            ;;
        *)
            echo "unsupported"
            ;;
    esac
}

# Get the latest release version from GitHub
get_latest_release() {
    local repo="$1"
    curl --silent "https://api.github.com/repos/${repo}/releases/latest" | # Get latest release from GitHub API
        grep '"tag_name":' |                                               # Get tag line
        sed -E 's/.*"v([^"]+)".*/\1/'                                      # Extract version without 'v' prefix
}

# Wait for a service to become active
wait_for_service() {
    local service_name="$1"
    local retries=10
    local count=0

    until sudo systemctl is-active --quiet "$service_name"; do
        sleep 1
        count=$((count + 1))
        if [ "$count" -ge "$retries" ]; then
            red_msg "$service_name did not start in time."
            exit 1
        fi
    done
}

# Wait for dpkg lock
wait_for_dpkg_lock() {
    while fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1; do
        yellow_msg "Waiting for dpkg lock to be released..."
        sleep 2
    done
}

# Ensure running as root
ensure_root() {
    if [ "$EUID" -ne 0 ]; then
        echo
        red_msg "This script must be run as root."
        echo 
        sleep 1
        exit 1
    fi
}

prepaire_vps() {
    echo
    yellow_msg "Update, Upgrade, Install dependencies..."
    echo
    sleep 1

    sudo apt update -q || handle_error "Failed to update package lists"
    sudo apt upgrade -yq || handle_error "Failed to upgrade packages"
    sudo apt autopurge -yq || handle_error "Failed to auto-purge packages"
    sudo apt install -yq build-essential sudo wget curl adduser libfontconfig1 musl || handle_error "Failed to install essential packages."

    echo
    green_msg "Done."
    echo

}

# Generic function to install a service from GitHub releases
install_service() {
    local name="$1"
    local repo="$2"

    # Get the latest version
    local VERSION=$(get_latest_release "$repo" | tr -d 'v')

    # Set file names based on architecture
    local FILE="${name}-${VERSION}.linux-${ARCH}.tar.gz"
    local DIR="${name}-${VERSION}.linux-${ARCH}"

    # Download and extract
    wget "https://github.com/${repo}/releases/download/v${VERSION}/${FILE}" || handle_error "Failed to download $name"
    tar xzf "$FILE" || handle_error "Failed to extract $name"
    sudo mv "$DIR" "/etc/$name" || handle_error "Failed to move $name directory to /etc/$name"

    # Create systemd service file
    cat <<EOL | sudo tee /etc/systemd/system/${name}.service
[Unit]
Description=$name
Wants=network-online.target
After=network-online.target

[Service]
ExecStart=/etc/$name/$name
Restart=always

[Install]
WantedBy=multi-user.target
EOL

    # Reload, restart, enable, and check status of the service
    sudo systemctl daemon-reload || handle_error "Failed to reload systemd"
    sudo systemctl restart $name || handle_error "Failed to restart $name"
    wait_for_service "$name"
    sudo systemctl enable "$name" || handle_error "Failed to enable $name"
}

install_prometheus() {
    echo
    yellow_msg "Installing Prometheus..."
    echo
    sleep 1

    # Install Prometheus
    install_service "prometheus" "prometheus/prometheus"

    # Create Prometheus systemd service file
    cat <<EOL | sudo tee /etc/systemd/system/prometheus.service
[Unit]
Description=Prometheus
Wants=network-online.target
After=network-online.target

[Service]
ExecStart=/etc/prometheus/prometheus --config.file=/etc/prometheus/prometheus.yml
Restart=always

[Install]
WantedBy=multi-user.target
EOL

    # Reload, restart, enable, and check status of Prometheus
    sudo systemctl daemon-reload || handle_error "Failed to reload systemd"
    sudo systemctl restart prometheus || handle_error "Failed to restart Prometheus"
    wait_for_service "prometheus"
    sudo systemctl enable prometheus || handle_error "Failed to enable Prometheus"

    echo
    green_msg "Done."
    echo
}

# Install Node Exporter
install_node_exporter() {
    echo
    yellow_msg "Installing Node Exporter..."
    echo
    sleep 1

    install_service "node_exporter" "prometheus/node_exporter"

    echo
    green_msg "Done."
    echo
}

configure_prometheus() {
    echo
    yellow_msg "Configuring Prometheus..."
    echo
    sleep 1

    # Remove existing Prometheus configuration
    sudo rm -f /etc/prometheus/prometheus.yml || handle_error "Failed to remove existing Prometheus configuration"

    # Create and edit new Prometheus configuration
    cat <<EOL | sudo tee /etc/prometheus/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
- job_name: node
  static_configs:
  - targets: ['$server_ip:9100']
EOL

    # Reload, restart, enable, and check status of Prometheus
    sudo systemctl daemon-reload || handle_error "Failed to reload systemd"
    sudo systemctl restart prometheus || handle_error "Failed to restart Prometheus"
    wait_for_service "prometheus"
    sudo systemctl enable prometheus || handle_error "Failed to enable Prometheus"

    echo
    green_msg "Done."
    echo
}

install_grafana() {

    echo
    yellow_msg "Installing Grafana..."
    echo
    sleep 1

    # Get the latest Grafana version
    local GRAFANA_VERSION=$(get_latest_release "grafana/grafana")

    # Set Grafana file names based on architecture
    local GRAFANA_FILE="grafana_${GRAFANA_VERSION}_${ARCH}.deb"

    # Wait for dpkg lock
    wait_for_dpkg_lock

    # Download and install Grafana
    wget "https://dl.grafana.com/oss/release/${GRAFANA_FILE}" || handle_error "Failed to download Grafana"
    sudo dpkg -i "$GRAFANA_FILE" || handle_error "Failed to install Grafana"

    # Restart, enable, and check status of Grafana
    sudo systemctl restart grafana-server || handle_error "Failed to restart Grafana"
    wait_for_service "grafana-server"
    sudo systemctl enable grafana-server || handle_error "Failed to enable Grafana"

    echo
    green_msg "Done."
    echo
}

main() {
    # Ensure the script is run as root
    ensure_root

    # Check for required tools
    command -v wget >/dev/null 2>&1 || handle_error "wget is required but it's not installed. Aborting."
    command -v curl >/dev/null 2>&1 || handle_error "curl is required but it's not installed. Aborting."
    command -v systemctl >/dev/null 2>&1 || handle_error "systemctl is required but it's not available. Aborting."

    # Get server IP address
    server_ip=$(get_server_ip)

    # Display confirmation message
    echo
    green_msg "VPS IP Address: $server_ip"
    echo 
    sleep 0.5

    # Get system architecture
    ARCH=$(get_architecture)
    if [ "$ARCH" == "unsupported" ]; then
        echo
        red_msg "Unsupported architecture: $(uname -m)"
        echo
        sleep 1
        exit 1
    else
        echo
        green_msg "VPS ARCH: $ARCH"
        echo
        sleep 0.5
    fi

    prepaire_vps
    sleep 1

    # Install Prometheus
    install_prometheus
    sleep 1

    # Install Node Expoter
    install_node_exporter
    sleep 1

    # Configure Prometheus
    configure_prometheus
    sleep 1

    # Install Grafana
    install_grafana
    sleep 1

    # Deleting & Cleanup Unnecessary Files
    cleanup
    green_msg "Cleanup completed successfully."
    
    echo
    green_msg "Prometheus & Grafana installed and configured successfully."
    yellow_msg "Access the Grafana Dashboard by going to the following web address:"
    yellow_msg "http://$server_ip:3000"
    echo
}

# Call the main function
main
