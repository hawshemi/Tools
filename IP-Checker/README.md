# IP-Checker

A simple Go script to retrieve detailed IP information, including ASN and threat data.

## Features

- Fetch ASN details (name, domain, route, type).
- Retrieve threat intelligence (TOR, proxy, datacenter, etc.).
- Scrape IP insights from `browserleaks.com`.
- Display results in a clean table format.

## Prerequisites

- Go installed ([download](https://golang.org/dl/)).
- API key from [ipdata.co](https://ipdata.co/).

## Installation

1. Clone the repo:

    ```bash
    git clone https://github.com/hawshemi/IP-Checker.git
    cd IP-Checker
    ```

2. Install dependencies:

    ```bash
    go mod tidy
    ```

## Usage

1. Export your `IPDATA_API_KEY`:
    Linux:
    ```bash
    export IPDATA_API_KEY=your_api_key
    ```
    Windows:
    ```bash
    $Env:IPDATA_API_KEY = "your_api_key"
    ```
    
2. Run the script:

    ```bash
    go run main.go --ip <IP_ADDRESS>
    ```

### Example:

```bash
go run main.go --ip 1.1.1.1
