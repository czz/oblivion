# Subdomains Search

A lightweight Oblivion Go module for discovering subdomains of a given domain by querying multiple free online sources in parallel.

## Features

* Fetch subdomains from:

  * CertSpotter
  * crt.sh
  * urlscan.io
  * AlienVault OTX
  * Anubis
  * HackerTarget
* Parallel queries across configurable sources
* Normalizes, deduplicates, and sorts results


## Options

| Name          | Default                        | Required | Description                           |
| ------------- | ------------------------------ | -------- | ------------------------------------- |
| `SOURCES_URI` | Preconfigured list (immutable) | ✓        | API endpoints to query for subdomains |
| `DOMAIN`      | (empty)                        | ✓        | Domain to search for subdomains       |
