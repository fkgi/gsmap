# Round-Robin MAP debugger
`roundrobin` is a component of the [gsmap](../README.md) project.
It is an HTTP server for round-robin debugging of MAP (Mobile Application Protocol) messages.
The server utilizes the MAP/TCAP/SCCP+M3UA/SCTP protocol stack and provides HTTP APIs to send, receive, and control MAP dialogs.

<img width="500" alt="Image" src="https://github.com/user-attachments/assets/73299f61-511b-45a7-9cb9-f0baf3b62005" />

Round-Robin can connect to multiple peer node.
But there is no routing function. Round-Robin select destination only by loadshare.

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

# Usage
No commandline options. Configuration parameters are indicated by environment variable.

```sh
roundrobin
```

Commandline example

```sh
export LOCAL_ADDR=10.255.0.11/10.255.1.11:12905
export LOCAL_POINT_CODE=2065
export ROUTING_CONTEXT=1
export NETWORK_INDICATOR=international
export GLOBAL_TITLE=999900010001
export SUBSYSTEM_NUMBER=msc
export PEER_ADDR0=10.255.0.13/10.255.1.13
export GATEWAY_POINT_CODE=2057
export LOCALAPI_ADDR=:18080
export BACKENDAPI_ADDR=localhost:18081
export VERBOSE=yes
roundrobin
```

## Environment Variables
- `LOCAL_ADDR`  
Local SCTP address. Value must have format `IP[/IP]:port` (e.g., 192.168.1.1:14000).
When using multi-homed SCTP, multiple IP addresses are separated by `/`.

- `LOCAL_POINT_CODE`  
Local Point Code digits. Value must decimal digitss.

- `ROUTING_CONTEXT`  
Routing Context (e.g., 101).

- `NETWORK_INDICATOR`  
Network Indicator (`international`/`spare`/`national`/`reserved`).
Default value is `international`.

- `NETWORK_APPEARANCE`  
Network Appearance digits. Not used in default.

- `GLOBAL_TITLE`  
Local Global Title (e.g., 999900000001).

- `SUBSYSTEM_NUMBER`  
Subsystem number (`msc`/`hlr`/`vlr`).
Default value is `msc`.

- `PEER_ADDR0` - `PEER_ADDR9`  
Multiple (max 10) SCTP peer address definition.
When using multi-homed SCTP, multiple IP addresses are separated by `/`.
Round-Robin connect to specified peers and activate ASP.

- `GATEWAY_POINT_CODE`  
Peer Point Code digits. Value must decimal digitss.
This value is always used for DPC of outgoing message (no GTT function).

- `TIMEOUT`  
Message timeout (seconds). Default is `30`.

- `LOCALAPI_ADDR`  
Local listening address and port for receiving HTTP REST request.
Value must have format `[host]:port`.
`host` is hostname or IP address.
IP address is resolved from hostname if hostname is specified.
`port` is port number.

- `BACKENDAPI_ADDR`  
Peer address and port for sending HTTP REST request.
Value must have format `host[:port]`.
`host` is hostname or IP address.
IP address is resolved from hostname if hostname is specified.
`port` is port number.

- `VERBOSE`  
Verbose log mode. Message trace log is logged.

# HTTP API
## Request URI
Request URI for TC-BEGIN of TCAP session has Application Context name and version.
```
http://{hsot}[:port]/mapmsg/v1/{context}/{version}
```
Request URI for TC-CONTINUE/END of TCAP session has session ID that is assigned by server side.
``` 
http://{hsot}[:port]/dialog/{id}
```

## HTTP Body
HTTP request/response Body part of TC-BEGIN/CONTINUE/END has JSON object that indicate MAP message.
In TC-BEGIN, JSON object that indicate CdPA is included. MAP message can be omitted for TCAP handshake.

``` sh
# >>> Tx >>>
POST /mapmsg/v1/{context}/{version}
Content-Type: application/json

{
  "cdpa": {...},
  "InvokeComponentName": {...}
}

# <<< Rx <<<
200 OK
Content-Type: application/json

{
  "cgpa": {...},
  "ReturnresultComponentName": {...}
}
```

## Mapping of TCAP message and HTTP message

| TCAP | HTTP request | JSON body | HTTP response | JSON body |
| - | - | - | - | - |
| TC-BEGIN | POST /mapmsg/v1/{context}/{version} | yes | - | - |
| TC-CONTINUE | POST /dialog/{id}/continue | yes | 201 Created<br>Location: /dialog/{id}<br><br>202 Accepted &dagger; | yes |
| TC-END | POST /dialog/{id}/end &Dagger; | yes | 200 OK | yes |
| TC-ABORT | POST /dialog/{id}/abort &Dagger; | no | 406 NotAcceptable | no |
| discard session | DELETE /dialog/{id} &Dagger; | no | 503 ServiceUnavailable | no |
| AC ver fallback | - | - | 302 Moved Permanently<br>Location: /mapmsg/v1/{context}/{version} | no |

- &dagger; `201 Created` is used if the TC-CONTINUE is response of TC-BEGIN. `202 Accepted` is used if the TC-CONTINUE is response of TC-CONTINUE.
- &Dagger; `204 No content` will be returned if request is TC-END/ABORT or discard.

## Example
Begin by HTTP request
``` sh
POST /mapmsg/v1/{context}/{version}
Content-Type: application/json

{
  "cdpa": {...},
  "InvokeComponentName": {...}
}
```

Continue by HTTP request
``` sh
POST /dialog/{id}/continue
Content-Type: application/json

{
  "Invoke/ReturnresultComponentName": {...}
}
```

Continue by HTTP response
``` sh
201 Created
Content-Type: application/json
Location: /dialog/{id}

{
  "Invoke/ReturnresultComponentName": {...}
}
```

End by HTTP request
``` sh
POST /dialog/{id}/end
Content-Type: application/json

{
  "ReturnresultComponentName": {...}
}
```

End by HTTP response
``` sh
200 OK
Content-Type: application/json

{
  "ReturnresultComponentName": {...}
}
```

Application Context version Fallback by HTTP response.
``` sh
301 Moved Permanently
Location: /mapmsg/v1/{context}/{version}
```

## Get Status
``` sh
GET /mapstate/v1/connection
GET /mapstate/v1/statistics
```

# License
This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.
