package modules

import (
    "fmt"

    "github.com/czz/oblivion/modules/subdomains_search"
//    "github.com/czz/oblivion/modules/s3enum"
    "github.com/czz/oblivion/modules/dnsbrute"
    "github.com/czz/oblivion/modules/portscanner"
    "github.com/czz/oblivion/modules/webspider"
)

// Module defines the methods that every module should implement
type Module interface {
    Name() string            // Name of the module
    Description() string     // Description of the module
    Author() string          // Author of the module
    Prompt() string          // Prompt related to the module
    Set(string, string) []string   // Set a value for the module
    Run() [][]string         // Run the module
    Running() bool           // Check if the module is running
    Start() error            // Start the module
    Stop() error             // Stop the module
    Options() []map[string]string // Module options (name, value, etc.)
    Save(string) error       // Save the module output to a file
    Help() [][]string
    Results() [][]string
}

// LoadModules loads all available modules and returns them in a slice
func LoadModules() *ModuleManager {

    manager := NewModuleManager()
    manager.Register(dnsbrute.NewDNSBrute())
    manager.Register(portscanner.NewPortScanner())
    manager.Register(subdomains_search.NewSubdomainsSearch())
//    manager.Register(s3enum.NewS3enum())
    manager.Register(webspider.NewWebSpider())

    // Log the number of modules loaded
    fmt.Printf("Loaded %d modules\n\n", len(manager.List()))

    return manager
}
