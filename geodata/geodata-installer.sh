#!/bin/bash

download_file() {
    local url="$1"
    local output_file="$2"

    echo "Downloading ${url}..."
    if ! wget -q -O "${output_file}" "${url}"; then
        echo "error: Download failed for ${url}. Please check your network or try again."
        exit 1
    fi
    echo "Downloaded ${output_file} successfully."
}

install_file() {
    local source_file="$1"
    local destination="$2"

    echo "Installing ${source_file} to ${destination}..."
    if ! install -m 644 "${source_file}" "${destination}"; then
        echo "error: Installation failed for ${source_file}. Please check permissions."
        exit 1
    fi
    echo "Installed ${source_file} successfully."
}

download_and_install_geodata() {
    local download_links=(
        "https://github.com/Chocolate4U/Iran-v2ray-rules/releases/latest/download/geoip.dat"
        "https://github.com/Chocolate4U/Iran-v2ray-rules/releases/latest/download/geosite.dat"
        "https://github.com/Chocolate4U/Iran-v2ray-rules/releases/latest/download/security-ip.dat"
        "https://github.com/Chocolate4U/Iran-v2ray-rules/releases/latest/download/security.dat"
    )
    
    local file_names=(
        "geoip_ir.dat"
        "geosite_ir.dat"
        "security-ip_ir.dat"
        "security_ir.dat"
    )

    local dir_tmp
    dir_tmp="$(mktemp -d)"
    local install_dir="/usr/local/share/xray"

    echo "Creating installation directory at ${install_dir}..."
    # Create installation directory if it doesn't exist
    install -d "${install_dir}"
    echo "Installation directory ready."

    for i in "${!download_links[@]}"; do
        local download_link="${download_links[$i]}"
        local file_name="${file_names[$i]}"
        local output_file="${dir_tmp}/${file_name}"

        # Download the file
        download_file "${download_link}" "${output_file}"

        # Install the file
        install_file "${output_file}" "${install_dir}/${file_name}"
    done

    rm -r "${dir_tmp}"
    echo "Temporary files cleaned up."
    echo "All files installed successfully in ${install_dir}."
}

# Call the function to download and install geodata
download_and_install_geodata
