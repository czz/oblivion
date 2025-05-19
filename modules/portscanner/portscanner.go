package portscanner

import (
    "encoding/json"
    "fmt"
    "net"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"
    "net/http"
    "sort"

    "github.com/go-ping/ping"
    "github.com/czz/oblivion/utils/option"
    "github.com/czz/oblivion/utils/help"
)

type JsonScanResult struct {
    IP       string         `json:"ip"`
    Open     map[int]string `json:"open_ports"`
    PingRTT  time.Duration  `json:"ping_rtt,omitempty"`
    Protocol map[int]string `json:"protocols,omitempty"`
}

type PortScanner struct {
    optionManager *option.OptionManager
    running       bool
    name          string
    author      string
    desc        string
    prompt      string
    help        *help.HelpManager
    results     [][]string
    jsonResults []JsonScanResult
    Client      *http.Client
}

func NewPortScanner() *PortScanner {
    om := option.NewOptionManager()

    om.Register(option.NewOption("TARGETS", []string{}, true, "Targets"))
    om.Register(option.NewOption("PORTS", []int{}, true, "Ports to scan"))
    om.Register(option.NewOption("TIMEOUT", 1, false, "Timeout in seconds"))
    om.Register(option.NewOption("THREADS", 200, false, "Number of threads"))
    om.Register(option.NewOption("ENABLE_ICMP", false, false, "Enable ICMP"))
    om.Register(option.NewOption("ENABLE_UDP", false, false, "Enable UDP scan"))
    om.Register(option.NewOption("RATE_LIMIT", 100, false, "Rate Limit"))

    var netTransport = &http.Transport{
        Dial: (&net.Dialer{
        Timeout: 5 * time.Second,
      }).Dial,
      TLSHandshakeTimeout: 3 * time.Second,
    }
    var client = &http.Client{
      Timeout: time.Second * 3,
      Transport: netTransport,
    }

    helpManager := help.NewHelpManager()
    helpManager.Register("portscanner", "Portscanner module",[][]string{
  		{"TARGETS", "example.com or abc.com,def.com or /pathtofile.txt", "Targets to scan"},
  		{"PORTS", "80 or 22,80 or 1-10000", "Ports to scan"},
  		{"TIMEOUT", "1", "Timeout in seconds"},
      {"THREADS", "200", "Number of threads"},
      {"ENABLE_ICMP", "true or false", "Enable ICMP"},
      {"ENABLE_UDP", "true or false", "Enable UDP scan"},
      {"RATE_LIMIT", "100", "Rate Limit"},
  	})

    return &PortScanner{
        optionManager:    om,
        name:   "Port Scanner",
        author: "Luca Cuzzolin",
        desc:   "Simple port scanner with UDP support",
        prompt: "portscanner",
        help:  helpManager,
        Client: client,
    }
}

func (p *PortScanner) Run() [][]string {
    var targets []string
    var threads int

    targets_opt, _ := p.optionManager.Get("TARGETS")
    if val, ok := targets_opt.Value.([]string); ok {
      targets = val
    }

    threads_opt, _ := p.optionManager.Get("THREADS")
    if val, ok := threads_opt.Value.(int); ok {
      threads = val
    }

    var wg sync.WaitGroup
    tasks := make(chan string, len(targets))
    results := make(chan JsonScanResult, len(targets))

    for _, target := range targets {
        ips := expandCIDR(target)
        for _, ip := range ips {
            tasks <- ip
        }
    }
    close(tasks)

    for i := 0; i < threads; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for target := range tasks {
                result := p.scanTarget(target)
                results <- result
            }
        }()
    }

    wg.Wait()
    close(results)

    var tableData [][]string
    for res := range results {
        p.jsonResults = append(p.jsonResults, res)
        for port, banner := range res.Open {
            proto := res.Protocol[port]
            row := []string{res.IP, fmt.Sprintf("%d/%s", port, proto), banner}
            tableData = append(tableData, row)
        }
    }
    p.results = tableData
    return tableData
}

