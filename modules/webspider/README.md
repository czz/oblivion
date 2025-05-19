# WebSpider

A Go module for crawling web pages with JavaScript support via a headless browser (Rod). Extracts links (including scripts, images, media) and supports crawl depth, domain restrictions, and optional HTML saving and categorization.

## Features

- **JavaScript rendering** via headless Chrome  
- Configurable **crawl depth**  
- Extracts various asset links: HTML anchors, scripts, images, stylesheets, video/audio sources  
- **Allowed domains** filter for recursion  
- **Optional HTML saving** for each page  
- **Link categorization** (scripts, images, media, etc.)  

## Options

| Name               | Default                                                                                             | Required | Description                                      |
|--------------------|-----------------------------------------------------------------------------------------------------|----------|--------------------------------------------------|
| `TARGETS`          | `[]`                                                                                                | ✓        | List of URLs to crawl (comma-separated or file)  |
| `DEPTH`            | `1`                                                                                                 |          | Maximum crawl depth                              |
| `SAVE_HTML`        | `false`                                                                                             |          | Save full page HTML (`true` or `false`)          |
| `USER_AGENT`       | `Mozilla/...Chrome/91.0.4472.124 Safari/537.36`                                                     |          | Custom User-Agent string                         |
| `ALLOWED_DOMAINS`  | `[]`                                                                                                |          | Domains to restrict recursion                    |
| `INCLUDE_CATEGORIES` | `true`                                                                                            |          | Categorize links by type (scripts, images, media)|

## Output

- **JSON** results when using `Save(filename)`  
- **Table** output (`[][]string`) for inline display  
