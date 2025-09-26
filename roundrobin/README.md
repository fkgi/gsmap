# roundrobin
`roundrobin` is a component of the [gsmap](../README.md) project.
It is an HTTP server for round-robin debugging of MAP (Mobile Application Protocol) messages.
The server utilizes the MAP/TCAP/SUA/SCCP+M3UA/SCTP protocol stack and provides HTTP APIs to send, receive, and control MAP dialogs.

`roundrobin` run as ASP in USA/M3UA and has onely one connection. MSU routing is roll of peer STP.

<img width="500" alt="Image" src="https://github.com/user-attachments/assets/73299f61-511b-45a7-9cb9-f0baf3b62005" />

## Features
- Control MAP message sending/receiving via HTTP API
- Support for TCAP transaction begin/continue/end
- Compatible with various MAP Application Contexts
- APIs for status and statistics
- SCTP multi-homing support

## Build
```sh
cd roundrobin
go build -o roundrobin
```

## Usage Example
```sh
./roundrobin -l <local_addr> -p <peer_addr> -r <routing_context> -g <global_title> [options]
```

Example:
```sh
./roundrobin -l 10.255.201.18/10.255.202.18:14001 -p 10.255.201.66/10.255.202.66:14001 -r 101 -g 999900000001
```

## Command Line Options
| Option | Description |
|---|---|
| `-l` | Local SCTP address (e.g., 192.168.1.1:14000) |
| `-p` | Peer SCTP address (e.g., 192.168.1.2:14001) |
| `-r` | Routing Context (e.g., 101) |
| `-g` | Global Title (e.g., 999900000001) |
| `-c` | Peer Point Code |
| `-d` | Local Point Code |
| `-s` | Subsystem number (`msc`/`hlr`/`vlr`) |
| `-a` | API server listen address (default: `:8080`) |
| `-b` | Backend API host (default: `localhost:80`) |
| `-t` | Message timeout (seconds) |
| `-v` | Verbose logging |

If Peer Point Code is not defined, `roundrobin` use SUA.
If Peer point Code is defined, `roundrobin` use M3UA.

## HTTP API
### Start Dialog
```
POST /mapmsg/v1/{context}/{version}
Content-Type: application/json

{
  "cdpa": {...},
  "InvokeComponentName": {...}
}
```

### Continue/End Dialog
```
POST /dialog/{id}
DELETE /dialog/{id}
```

### Get Status
```
GET /mapstate/v1/connection
GET /mapstate/v1/statistics
```

## License
MIT
