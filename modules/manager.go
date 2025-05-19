package modules

// ModuleManager is responsible for managing modules dynamically.
type ModuleManager struct {
    modules map[string]Module
}

// NewManager creates a new ModuleManager instance.
func NewModuleManager() *ModuleManager {
    return &ModuleManager{
        modules: make(map[string]Module),
    }
}

// Register adds a module to the manager.
func (m *ModuleManager) Register(module Module) {
    m.modules[module.Prompt()] = module
}

// Get retrieves a registered module by name.
func (m *ModuleManager) Get(name string) (Module, bool) {
    mod, ok := m.modules[name]
    return mod, ok
}

// List returns a list of names of all registered modules.
func (m *ModuleManager) List() []string {
    keys := make([]string, 0, len(m.modules))
    for k := range m.modules {
        keys = append(keys, k)
    }
    return keys
}

