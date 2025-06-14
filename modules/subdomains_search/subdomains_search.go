package subdomains_search

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"context"

	"github.com/czz/oblivion/utils/option"
	"github.com/czz/oblivion/utils/help"
)

const (
	NO_SOURCES     = "No Sources Found"  // Constant for no sources found
	ERROR_FETCHING = "Error Fetching"    // Constant for error while fetching
)

// Default API sources for subdomain enumeration
var defaultSources = []string{
	"https://api.certspotter.com/v1/issuances?domain=%s&expand=dns_names&expand=issuer",
	"https://crt.sh/?q=%s&output=json",
	"https://urlscan.io/api/v1/search/?q=domain:%s",
	"https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns",
	"https://jldc.me/anubis/subdomains/%s",
  "https://api.hackertarget.com/hostsearch/?q=%s",
}

// SubdomainsSearch holds state and configuration for a subdomain scan
type SubdomainsSearch struct {
	optionManager *option.OptionManager // Manager for module options
	running       bool                  // Indicates if the module is currently running
	name          string                // Name of the module
	author        string                // Author of the module
	desc          string                // Description of the module
	prompt        string                // CLI prompt name
	help          *help.HelpManager     // Help manager for CLI usage
	results       []string              // Stores the found subdomains
}

// fetchSubdomains queries a given source URL and extracts subdomains from the response
func (s *SubdomainsSearch) fetchSubdomains(url string, domain string) ([]string, error) {
	formattedURL := fmt.Sprintf(url, domain)
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(formattedURL)
	if err != nil {
		log.Printf("Error fetching from %s: %v", formattedURL, err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading body from %s: %v", formattedURL, err)
		return nil, err
	}

	var subdomains []string

	switch {
	case strings.Contains(url, "crt.sh"):
		var jsonData []map[string]interface{}
		if err := json.Unmarshal(body, &jsonData); err == nil {
			for _, entry := range jsonData {
				if names, ok := entry["name_value"].(string); ok {
					for _, name := range strings.Split(names, "\n") {
						subdomains = append(subdomains, normalizeDomain(name))
					}
				}
			}
		} else {
			log.Printf("JSON parsing error from crt.sh: %v", err)
		}

	case strings.Contains(url, "urlscan.io"):
		var result struct {
			Results []struct {
				Page struct {
					Domain string `json:"domain"`
				} `json:"page"`
			} `json:"results"`
		}
		if err := json.Unmarshal(body, &result); err == nil {
			for _, r := range result.Results {
				subdomains = append(subdomains, normalizeDomain(r.Page.Domain))
			}
		} else {
			log.Printf("JSON parsing error from urlscan: %v", err)
		}

	case strings.Contains(url, "alienvault"):
		var result struct {
			PassiveDNS []struct {
				Hostname string `json:"hostname"`
			} `json:"passive_dns"`
		}
		if err := json.Unmarshal(body, &result); err == nil {
			for _, r := range result.PassiveDNS {
				subdomains = append(subdomains, normalizeDomain(r.Hostname))
			}
		} else {
			log.Printf("JSON parsing error from alienvault: %v", err)
		}

	case strings.Contains(url, "jldc.me"):
		var result []string
		if err := json.Unmarshal(body, &result); err == nil {
			for _, name := range result {
				subdomains = append(subdomains, normalizeDomain(name))
			}
		} else {
			log.Printf("JSON parsing error from jldc: %v", err)
		}

	case strings.Contains(url, "certspotter"):
		var jsonData []map[string]interface{}
		if err := json.Unmarshal(body, &jsonData); err == nil {
			for _, entry := range jsonData {
				if names, ok := entry["dns_names"].([]interface{}); ok {
					for _, name := range names {
						if str, ok := name.(string); ok {
							subdomains = append(subdomains, normalizeDomain(str))
						}
					}
				}
			}
		} else {
			log.Printf("JSON parsing error from certspotter: %v", err)
		}
	}

	return subdomains, nil
}

// normalizeDomain cleans up a domain name by lowercasing and removing trailing dots
func normalizeDomain(name string) string {
	return strings.ToLower(strings.TrimSuffix(name, "."))
}

// uniqueSortedList removes duplicates and sorts the subdomains
func (s *SubdomainsSearch) uniqueSortedList(elements []string) []string {
	set := make(map[string]struct{})
	for _, elem := range elements {
		set[elem] = struct{}{}
	}

	unique := make([]string, 0, len(set))
	for key := range set {
		unique = append(unique, key)
	}

	sort.Strings(unique)
	return unique
}

func filterByDomain(subdomains []string, domain string) []string {
	var filtered []string
	for _, sub := range subdomains {
		if strings.HasSuffix(sub, "."+domain) || sub == domain {
			filtered = append(filtered, sub)
		}
	}
	return filtered
}

// NewSubdomainsSearch initializes a SubdomainsSearch instance with default options
func NewSubdomainsSearch() *SubdomainsSearch {
	om := option.NewOptionManager()

	om.Register(option.NewOption("SOURCES_URI", defaultSources, true, "Sources to retrieve subdomains (all free sources)"))
	om.Register(option.NewOption("DOMAIN", "", true, "Domain to search subdomains"))
  om.Register(option.NewOption("FILTER_BY_DOMAIN", false, false, "Filter subdomains to match only the base domain"))

	helpManager := help.NewHelpManager()
	helpManager.Register("subdomains_search", "Subdomains search module", [][]string{
		{"SOURCES_URI", "not changable", "Uri from where to fetch subdomains about a domain"},
		{"DOMAIN", "example.com", "The domain to search for subdomains"},
		{"FILTER_BY_DOMAIN", "false", "Filter subdomains to match only the base domain"},
	})

	return &SubdomainsSearch{
		optionManager: om,
		name:          "Subdomains Search",
		author:        "Luca Cuzzolin",
		desc:          "Simple subdomains search",
		prompt:        "subdomains_search",
		help:          helpManager,
	}
}

// Options returns a list of formatted options for CLI display
func (s *SubdomainsSearch) Options() []map[string]string {
	opt := make([]map[string]string, len(s.optionManager.List()))
	for i, v := range s.optionManager.List() {
		opt[i] = v.Format()
	}
	return opt
}

// Run launches the subdomain enumeration across all configured sources
func (s *SubdomainsSearch) Run(ctx context.Context) [][]string {
    var allSubdomains []string
    var sources []string
    var domain string

    // Prendi le sorgenti e il dominio dalle opzioni
    if sopt, _ := s.optionManager.Get("SOURCES_URI"); sopt.Value != nil {
        sources = sopt.Value.([]string)
    }
    if dopt, _ := s.optionManager.Get("DOMAIN"); dopt.Value != nil {
        domain = dopt.Value.(string)
    }

    ch := make(chan []string, len(sources))
    var wg sync.WaitGroup

    // Disparo una goroutine per ogni fonte
    for _, u := range sources {
        // Controllo se è già arrivata la cancel
        select {
        case <-ctx.Done():
            return nil
        default:
        }

        wg.Add(1)
        go func(url string) {
            defer wg.Done()

            // Anche dentro il worker controllo la cancel
            select {
            case <-ctx.Done():
                return
            default:
            }

            subs, err := s.fetchSubdomains(url, domain)
            if err == nil {
                ch <- subs
            }
        }(u)
    }

    // Chiudo il canale quando tutte le fetch sono completate
    go func() {
        wg.Wait()
        close(ch)
    }()

    // Raccolgo i risultati finché non arriva una cancel o il canale non viene chiuso
    for {
        select {
        case <-ctx.Done():
            return nil
        case subs, ok := <-ch:
            if !ok {
                goto EMIT
            }
            allSubdomains = append(allSubdomains, subs...)
        }
    }



EMIT:

    if fopt, _ := s.optionManager.Get("FILTER_BY_DOMAIN"); fopt.Value != nil {
		    if fopt.Value.(bool) {
				    allSubdomains = filterByDomain(allSubdomains, domain)
		    }
    }
    // Unisco, filtro i duplicati e ordino
    s.results = s.uniqueSortedList(allSubdomains)

    var table [][]string
    for _, d := range s.results {
        table = append(table, []string{d})
    }
    return table
}

// Save writes the collected subdomains to a specified file
func (s *SubdomainsSearch) Save(filename string) error {
	if s.results == nil || len(s.results) == 0 {
		return fmt.Errorf("no results to save")
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, sub := range s.results {
		if _, err := file.WriteString(sub + "\n"); err != nil {
			return err
		}
	}
	return nil
}

// Set allows the user to modify a modifiable option (except SOURCES_URI)
func (s *SubdomainsSearch) Set(n string, v string) []string {
	if n == "SOURCES_URI" {
		return []string{"SOURCES_URI", "cannot change this value"}
	}

	om := *s.optionManager
	m, ok := om.Get(n)
	if ok {
		if n == "FILTER_BY_DOMAIN" {
			switch strings.ToLower(v) {
			case "true":
				m.Set(true)
				return []string{m.Name, "true"}
			case "false":
				m.Set(false)
				return []string{m.Name, "false"}
			default:
				return []string{m.Name, "Value must be 'true' or 'false'"}
			}
		}

		m.Set(v)
		return []string{m.Name, fmt.Sprintf("%v", m.Value)}
	}
	return []string{"Error", "Option not found"}
}

// Help returns usage help for the CLI
func (s *SubdomainsSearch) Help() [][]string {
	help, _ := s.help.Get(s.prompt)
	return help
}

func (s *SubdomainsSearch) Results() [][]string {
    var res [][]string
    for _, v := range s.results {
        res = append(res, []string{v})
    }
    return res
}

// Metadata functions
func (s *SubdomainsSearch) Name() string     { return s.name }
func (s *SubdomainsSearch) Author() string   { return s.author }
func (s *SubdomainsSearch) Description() string { return s.desc }
func (s *SubdomainsSearch) Prompt() string   { return s.prompt }
func (s *SubdomainsSearch) Running() bool    { return s.running }
func (s *SubdomainsSearch) Start() error     { s.running = true; return nil }
func (s *SubdomainsSearch) Stop() error      { s.running = false; return nil }
