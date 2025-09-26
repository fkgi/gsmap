# gsmap

`gsmap` is a Mobile Application Protocol (MAP) implementation written in Go.
It provides a full protocol stack for mobile core network signaling, supporting the C, D, and E interfaces of MAP.
This package is designed for use in telecom signaling, testing, and development environments.

[![Go Reference](https://pkg.go.dev/badge/github.com/fkgi/gsmap.svg)](https://pkg.go.dev/github.com/fkgi/gsmap)
[![Go Report Card](https://goreportcard.com/badge/github.com/fkgi/gsmap)](https://goreportcard.com/report/github.com/fkgi/gsmap)

## Features
- MAP protocol implementation (C, D, E interfaces)
- TCAP, SUA, SCCP+M3UA, and SCTP protocol stacks
- SCTP support via Linux kernel module
- Modular and extensible Go codebase
- Suitable for building MAP-based applications, simulators, and test tools

## Architecture
```
+-------------------+
|      MAP          |
+-------------------+
|      TCAP         |
+-------------------+
| SUA | SCCP+M3UA   |
+-------------------+
|      SCTP         |
+-------------------+
|   Linux Kernel    |
+-------------------+
```

- **MAP**: Implements Mobile Application Protocol logic for C, D, E interfaces.
- **TCAP**: Transaction Capabilities Application Part for dialog management.
- **SUA/SCCP+M3UA**: Signaling transport over IP networks.
- **SCTP**: Stream Control Transmission Protocol (Linux only).

## Requirements
- Linux OS (SCTP support requires Linux kernel module)
- Not available on non-Linux platforms

## Restriction
- Only UDT is supported (XUDT/LUDT is not supported) for SCCP
- Only connection less classes are supported for SCCP/SUA
- For SSNM xUA message, only recieveing is supported

## Usage
You can use `gsmap` as a library in your Go project or as a base for building MAP-related tools and servers.
Example (see `roundrobin` directory for a sample application):

```go
import "github.com/fkgi/gsmap"
// Use gsmap APIs to build your MAP application
```


## License
MIT
