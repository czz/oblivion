package dnsbrute

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"strconv"

	"github.com/czz/oblivion/utils/option"
	"github.com/czz/oblivion/utils/help"
)

// DNSBrute is the main struct for the brute-forcing module
type DNSBrute struct {
	optionManager *option.OptionManager
	help          *help.HelpManager
	results       []string
	running       bool
	name          string
	author        string
	desc          string
	prompt        string
}

// NewDNSBrute creates a new instance
func NewDNSBrute() *DNSBrute {
	om := option.NewOptionManager()

	om.Register(option.NewOption("DOMAIN", "", true, "Domain to brute force"))
	om.Register(option.NewOption("WORDLIST", "", true, "Path to wordlist file"))
	om.Register(option.NewOption("THREADS", "20", false, "Number of concurrent goroutines"))
	om.Register(option.NewOption("SUFFIXES", "false", false, "Whether to append numeric suffixes to subdomains"))

	helpManager := help.NewHelpManager()
	helpManager.Register("dnsbrute", "Subdomain brute-force module", [][]string{
		{"DOMAIN", "example.com", "Domain to brute force"},
		{"WORDLIST", "wordlist.txt", "Path to file containing subdomain prefixes"},
		{"THREADS", "20", "Number of threads to use"},
		{"SUFFIXES", "false", "Whether to append numeric suffixes to subdomains"},
	})

	return &DNSBrute{
		optionManager: om,
		help:          helpManager,
		name:          "DNS Brute Force",
		author:        "Luca Cuzzolin",
		desc:          "Brute-force subdomains using a wordlist and DNS resolution, optionally appending numeric suffixes to found hosts.",
		prompt:        "dnsbrute",
	}
}

// Run starts the brute force process
func (b *DNSBrute) Run() [][]string {
	b.results = []string{}
	domainOpt, _ := b.optionManager.Get("DOMAIN")
	wordlistOpt, _ := b.optionManager.Get("WORDLIST")
	threadsOpt, _ := b.optionManager.Get("THREADS")
	suffixesOpt, _ := b.optionManager.Get("SUFFIXES")

	domain := domainOpt.Value.(string)
	wordlistPath := wordlistOpt.Value.(string)
	threadCount, err := strconv.Atoi(threadsOpt.Value.(string))
	if err != nil || threadCount <= 0 {
		threadCount = 20 // default fallback
	}

	suffixes, _ := strconv.ParseBool(suffixesOpt.Value.(string))

	if domain == "" || wordlistPath == "" {
		return [][]string{{"Error: DOMAIN or WORDLIST not set"}}
	}

	words, err := loadWordlist(wordlistPath)
	if err != nil {
		return [][]string{{"Error reading wordlist"}}
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, threadCount)
	resCh := make(chan string, len(words))

	for _, prefix := range words {
		wg.Add(1)
		sem <- struct{}{}
		go func(sub string) {
			defer wg.Done()
			defer func() { <-sem }()

			fqdn := fmt.Sprintf("%s.%s", sub, domain)
			if resolveDomain(fqdn) {
				// Add the base subdomain to the results
				resCh <- fqdn

				// If suffixes are enabled, append numeric suffixes
				if suffixes {
					for _, suffix := range generateNumberSuffixes() {
						fqdnWithSuffix := fmt.Sprintf("%s.%s", sub+suffix, domain)
						if resolveDomain(fqdnWithSuffix) {
							resCh <- fqdnWithSuffix
						}
					}
				}
			}
		}(prefix)
	}

	wg.Wait()
	close(resCh)

	// Collect unique results
	unique := make(map[string]struct{})
	for r := range resCh {
		unique[r] = struct{}{}
	}

	// Sort the results and convert them to a list
	for k := range unique {
		b.results = append(b.results, k)
	}
	sort.Strings(b.results)

	var out [][]string
	for _, r := range b.results {
		out = append(out, []string{r})
	}
	return out
}

// Helper: generate numeric suffixes to append
func generateNumberSuffixes() []string {
	var suffixes []string
	for i := 1; i <= 10; i++ {
		// Add number as 1, 01, 001
		suffixes = append(suffixes, fmt.Sprintf("%d", i), fmt.Sprintf("%02d", i), fmt.Sprintf("%03d", i))

		// Add negative suffixes -1, -01, -001
		suffixes = append(suffixes, fmt.Sprintf("-%d", i), fmt.Sprintf("-0%d", i), fmt.Sprintf("-00%d", i))
	}
	return suffixes
}

// Helper: resolve a domain using DNS
func resolveDomain(name string) bool {
	_, err := net.LookupHost(name)
	return err == nil
}

// Helper: load wordlist from file
func loadWordlist(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	var words []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			words = append(words, line)
		}
	}
	return words, nil
}

// Save the results to a file
func (b *DNSBrute) Save(path string) error {
	if len(b.results) == 0 {
		return fmt.Errorf("no results to save")
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, r := range b.results {
		_, err := f.WriteString(r + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

// Help returns the help text for the module
func (b *DNSBrute) Help() [][]string {
	help, _ := b.help.Get(b.prompt)
	return help
}

// Options returns the available options for the module
func (b *DNSBrute) Options() []map[string]string {
	opts := b.optionManager.List()
	out := make([]map[string]string, len(opts))
	for i, o := range opts {
		out[i] = o.Format()
	}
	return out
}

// Results returns the results of the brute force operation
func (b *DNSBrute) Results() [][]string {
	var res [][]string
	for _, r := range b.results {
		res = append(res, []string{r})
	}
	return res
}

// Set allows dynamic setting of options
func (b *DNSBrute) Set(name, value string) []string {
	opt, _ := b.optionManager.Get(name)
	if opt != nil {
		opt.Set(value)
		return []string{opt.Name, fmt.Sprintf("%v", opt.Value)}
	}
	return []string{"Error", "Option not found"}
}

// Metadata
func (b *DNSBrute) Name() string     { return b.name }
func (b *DNSBrute) Author() string   { return b.author }
func (b *DNSBrute) Description() string { return b.desc }
func (b *DNSBrute) Prompt() string   { return b.prompt }
func (b *DNSBrute) Running() bool    { return b.running }
func (b *DNSBrute) Start() error     { b.running = true; return nil }
func (b *DNSBrute) Stop() error      { b.running = false; return nil }
