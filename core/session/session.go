package session

import (
    "os"
    "time"
    "unicode/utf8"
    "context"
    "sync"

    "github.com/chzyer/readline"
    "github.com/czz/oblivion/modules"
    "github.com/czz/oblivion/core/tui"
)

// commandFunc defines the function signature for a CLI command handler.
type commandFunc func(args []string)

// Session represents a CLI session with state, modules, and user interaction.
type Session struct {
    StartedAt     time.Time                  // Timestamp when the session started
    Active        bool                       // Indicates if the session is active
    Modules       **modules.ModuleManager    // Pointer to the module manager
    Tui           *tui.Tui                   // Text-based UI utilities
    ReadLine      *readline.Instance         // Readline instance for CLI interaction
    activeModule  *modules.Module            // Currently active module
    terminalWidth int                        // Terminal width in characters
    commands      map[string]commandFunc     // Registered CLI commands
    runningCancels map[string]context.CancelFunc
    mu             sync.Mutex
}

// NewSession initializes and returns a new Session instance.
func NewSession() *Session {
    mods := modules.LoadModules()
    s := &Session{
        Active:       false,
        Modules:      &mods,
        Tui:          tui.NewTui(),
        activeModule: nil,
        commands:     make(map[string]commandFunc),
        runningCancels: make(map[string]context.CancelFunc),
    }
    s.registerCommands()
    return s
}

// Start begins the interactive session, setting up readline and logging.
func (s *Session) Start() {
    // Get the user's home directory
    userDir, err := os.UserHomeDir()
    if err != nil {
        panic("Unable to determine user home directory: " + err.Error())
    }

    // Create ~/.oblivion directory if it doesn't exist
    oblivionDir := userDir + "/.oblivion"
    if _, err := os.Stat(oblivionDir); os.IsNotExist(err) {
        if err := os.MkdirAll(oblivionDir, 0755); err != nil {
            panic("Unable to create .oblivion directory: " + err.Error())
        }
    }

    // Configure prompt and readline with history and autocomplete
    s.Tui.SetPrompt(s.Tui.Green("oblv>"))
    rl, err := readline.NewEx(&readline.Config{
        Prompt:          s.Tui.GetPrompt(),
        HistoryFile:     oblivionDir + "/history.tmp",
        InterruptPrompt: "^C",
        EOFPrompt:       "exit",
        HistoryLimit:    500,
        AutoComplete:    s.commandCompleter(),
    })
    if err != nil {
        panic(err)
    }

    s.ReadLine = rl
    s.Active = true

    // Determine terminal width
    width, _, err := readline.GetSize(int(os.Stdout.Fd()))
    if err != nil {
        s.terminalWidth = 80 // fallback width
    } else {
        s.terminalWidth = width
    }

    logInfo("Session started") // Log session start
    s.StartedAt = time.Now()
}

// isModuleActive checks if a module is currently selected.
func (s *Session) isModuleActive() bool {
    return s.activeModule != nil
}

// calculateActualWidth returns the visual width of a string,
// accounting for multi-byte Unicode characters.
func (s *Session) calculateActualWidth(str string) int {
    width := 0
    for i := 0; i < len(str); i++ {
        _, size := utf8.DecodeRuneInString(str[i:])
        width++
        if size > 1 {
            width++
        }
        i += size - 1
    }
    return width
}

// commandCompleter builds a dynamic autocomplete tree based on current session state.
func (s *Session) commandCompleter() *readline.PrefixCompleter {
    useChildren := []readline.PrefixCompleterInterface{}
    setChildren := []readline.PrefixCompleterInterface{}

    manager := *s.Modules
    modulesList := manager.List()
    for _, name := range modulesList {
        mod, ok := manager.Get(name)
        if ok {
            useChildren = append(useChildren, readline.PcItem(mod.Prompt()))
        }
    }

    // Base commands available in all contexts
    base := []readline.PrefixCompleterInterface{
        readline.PcItem("help"),
        readline.PcItem("search"),
        readline.PcItem("use", useChildren...),
        readline.PcItem("show", useChildren...),
        readline.PcItem("stop", useChildren...),
        readline.PcItem("exit"),
    }

    // Additional commands if a module is currently active
    if s.isModuleActive() {
        m := *s.activeModule
        for _, opt := range m.Options() {
            setChildren = append(setChildren, readline.PcItem(opt["name"]))
        }

        setRunBackground := []readline.PrefixCompleterInterface{
            readline.PcItem("&"),
        }

        base = append(base,
            readline.PcItem("options"),
            readline.PcItem("set", setChildren...),
            readline.PcItem("run", setRunBackground...),
            readline.PcItem("save"),
            readline.PcItem("back"),
        )
    }

    return readline.NewPrefixCompleter(base...)
}