func (p *PortScanner) scanTarget(ip string) JsonScanResult {
    var enICMP, enUDP bool
    var timeout, threadCount int
    var ports []int

    if val, ok := p.optionManager.Get("ENABLE_ICMP"); ok {
        enICMP, _ = val.Value.(bool)
    }
    if val, ok := p.optionManager.Get("ENABLE_UDP"); ok {
        enUDP, _ = val.Value.(bool)
    }
    if val, ok := p.optionManager.Get("TIMEOUT"); ok {
        timeout, _ = val.Value.(int)
    }
    if val, ok := p.optionManager.Get("THREADS"); ok {
        threadCount, _ = val.Value.(int)
    }
    if val, ok := p.optionManager.Get("PORTS"); ok {
        ports, _ = val.Value.([]int)
    }

    result := JsonScanResult{
        IP:       ip,
        Open:     make(map[int]string),
        Protocol: make(map[int]string),
    }

    if enICMP {
        go func() {
            pinger, err := ping.NewPinger(ip)
            if err == nil {
                pinger.Count = 1
                pinger.Timeout = time.Duration(timeout) * time.Second
                if err := pinger.Run(); err == nil {
                    stats := pinger.Statistics()
                    result.PingRTT = stats.AvgRtt
                }
            }
        }()
    }

    type scanResult struct {
        port   int
        banner string
        proto  string
    }

    tasks := make(chan int, len(ports))
    output := make(chan scanResult, len(ports)*2)

    for _, port := range ports {
        tasks <- port
    }
    close(tasks)

    var wg sync.WaitGroup

    rateLimit := 0
    if val, ok := p.optionManager.Get("RATE_LIMIT"); ok {
        rateLimit, _ = val.Value.(int)
    }
    var delay time.Duration
    if rateLimit > 0 {
        delay = time.Second / time.Duration(rateLimit)
    }

    for i := 0; i < threadCount; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for port := range tasks {
                if delay > 0 {
                    time.Sleep(delay)
                }

                address := net.JoinHostPort(ip, strconv.Itoa(port))

                conn, err := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Second)
                if err == nil {
                    banner := getBanner(conn)
                    conn.Close()
                    if banner == "" {
                        banner = httpGrabBanner(p.Client, ip, port)
                    }
                    output <- scanResult{port, banner, "tcp"}
                }

                if enUDP {
                    udpAddr, err := net.ResolveUDPAddr("udp", address)
                    if err != nil {
                        continue
                    }
                    conn, err := net.DialUDP("udp", nil, udpAddr)
                    if err != nil {
                        continue
                    }
                    conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
                    _, err = conn.Write([]byte{0x0})
                    if err != nil {
                        conn.Close()
                        continue
                    }
                    buf := make([]byte, 1024)
                    _, _, err = conn.ReadFromUDP(buf)
                    conn.Close()

                    if err != nil {
                        output <- scanResult{port, "[no response]", "udp"}
                    } else {
                        output <- scanResult{port, "[responded]", "udp"}
                    }
                }
            }
        }()
    }

    wg.Wait()
    close(output)

    for res := range output {
        result.Open[res.port] = res.banner
        result.Protocol[res.port] = res.proto
    }

    return result
}

func getBanner(conn net.Conn) string {
    conn.SetReadDeadline(time.Now().Add(2 * time.Second))
    buf := make([]byte, 1024)
    n, err := conn.Read(buf)
    if err != nil {
        return ""
    }
    return strings.TrimSpace(string(buf[:n]))
}

func httpGrabBanner(client *http.Client, ip string, port int) string {
    address := fmt.Sprintf("http://%s:%d", ip, port)
    resp, err := client.Get(address)
    if err != nil {
        return ""
    }
    defer resp.Body.Close()
    return fmt.Sprintf("HTTP %d %s", resp.StatusCode, resp.Status)
}

func expandCIDR(input string) []string {
    var results []string
    if strings.Contains(input, "/") {
        ip, ipnet, err := net.ParseCIDR(input)
        if err != nil {
            return []string{input}
        }
        ip = ip.To4()
        if ip == nil {
            return []string{input}
        }
        for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); ip = nextIP(ip) {
            ipStr := ip.String()
            if ipStr != "127.0.0.1" && !strings.HasPrefix(ipStr, "169.254") {
                results = append(results, ipStr)
            }
        }
        return results
    }
    return []string{input}
}

