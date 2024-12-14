# DNS Tester

This Go script tests the latency (ping) and DNS resolution time for a list of popular DNS providers (IPv4 and IPv6). It performs the following tasks:
- Ping each provider to check network latency.
- Resolve domains via DNS for each provider and calculate the average resolution time.

## Features
- Tests both IPv4 and IPv6 DNS providers.
- Measures average ping and DNS resolution time for predefined domains.
- Supports Cloudflare, Google, Quad9, Adguard, and other well-known DNS providers.

## Requirements
- Golang 1.18+ (install by `sudo apt install golang` on Ubuntu).
- Linux/macOS/Windows environment (the script uses `ping` and `dig` commands).
- IPv6 support (optional, the script will skip IPv6 tests if not enabled).

## Installation

1. Clone the repository or download the script:
   ```bash
   git clone https://github.com/hawshemi/dns-tester.git
   cd dns-tester
   
2. Install Golang (if not already installed):
   ```bash
   sudo apt install golang
   
3. Run the script:
   ```bash
   go run main.go
