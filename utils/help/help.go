package help

import "strings"

// HelpEntry represents a single help section with a name, description, and table of entries.
type HelpEntry struct {
    Name        string     // Name of the help topic
    Description string     // Short description of the topic
    Table       [][]string // Help table: typically [option, syntax, description]
}

// HelpManager manages multiple HelpEntry objects, allowing registration and retrieval.
type HelpManager struct {
    entries map[string]*HelpEntry
}

// NewHelpManager initializes and returns a new HelpManager instance.
func NewHelpManager() *HelpManager {
    return &HelpManager{entries: make(map[string]*HelpEntry)}
}

// Register adds a new HelpEntry to the manager.
func (h *HelpManager) Register(name, description string, table [][]string) {
    h.entries[name] = &HelpEntry{
        Name:        name,
        Description: description,
        Table:       table,
    }
}

// Get retrieves the formatted help table for a given entry name.
// The output includes a header and a standardized table layout.
// The first column of each row is always prefixed with exactly two spaces.
func (h *HelpManager) Get(name string) ([][]string, bool) {
    entry, ok := h.entries[name]
    if !ok {
        return nil, false
    }

    res := [][]string{
        {"Options " + entry.Name, "", ""},
        {"========" + strings.Repeat("=", len(entry.Name)), "", ""},
        {"  Option", "Syntax", "Description"},
        {"  -------", "------", "-----------"},
    }

    for _, row := range entry.Table {
        newRow := make([]string, len(row))
        copy(newRow, row)
        if len(newRow) > 0 {
            // Always ensure exactly two leading spaces in the first column
            newRow[0] = "  " + strings.TrimLeft(newRow[0], " ")
        }
        res = append(res, newRow)
    }

    return res, true
}

// List returns all registered HelpEntries in a slice.
func (h *HelpManager) List() []*HelpEntry {
    var list []*HelpEntry
    for _, v := range h.entries {
        list = append(list, v)
    }
    return list
}
