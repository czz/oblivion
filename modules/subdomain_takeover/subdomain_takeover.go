package subdomain_takeover

import (
    "fmt"
    "io"
    "net"
    "net/http"
    "os"
    "strings"
    "sync"
    "time"
    "context"

    "github.com/czz/oblivion/utils/help"
    "github.com/czz/oblivion/utils/option"
)

type Service struct {
    Service   string   `json:"service"`
    Cname     []string `json:"cname"`
    Fingerprint []string `json:"fingerprint"`
    Status      string `json:"status"`
    Vulnerable  bool `json:"vulnerable"`
}

type SubdomainTakeover struct {
    optionManager *option.OptionManager
    running bool
    name    string
    author  string
    desc    string
    prompt  string
    help    *help.HelpManager
    results [][]string
}


const (
	MaxConcurrency = 50
	Timeout        = 10 * time.Second
)

var httpClient = &http.Client{
	Timeout: Timeout,
}



func NewSubdomainTakeover() *SubdomainTakeover {
    om := option.NewOptionManager()

    om.Register(option.NewOption("DOMAINS", "", true, "example.com or abc.com,def.com or /pathtofile.txt"))

    hm := help.NewHelpManager()
    hm.Register("subdomain_takeover", "Detect subdomain takeover by matching CNAME and response signature", [][]string{
        {"DOMAINS", "domains.txt", "example.com or abc.com,def.com or /pathtofile.txt"},
    })

    return &SubdomainTakeover{
        optionManager: om,
        name:          "Subdomain Takeover",
        author:        "Luca Cuzzolin",
        desc:          "Detects subdomain takeover via CNAME and HTTP response analysis",
        prompt:        "subdomain_takeover",
        help:          hm,
    }
}

func (s *SubdomainTakeover) Run(ctx context.Context) [][]string {
    // Prepara la lista di domini
    var domains []string
    if val, ok := s.optionManager.Get("DOMAINS"); ok {
        if vs, ok := val.Value.([]string); ok {
            domains = vs
        }
    }

    sem := make(chan struct{}, MaxConcurrency)
    var wg sync.WaitGroup

    resultsCh := make(chan []string)
    var mu sync.Mutex
    var results [][]string

    // Collector
    go func() {
        for rec := range resultsCh {
            mu.Lock()
            results = append(results, rec)
            mu.Unlock()
        }
    }()

    // Lancia i worker
    for _, domain := range domains {
        domain = strings.TrimSpace(domain)
        if domain == "" {
            continue
        }

        select {
        case <-ctx.Done():
            break // esce dal for
        default:
        }

        wg.Add(1)
        sem <- struct{}{} // prende uno slot
        go func(d string) {
            defer wg.Done()
            defer func() { <-sem }()

            // Rispetta la cancellazione
            select {
            case <-ctx.Done():
                return
            default:
            }

            cname, err := net.LookupCNAME(d)
            if err != nil {
                select {
                case resultsCh <- []string{d, "", "", "NXDOMAIN", "false"}:
                case <-ctx.Done():
                }
                return
            }
            cname = strings.TrimSuffix(cname, ".")

            if cname == d {
                return
            }

            // Per ciascun servizio noto
            for _, svc := range services {
                for _, pattern := range svc.Cname {
                    if strings.HasSuffix(cname, pattern) {
                        if !svc.Vulnerable {
                            return
                        }

                        // NXDOMAIN fingerprint
                        for _, fp := range svc.Fingerprint {
                            if fp == "NXDOMAIN" {
                                if _, err := net.LookupHost(cname); err != nil {
                                    select {
                                    case resultsCh <- []string{d, cname, svc.Service, "Vulnerable", "true"}:
                                    case <-ctx.Done():
                                    }
                                }
                                return
                            }
                        }

                        // HTTP fingerprint con contesto
                        client := &http.Client{Timeout: 10 * time.Second}
                        for _, scheme := range []string{"http://", "https://"} {
                            req, err := http.NewRequestWithContext(ctx, "GET", scheme+d, nil)
                            if err != nil {
                                continue
                            }
                            resp, err := client.Do(req)
                            if err != nil {
                                continue
                            }
                            body, _ := io.ReadAll(resp.Body)
                            resp.Body.Close()

                            for _, fp := range svc.Fingerprint {
                                if strings.Contains(string(body), fp) {
                                    select {
                                    case resultsCh <- []string{d, cname, svc.Service, "Vulnerable", "true"}:
                                    case <-ctx.Done():
                                    }
                                    return
                                }
                            }
                        }
                        return
                    }
                }
            }
        }(domain)
    }

    // Chiude resultsCh solo quando tutti i worker hanno finito
    go func() {
        wg.Wait()
        close(resultsCh)
    }()

    // Attende che la cancellazione o il termine dei worker
    select {
    case <-ctx.Done():
        // qui possiamo loggare o pulire
    case <-func() chan struct{} {
        done := make(chan struct{})
        go func() {
            wg.Wait()
            close(done)
        }()
        return done
    }():
    }

    s.results = results
    if len(results) == 0 {
        return nil
    }
    return s.results
}

func (s *SubdomainTakeover) Save(path string) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()
    for _, line := range s.results {
        f.WriteString(strings.Join(line, ",") + "\n")
    }
    return nil
}

func (s *SubdomainTakeover) Set(n, v string) []string {
    om := *s.optionManager
    op, ok := om.Get(n)

    if ok {
        if op.Name == "DOMAINS" {
            var targets []string
            v = strings.TrimSpace(v)
            if fileInfo, err := os.Stat(v); err == nil && !fileInfo.IsDir() {
                content, err := os.ReadFile(v)
                if err != nil {
                    return []string{"DOMAINS", "Error reading file"}
                }
                lines := strings.Split(string(content), "\n")
                for _, line := range lines {
                    line = strings.TrimSpace(line)
                    if line != "" {
                        //if !checkURL(line) { // regez per vedere se è un sotodominio o dominio
                        //    return []string{"DOMAINS", "Error: url must start with http:// or https:// check your file"}
                        //}
                        targets = append(targets, line)
                    }
                }
            } else {
                for _, t := range strings.Split(v, ",") {
                    t = strings.TrimSpace(t)
                    if t != "" {
                        //if !checkURL(t) {
                        //   return []string{"TARGETS", "Error: url must start with http:// or https:// check your file"}
                        //}
                        targets = append(targets, t)
                    }
                }
            }
            op.Set(targets)
            return []string{n, fmt.Sprint(targets)}
        }

        op.Set(v)
        return []string{op.Name, fmt.Sprintf("%v", op.Value)}
    }

    return nil
}


func (s *SubdomainTakeover) Help() [][]string            { help, _ := s.help.Get(s.prompt); return help }
func (s *SubdomainTakeover) Options() []map[string]string { opt := make([]map[string]string, len(s.optionManager.List())); for i, v := range s.optionManager.List() { opt[i] = v.Format() }; return opt }
func (s *SubdomainTakeover) Results() [][]string         { return s.results }
func (s *SubdomainTakeover) Name() string                { return s.name }
func (s *SubdomainTakeover) Author() string              { return s.author }
func (s *SubdomainTakeover) Description() string         { return s.desc }
func (s *SubdomainTakeover) Prompt() string              { return s.prompt }
func (s *SubdomainTakeover) Running() bool               { return s.running }
func (s *SubdomainTakeover) Start() error                { s.running = true; return nil }
func (s *SubdomainTakeover) Stop() error                 { s.running = false; return nil }
