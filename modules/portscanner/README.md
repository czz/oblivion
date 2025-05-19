## PortScanner

A versatile Oblivion Go module for scanning TCP and optional UDP ports across multiple targets, with ICMP ping support and configurable rate limiting.

## Features

- TCP port scanning with configurable timeout and concurrency  
- Optional UDP scanning  
- Optional ICMP ping RTT measurement  
- Rate limiting of scan requests  
- CIDR expansion for target networks  
- Banner grabbing via raw TCP read or HTTP request  
- JSON output of structured results  
- CLI‑style interface for integration in larger tools  

## Options

| Name          | Default   | Required | Description                             |
|---------------|-----------|----------|-----------------------------------------|
| `TARGETS`     | (empty)   | ✓        | Comma‑separated hosts, IPs, or CIDRs    |
| `PORTS`       | (empty)   | ✓        | Comma or range (e.g. `1-1024`)          |
| `TIMEOUT`     | `1`       |          | Timeout per port in seconds             |
| `THREADS`     | `200`     |          | Number of concurrent scan goroutines    |
| `ENABLE_ICMP` | `false`   |          | Measure ICMP ping RTT                   |
| `ENABLE_UDP`  | `false`   |          | Perform UDP packet scan                 |
| `RATE_LIMIT`  | `100`     |          | Max packets per second per thread       |

## Output

- **Table form** (`[][]string`): each row `[IP, "port/proto", "banner"]`  
- **JSON** (via `Save(path)`): structured array of

```json
[
  {
    "ip": "1.2.3.4",
    "open_ports": {"80": "HTTP 200 OK", "22": ""},
    "ping_rtt": "10ms",
    "protocols": {"80": "tcp", "53": "udp"}
  }
]
```
