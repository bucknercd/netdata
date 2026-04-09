# Architecture

## Overview

The `netdata` tool will be structured as a single Go binary with modular packages to handle different aspects of network information gathering.

## Modules

- **cmd/netdata/main.go**: Entry point for the CLI tool. Handles command-line arguments and invokes appropriate functions.
- **pkg/interfaces/interfaces.go**: Contains functions to list network interfaces.
- **pkg/routes/routes.go**: Contains functions to retrieve routing table information.
- **pkg/ip/ip.go**: Contains functions to determine public and current IP addresses.
- **pkg/dns/dns.go**: Contains functions to gather DNS resolution information.

## Interaction

- The `main.go` will parse command-line arguments and call functions from the respective packages.
- Each package will interact with the system's network interfaces and configurations to gather data.
- Output will be formatted and printed to the console by the main package.