# Geodata Download and Installation Script

This script automates the download and installation of geodata files required for Xray-core. It fetches the latest versions of the files from a specified repository and installs them into the appropriate directory.

## Features

- Downloads the latest geodata files from [Chocolate4u](https://github.com/Chocolate4U/Iran-v2ray-rules/) repository:
  - `geoip.dat`
  - `geosite.dat`
  - `security-ip.dat`
  - `security.dat`
- Installs files into `/usr/local/share/xray`.


## Requirements

1.
```
apt install wget sudo -y
```

2.
```
sudo -i
```

## Run

```
wget "https://raw.githubusercontent.com/hawshemi/tools/main/geodata/geodata-installer.sh" -O geodata-installer.sh && chmod +x geodata-installer.sh && bash geodata-installer.sh
```
