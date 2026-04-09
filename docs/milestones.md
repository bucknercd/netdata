# Milestones

## Milestone 1: Implement Interface and Route Discovery
- **Objective**: Implement functionality to list network interfaces and display routing table information.
- **Scope**: Implement under `pkg/interfaces/interfaces.go` and `pkg/routes/routes.go` with integration in `cmd/netdata/main.go`.
- **Validation**: Verify that running `netdata -i` lists interfaces and `netdata -r` displays routes.

Status: completed

<!-- FORGE:STATUS START -->

* [x] Implement Interface and Route Discovery

<!-- FORGE:STATUS END -->


## Milestone 2: Implement IP Address Retrieval
- **Depends On**: Milestone 1
- **Objective**: Implement functionality to retrieve and display public and current IP addresses.
- **Scope**: Implement under `pkg/ip/ip.go` with integration in `cmd/netdata/main.go`.
- **Validation**: Verify that running `netdata -p` shows the public IP and `netdata -c` shows current IPs.

Status: completed

<!-- FORGE:STATUS START -->

* [x] Implement IP Address Retrieval

<!-- FORGE:STATUS END -->


## Milestone 3: Implement DNS Information Gathering
- **Depends On**: Milestone 2
- **Objective**: Implement functionality to gather and display DNS resolution information.
- **Scope**: Implement under `pkg/dns/dns.go` with integration in `cmd/netdata/main.go`.
- **Validation**: Verify that running `netdata -d` displays DNS information.

Status: completed

<!-- FORGE:STATUS START -->

* [x] Implement DNS Information Gathering

<!-- FORGE:STATUS END -->


## Milestone 4: Implement Comprehensive Information Display
- **Depends On**: Milestone 3
- **Objective**: Implement the `-a` flag to display all network information in a single command.
- **Scope**: Integrate all functionalities in `cmd/netdata/main.go` to support the `-a` flag.
- **Validation**: Verify that running `netdata -a` displays interfaces, routes, IPs, and DNS info.

Status: completed

<!-- FORGE:STATUS START -->

* [x] Implement Comprehensive Information Display

<!-- FORGE:STATUS END -->


## Milestone 5: Error Handling and Help Information
- **Depends On**: Milestone 4
- **Objective**: Implement robust error handling and display help information for unrecognized commands.
- **Scope**: Enhance `cmd/netdata/main.go` to handle errors and provide user guidance.
- **Validation**: Verify that incorrect commands result in helpful error messages and that `netdata --help` displays usage information.