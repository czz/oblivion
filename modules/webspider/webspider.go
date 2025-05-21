package webspider

import (
    "encoding/json"
    "fmt"
    "os"
    "strconv"
    "strings"
    "net/url"
    "net"
    "context"

    "github.com/czz/oblivion/utils/option"
    "github.com/czz/oblivion/utils/help"
    "github.com/go-rod/rod/lib/proto"
    "github.com/go-rod/rod"
    "github.com/go-rod/rod/lib/launcher"
)

type CrawlResult struct {
    URL      string   `json:"url"`
    Title    string   `json:"title"`
    Links    []string `json:"links"`
    FullHTML string   `json:"html,omitempty"`
}

type WebSpider struct {
    optionManager *option.OptionManager
    name         string
    author       string
    desc         string
    prompt       string
    help         *help.HelpManager
    results      []CrawlResult
    running      bool
    visited      map[string]bool
    table        [][]string
}

func NewWebSpider() *WebSpider {
    om := option.NewOptionManager()

    om.Register(option.NewOption("TARGETS", []string{}, true, "Target URLs to crawl"))
    om.Register(option.NewOption("DEPTH", 1, false, "Crawl depth"))
    om.Register(option.NewOption("SAVE_HTML", false, false, "Save full HTML content"))
    om.Register(option.NewOption("USER_AGENT", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", false, "User-Agent string to use"))
    om.Register(option.NewOption("ALLOWED_DOMAINS", []string{}, false, "List of allowed domains for recursion"))
    om.Register(option.NewOption("INCLUDE_CATEGORIES", true, false, "Categorize extracted links (scripts, images, media, etc.)"))

    helpManager := help.NewHelpManager()
    helpManager.Register("webspider", "Webspider module", [][]string{
        {"TARGETS", "example.com or abc.com,def.com or /pathtofile.txt", "Target URLs to crawl"},
        {"DEPTH", "1", "Crawl depth"},
        {"SAVE_HTML", "true or false", "Timeout in seconds"},
        {"USER_AGENT", "Mozilla/5.0 ...", "User agent string"},
        {"ALLOWED_DOMAINS", "example.com or abc.com,def.com or /pathtofile.txt", "List of allowed domains"},
        {"INCLUDE_CATEGORIES", "true or false", "Classify links into images, scripts, media, etc."},
    })

    return &WebSpider{
        optionManager: om,
        name:    "Web Spider",
        author:  "Luca Cuzzolin",
        desc:    "Web spider with JavaScript support via headless browser",
        prompt:  "webspider",
        help: helpManager,
        visited: make(map[string]bool),
    }
}

func (w *WebSpider) Run(ctx context.Context) [][]string {
    var targets []string
    var saveFullHTML bool
    var depth int
    var userAgent string
    var allowedDomains []string
    var includeCategories = true

    if v, ok := w.optionManager.Get("TARGETS"); ok {
        targets = v.Value.([]string)
    }
    if v, ok := w.optionManager.Get("DEPTH"); ok {
        depth = v.Value.(int)
    }
    if v, ok := w.optionManager.Get("SAVE_HTML"); ok {
        saveFullHTML = v.Value.(bool)
    }
    if v, ok := w.optionManager.Get("USER_AGENT"); ok {
        userAgent = v.Value.(string)
    }
    if v, ok := w.optionManager.Get("ALLOWED_DOMAINS"); ok {
        allowedDomains = v.Value.([]string)
    }
    if v, ok := w.optionManager.Get("INCLUDE_CATEGORIES"); ok {
        includeCategories = v.Value.(bool)
    }

    var rows [][]string
    for _, u := range targets {
        // Exit early if canceled
        select {
        case <-ctx.Done():
            w.visited = make(map[string]bool)
            w.table = rows
            return rows
        default:
        }
        w.recursiveCrawl(ctx, u, saveFullHTML, 0, depth, userAgent, allowedDomains, includeCategories, &rows)
    }
    w.visited = make(map[string]bool)
    w.table = rows
    return rows
}

func (w *WebSpider) recursiveCrawl(
    ctx context.Context,
    urlStr string,
    includeHTML bool,
    currentDepth, maxDepth int,
    userAgent string,
    allowedDomains []string,
    includeCategories bool,
    rows *[][]string,
) {
    // Respect depth and visited
    if currentDepth >= maxDepth || w.visited[urlStr] {
        return
    }
    // Cancel if requested
    select {
    case <-ctx.Done():
        return
    default:
    }

    w.visited[urlStr] = true
    result := w.crawl(urlStr, includeHTML, userAgent, includeCategories)
    w.results = append(w.results, result)

    *rows = append(*rows, []string{"URL", "TITLE", "TOTAL LINKS"})
    *rows = append(*rows, []string{"---", "-----", "-----------"})
    *rows = append(*rows, []string{result.URL, result.Title, fmt.Sprintf("%d links", len(result.Links))})
    *rows = append(*rows, []string{"", "", ""})
    for _, link := range result.Links {
        *rows = append(*rows, []string{link, "", ""})
    }

    for _, link := range result.Links {
        if isAllowed(link, allowedDomains) {
            w.recursiveCrawl(ctx, link, includeHTML, currentDepth+1, maxDepth, userAgent, allowedDomains, includeCategories, rows)
        }
    }
}

func (w *WebSpider) crawl(urlStr string, includeHTML bool, userAgent string, includeCategories bool) CrawlResult {
    if !isResolvable(urlStr) {
        return CrawlResult{URL: urlStr, Title: "Unresolvable host", Links: []string{}}
    }

    u := launcher.New().Headless(true).Leakless(false).Set("user-agent", userAgent).MustLaunch()
    browser := rod.New().ControlURL(u).MustConnect()
    defer browser.MustClose()

    page, err := browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
    if err != nil {
        return CrawlResult{URL: urlStr, Title: "Failed to create page: " + err.Error(), Links: []string{}}
    }

    if err := page.Navigate(urlStr); err != nil {
        return CrawlResult{URL: urlStr, Title: "Navigation failed: " + err.Error(), Links: []string{}}
    }

    page.MustWaitLoad()
    title := page.MustEval(`() => document.title`).String()
    baseURL, _ := url.Parse(page.MustInfo().URL)

    var links []string

    selectors := map[string]string{
        "a[href]":                 "href",
        "script[src]":            "src",
        "img[src]":               "src",
        "link[rel=stylesheet]":   "href",
        "video[src]":             "src",
        "audio[src]":             "src",
        "source[src]":            "src",
    }

    for sel, attr := range selectors {
        elements := page.MustElements(sel)
        for _, el := range elements {
            val, err := el.Attribute(attr)
            if err != nil || val == nil {
                continue
            }
            href := strings.TrimSpace(*val)
            if href == "" || strings.HasPrefix(href, "javascript:") {
                continue
            }
            parsed, err := url.Parse(href)
            if err != nil {
                continue
            }
            resolved := baseURL.ResolveReference(parsed)
            if resolved.Scheme != "http" && resolved.Scheme != "https" {
                continue
            }
            if isResolvable(resolved.String()) {
                links = append(links, resolved.String())
            }
        }
    }

    var html string
    if includeHTML {
        htmlEl := page.MustElement("html")
        html = htmlEl.MustHTML()
    }

    return CrawlResult{
        URL:      urlStr,
        Title:    title,
        Links:    links,
        FullHTML: html,
    }
}

func (w *WebSpider) Save(filename string) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    return encoder.Encode(w.results)
}

func (w *WebSpider) Options() []map[string]string {
    opt := make([]map[string]string, len(w.optionManager.List()))
    for i, v := range w.optionManager.List() {
        opt[i] = v.Format()
    }
    return opt
}

func (w *WebSpider) Set(n string, v string) []string {
    opt, ok := w.optionManager.Get(n)
    if ok {
        switch opt.Name {
        case "TARGETS":
            var targets []string
            v = strings.TrimSpace(v)
            if fileInfo, err := os.Stat(v); err == nil && !fileInfo.IsDir() {
                content, err := os.ReadFile(v)
                if err != nil {
                    return []string{"TARGETS", "Error reading file"}
                }
                lines := strings.Split(string(content), "\n")
                for _, line := range lines {
                    line = strings.TrimSpace(line)
                    if line != "" {
                        if !checkURL(line) {
                            return []string{"TARGETS", "Error: url must start with http:// or https://"}
                        }
                        targets = append(targets, line)
                    }
                }
            } else {
                for _, t := range strings.Split(v, ",") {
                    t = strings.TrimSpace(t)
                    if t != "" {
                        if !checkURL(t) {
                            return []string{"TARGETS", "Error: url must start with http:// or https://"}
                        }
                        targets = append(targets, t)
                    }
                }
            }
            opt.Set(targets)
            return []string{n, fmt.Sprint(targets)}
        case "DEPTH":
            if depth, err := strconv.Atoi(v); err == nil {
                opt.Set(depth)
                return []string{n, v}
            }
            return []string{n, "Invalid depth"}
        case "SAVE_HTML":
            opt.Set((v == "true"))
            return []string{n, v}
        case "USER_AGENT":
            opt.Set(v)
            return []string{n, v}
        case "ALLOWED_DOMAINS":
            var domains []string
            for _, d := range strings.Split(v, ",") {
                d = strings.TrimSpace(d)
                if d != "" {
                    domains = append(domains, d)
                }
            }
            opt.Set(domains)
            return []string{n, fmt.Sprint(domains)}
        case "INCLUDE_CATEGORIES":
            opt.Set((v == "true"))
            return []string{n, v}
        }
    }
    return []string{"Error", "Option not found"}
}

func (w *WebSpider) Help() [][]string {
    help, _ := w.help.Get(w.prompt)
    return help
}

func (w *WebSpider) Results() [][]string {
    return w.table
}

func (w *WebSpider) Name() string       { return w.name }
func (w *WebSpider) Author() string     { return w.author }
func (w *WebSpider) Description() string { return w.desc }
func (w *WebSpider) Prompt() string     { return w.prompt }
func (w *WebSpider) Running() bool      { return w.running }
func (w *WebSpider) Start() error       { w.running = true; return nil }
func (w *WebSpider) Stop() error        { w.running = false; return nil }

func isAllowed(link string, domains []string) bool {
    if len(domains) == 0 {
        return true
    }
    for _, domain := range domains {
        if strings.Contains(link, domain) {
            return true
        }
    }
    return false
}

func checkURL(url string) bool {
    return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func isResolvable(rawURL string) bool {
    parsedURL, err := url.Parse(rawURL)
    if err != nil || parsedURL.Hostname() == "" {
        return false
    }
    _, err = net.LookupHost(parsedURL.Hostname())
    return err == nil
}
