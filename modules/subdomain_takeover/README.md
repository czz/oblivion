# Subdomain Takeover

A Oblivion Go module for detecting subdomain takeovers by inspecting CNAME records and HTTP response signatures across multiple targets concurrently.

## Features

- Concurrent DNS lookups with configurable maximum concurrency (default 50)  
- Detection of common takeover fingerprints via DNS NXDOMAIN and HTTP response content  
- Supports multiple target input formats: comma-separated list or file  
- Save results to CSV file or retrieve programmatically  

## Options

| Name     | Default | Required | Description                                           |
|----------|---------|----------|-------------------------------------------------------|
| `DOMAINS`| (empty) | ✓        | Comma-separated domains or path to domains file       |

## Output

- **Table form** (`[][]string`): each row `[domain, cname, service, status, vulnerable]`  
- **CSV** via `Save(path)` method  
