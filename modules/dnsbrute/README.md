# DNSBrute

A simple, concurrent Go module for brute-forcing subdomains using a wordlist and DNS resolution. Supports optional numeric suffixes and configurable threading.

## Features

- Brute-force subdomains from a wordlist  
- Optional numeric suffix permutations (e.g. `sub1`, `sub01`, `sub-1`, etc.)  
- Concurrent scanning with a configurable number of goroutines  
- Deduplication and sorted output  
- Save results to file or retrieve programmatically  

## Options

| Name       | Default | Required | Description                                                      |
|------------|---------|----------|------------------------------------------------------------------|
| `DOMAIN`   | (empty) | ✓        | Base domain for brute-forcing (e.g. `example.com`)               |
| `WORDLIST` | (empty) | ✓        | Path to file containing subdomain prefixes, one per line         |
| `THREADS`  | `20`    |          | Number of concurrent goroutines to use                           |
| `SUFFIXES` | `false` |          | Append numeric suffixes (`1`, `01`, `-1`, etc.) to each prefix   |

## Output

- **Table form** (`[][]string`): each row contains one discovered subdomain  
- **Save to file** via the `Save(path)` method: writes one subdomain per line  
