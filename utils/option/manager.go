package option

// OptionManager manages a collection of options.
type OptionManager struct {
    options   map[string]*Option // A map of options indexed by their name.
    positions map[int]string     // Index of registration order
    nextPos   int                // Automatically managed position counter
}

// NewOptionManager creates a new OptionManager instance.
func NewOptionManager() *OptionManager {
    return &OptionManager{
        options:   make(map[string]*Option),
        positions: make(map[int]string),
        nextPos:   0,
    }
}

// Register registers a new option with the OptionManager, assigning position automatically.
func (m *OptionManager) Register(opt *Option) {
    m.options[opt.Name] = opt
    m.positions[m.nextPos] = opt.Name
    m.nextPos++
}

// Get retrieves an option by its name from the OptionManager.
func (m *OptionManager) Get(name string) (*Option, bool) {
    opt, ok := m.options[name]
    return opt, ok
}

// List returns all options in registration order.
func (m *OptionManager) List() []*Option {
    opts := make([]*Option, 0, len(m.options))
    for i := 0; i < m.nextPos; i++ {
        name, ok := m.positions[i]
        if !ok {
            continue
        }
        if opt, exists := m.options[name]; exists {
            opts = append(opts, opt)
        }
    }
    return opts
}