func nextIP(ip net.IP) net.IP {
    ip = ip.To4()
    for j := len(ip) - 1; j >= 0; j-- {
        ip[j]++
        if ip[j] != 0 {
            break
        }
    }
    return ip
}

func (p *PortScanner) saveJSON(filename string) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    return encoder.Encode(p.jsonResults)
}

func parsePorts(portStr string) []int {
    var ports []int
    seen := make(map[int]bool)
    for _, part := range strings.Split(portStr, ",") {
        part = strings.TrimSpace(part)
        if strings.Contains(part, "-") {
            bounds := strings.Split(part, "-")
            if len(bounds) != 2 {
                continue
            }
            start, _ := strconv.Atoi(bounds[0])
            end, _ := strconv.Atoi(bounds[1])
            for i := start; i <= end; i++ {
                if !seen[i] {
                    ports = append(ports, i)
                    seen[i] = true
                }
            }
        } else {
            p, err := strconv.Atoi(part)
            if err == nil && !seen[p] {
                ports = append(ports, p)
                seen[p] = true
            }
        }
    }
    return ports
}

func compressPorts(nums []int) string {
    if len(nums) == 0 {
        return ""
    }
    sort.Ints(nums)
    var result []string
    start := nums[0]
    end := nums[0]
    for i := 1; i < len(nums); i++ {
        if nums[i] == end+1 {
            end = nums[i]
        } else {
            if start == end {
                result = append(result, fmt.Sprintf("%d", start))
            } else {
                result = append(result, fmt.Sprintf("%d-%d", start, end))
            }
            start = nums[i]
            end = nums[i]
        }
    }
    if start == end {
        result = append(result, fmt.Sprintf("%d", start))
    } else {
        result = append(result, fmt.Sprintf("%d-%d", start, end))
    }
    return strings.Join(result, ",")
}

func (p *PortScanner) Save(filename string) error {
    return p.saveJSON(filename)
}

func (p *PortScanner) Options() []map[string]string {
    res := make([]map[string]string, 0, len(p.optionManager.List()))
    for _, opt := range p.optionManager.List() {
        switch opt.Name {
        case "PORTS":
            if val, ok := opt.Value.([]int); ok {
                ports := map[string]string{"name": opt.Name, "value": compressPorts(val), "required": "true", "description": opt.Description}
                res = append(res, ports)
            }
        default:
            res = append(res, opt.Format())
        }
    }
    return res
}

func (p *PortScanner) Set(n string, v string) []string {

    opt, ok := p.optionManager.Get(n)
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
                            targets = append(targets, line)
                        }
                    }
                } else {
                    for _, t := range strings.Split(v, ",") {
                        t = strings.TrimSpace(t)
                        if t != "" {
                            targets = append(targets, t)
                        }
                    }
                }
                opt.Set(targets)
                return []string{opt.Name, fmt.Sprint(targets)}
            case "PORTS":
                ports := parsePorts(v)
                opt.Set(ports)
                return []string{opt.Name, fmt.Sprint(v)}
            case "TIMEOUT", "RATE_LIMIT", "THREADS":
                if intVal, err := strconv.Atoi(v); err == nil {
                    opt.Set(intVal)
                    return []string{opt.Name, fmt.Sprint(intVal)}
                } else {
                    return []string{opt.Name, "Invalid integer value"}
                }
            case "ENABLE_UDP", "ENABLE_ICMP":
                if v == "true" {
                    opt.Set(true)
                } else {
                    opt.Set(false)
                }
                return []string{opt.Name, fmt.Sprint(opt.Value)}
            default:
                opt.Set(v)
                return []string{opt.Name, fmt.Sprint(v)}
            }
        }
    return []string{"Error", "Option not found"}
}

func (p *PortScanner) Help() [][]string {
  help, _ := p.help.Get(p.prompt)
  return help

}

func (p *PortScanner) Results() [][]string {
    return p.results
}

func (s *PortScanner) Name() string       { return s.name }
func (s *PortScanner) Author() string     { return s.author }
func (s *PortScanner) Description() string { return s.desc }
func (s *PortScanner) Prompt() string     { return s.prompt }
func (s *PortScanner) Running() bool      { return s.running }
func (s *PortScanner) Start() error       { s.running = true; return nil }
func (s *PortScanner) Stop() error        { s.running = false; return nil }
