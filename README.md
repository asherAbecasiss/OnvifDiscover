# ONVIF Discovery CLI Tool

A Golang-based command-line tool to discover ONVIF devices on the network. It lists all available network interfaces, allows the user to select one, and scans for ONVIF devices broadcasting on the selected interface.

## Features

- Lists all network interfaces on the system.
- Allows the user to select a specific interface for scanning.
- Displays a spinner while scanning for ONVIF devices.
- Outputs discovered ONVIF device streaming URLs in the format: `http://<IP>:<Port>`.

## Prerequisites

- Go 1.18 or later
- A network with ONVIF-compatible devices

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/onvif-discovery-cli.git
   cd onvif-discovery-cli

2. Run the tool.
   ```bash
   ./onvif-discovery
  
3. Example Output
   ```bash
   Available network interfaces:
    [0] enp4s0 (Flags: up|broadcast|multicast)
    [1] enp0 (Flags: up|broadcast|multicast|loopback)
    Select an interface (enter the number): 0
    Selected interface: enp4s0
    Discovered ONVIF streaming URLs:
    http://192.168.1.10:8080/onvif/device_service
    http://192.168.1.11:8080/onvif/device_service
