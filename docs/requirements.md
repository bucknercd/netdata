# Requirements

## Net Data Tool

- **Command**: The tool will be invoked via the command line using `netdata`.
- **Flags**: 
  - `-i` or `--interfaces`: List all network interfaces.
  - `-r` or `--routes`: Display routing table.
  - `-p` or `--public-ip`: Show public IP address.
  - `-c` or `--current-ip`: Show current IP address of each interface.
  - `-d` or `--dns`: Display DNS resolution information.
  - `-a` or `--all`: Display all available network information.
- **Inputs**: No additional input files required; all data is gathered from the system.
- **Outputs**: Information will be printed to the console in a human-readable format.
- **Error Cases**:
  - If a flag is not recognized, the tool will output an error message and display help information.
  - If network information cannot be retrieved, the tool will output an appropriate error message.

## Constraints

- The tool must be implemented in Golang.
- Minimize the use of external dependencies, preferring the Go standard library where possible.