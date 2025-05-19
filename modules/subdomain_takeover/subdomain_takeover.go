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
        author:        "Your Name",
        desc:          "Detects subdomain takeover via CNAME and HTTP response analysis",
        prompt:        "subdomain_takeover",
        help:          hm,
    }
}

func (s *SubdomainTakeover) Run() [][]string {
    var domains []string
    dopt, _ := s.optionManager.Get("DOMAINS")
/*    if path, ok := dopt.Value.(string); ok && path != "" {
        file, err := os.Open(path)
        if err != nil {
            return [][]string{{"Error opening domains file:",fmt.Sprintf("%v", err)}}
        }
        defer file.Close()
        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            domains = append(domains, scanner.Text())
        }
    }
*/
if val, ok := dopt.Value.([]string); ok {
  domains = val
}

//fmt.Println("DOMAIN",domains)
    sem := make(chan struct{}, MaxConcurrency)
  	var wg sync.WaitGroup

  	resultsCh := make(chan []string)
  	var results [][]string
  	var resultsMu sync.Mutex

  	go func() {
    		for rec := range resultsCh {
      			resultsMu.Lock()
  		    	results = append(results, rec)
    			  resultsMu.Unlock()
    		}
    }()

    for _, domain := range domains {
        domain = strings.TrimSpace(domain)
        if domain == "" {
            continue
        }
        wg.Add(1)
        sem <- struct{}{} // acquisisce uno slot
        go func(d string) {
            defer wg.Done()
            defer func() { <-sem }() // rilascia lo slot

            // Lookup CNAME
            cname, err := net.LookupCNAME(d)
            if err != nil {
  //            fmt.Println("CNAME","ERROR")
                // Dominio inesistente (NXDOMAIN) o altro errore DNS
                resultsCh <- []string{d, "", "", "NXDOMAIN", "false"}
                return
            }
            cname = strings.TrimSuffix(cname, ".")

            if cname == d {
                // Nessun record CNAME (record A diretto o nessun record)
//fmt.Println("A",cname)
                return
            }

            // Controlla contro i servizi vulnerabili noti
            for _, svc := range services {
                for _, pattern := range svc.Cname {
                    if strings.HasSuffix(cname, pattern) {
                        // Trovato un servizio che corrisponde al pattern
                        if !svc.Vulnerable {
                            // Servizio noto come sicuro
                            return
                        }
                        // Verifica il fingerprint NXDOMAIN
                        for _, fp := range svc.Fingerprint {
                            if fp == "NXDOMAIN" {
                                _, err := net.LookupHost(cname)
                                if err != nil {
  //                                  fmt.Println("LOOKUP HOST KO")
                                    // CNAME target non trovato -> vulnerabile
                                    resultsCh <- []string{d, cname, svc.Service, "Vulnerable", "true"}
                                    //fmt.Printf("[!] Takeover trovato: %s (%s)\n", d, svc.Service)
                                    //resultsCh <- []string{domain, service, fullURL}
                                }
  //                              fmt.Println("NXDOMAIN")
                                return
                            }
                        }
                        // Esegui richieste HTTP/HTTPS per controllare i fingerprint
                        client := &http.Client{Timeout: 10 * time.Second}
                        for _, scheme := range []string{"http://", "https://"} {
                            req, err := http.NewRequest("GET", scheme+d, nil)
                            if err != nil {
                                continue
                            }
                            req.Header.Set("User-Agent", "subdomain-takeover-scanner")
                            resp, err := client.Do(req)
                            if err != nil {
                                continue
                            }
                            content, err := io.ReadAll(resp.Body)
                            resp.Body.Close()
                            if err != nil {
                                continue
                            }
                            body := string(content)
//                            fmt.Println("BODY")
                            for _, fp := range svc.Fingerprint {
                                if strings.Contains(body, fp) {
                                    resultsCh <- []string{d, cname, svc.Service, "Vulnerable", "true"}
//                                    fmt.Println("VULN")
                                    return
                                }
                            }
                        }
                        // Se non trovato fingerprint
//                        fmt.Println("NO FINGER",domains)
                        return
                    }
                }
            }
            // Nessun servizio che corrisponde per questo CNAME
      	}(domain)
    }

    wg.Wait()
    close(resultsCh)

    if len(results) == 0 { return nil }
    s.results = results
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
