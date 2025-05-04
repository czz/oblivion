package option

// OptionManager manages a collection of options.
type OptionManager struct {
    options map[string]*Option // A map of options indexed by their name.
}

// NewOptionManager creates a new OptionManager instance.
func NewOptionManager() *OptionManager {
    return &OptionManager{
        options: make(map[string]*Option), // Initializes the map of options.
    }
}

// Register registers a new option with the OptionManager.
func (m *OptionManager) Register(opt *Option) {
    m.options[opt.Name] = opt // Adds the option to the map of options.
}

// Get retrieves an option by its name from the OptionManager.
// Returns the option and a boolean indicating if it exists.
func (m *OptionManager) Get(name string) (*Option, bool) {
    opt, ok := m.options[name] // Attempts to retrieve the option from the map.
    return opt, ok
}

// List returns a slice of all options managed by the OptionManager.
func (m *OptionManager) List() []*Option {
    opts := []*Option{}
    for _, o := range m.options {
        opts = append(opts, o) // Adds each option to the slice.
    }
    return opts
}
