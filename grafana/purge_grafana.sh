#!/bin/bash

clear

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

# Function to display an error message
display_error() {
    echo
    red_msg "Invalid response. Please type 'YES' or 'NO' to confirm or cancel the purge."
    echo
}

# Function to check if the script is run as root
check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        echo
        red_msg "This script must be run as root."
        echo
        exit 1
    fi
}

# Stop and disable a service
stop_and_disable_service() {
    local service_name=$1
    if sudo systemctl is-active --quiet "$service_name"; then
        sudo systemctl stop "$service_name"
    fi
    sudo systemctl disable "$service_name"
}

# Confirmation prompt with three attempts
confirm_deletion() {
    for attempt in {1..3}; do
        echo
        yellow_msg "You are initiating the deletion of the Grafana Monitoring package."
        yellow_msg "Are you sure you would like to proceed with the deletion?"
        echo
        yellow_msg "To cancel the deletion, type 'NO' and to confirm deletion, type 'YES'."
        echo
        read -p "Please confirm your answer: " answer
        echo

        # Convert the user's response to uppercase for case-insensitive comparison
        answer_uppercase=$(echo "$answer" | tr '[:lower:]' '[:upper:]')

        # Check user's response
        case "$answer_uppercase" in
            YES)
                echo
                yellow_msg "Purging and cleaning related components..."
                echo
                return 0
                ;;
            NO)
                echo
                red_msg "Cleanup aborted. No changes were made."
                echo
                exit 0
                ;;
            *)
                display_error
                ;;
        esac

        if [ "$attempt" -eq 3 ]; then
            red_msg "Too many incorrect attempts. Exiting."
            exit 1
        fi
    done
}

# Main function to perform the deletion
perform_deletion() {
    # Update package list
    sudo apt update -qq
    sleep 0.5

    # Stop and disable services
    services=("prometheus" "node_exporter" "grafana-server")
    for service in "${services[@]}"; do
        stop_and_disable_service "$service"
        sleep 0.5
    done

    # Remove packages
    sudo apt remove --purge -yq prometheus node_exporter grafana-enterprise grafana
    sleep 0.5

    sudo apt autopurge -yq
    sleep 0.5

    # Remove configuration files and directories
    sudo rm -rf /etc/prometheus
    sleep 0.3
    sudo rm -f /etc/systemd/system/prometheus.service
    sleep 0.3
    sudo rm -f /etc/node_exporter/node_exporter
    sleep 0.3
    sudo rm -f /etc/systemd/system/node_exporter.service
    sleep 0.3
    sudo rm -f /etc/prometheus/prometheus.yml
    sleep 0.3
    sudo rm -f /usr/sbin/grafana-server
    sleep 0.3
    sudo rm -f /etc/apt/sources.list.d/grafana.list
    sleep 0.3

    # Reload systemd
    sudo systemctl daemon-reload
    sleep 0.5

    echo
    green_msg "Deletion successfully completed."
    green_msg "All components and files related to Grafana Monitoring have been removed."
    echo
    sleep 0.5
}

# Start script execution
check_root
confirm_deletion
perform_deletion
