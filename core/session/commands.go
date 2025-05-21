package session

import (
    "fmt"
    "os"
    "strings"
    "context"

    "github.com/czz/oblivion/core/tui"
    "github.com/czz/oblivion/modules"
)

// registerCommands initializes the command map with available command handlers.
func (s *Session) registerCommands() {
    s.commands = map[string]commandFunc{
        "exit":    s.handleExit,
        "help":    s.handleHelp,
        "search":  s.handleSearch,
        "use":     s.handleUse,
        "options": s.handleOptions,
        "set":     s.handleSet,
        "run":     s.handleRun,
        "stop":    s.handleStop,
        "show":    s.handleShow,
        "save":    s.handleSave,
        "back":    s.handleBack,
    }
}

// handleHelp displays help for core and module commands.
func (s *Session) handleHelp(args []string) {
    coreHelp := [][]string{
        {"Core Commands", ""},
        {"=============", ""},
        {"  Command", "Description"},
        {"  -------", "-----------"},
        {"  help", "Help Menu"},
        {"", ""},
        {"Module Commands", ""},
        {"=============", ""},
        {"  Command", "Description"},
        {"  -------", "-----------"},
        {"  search", "Searches module names and descriptions"},
        {"  use <module>", "Selects a module to use"},
        {"  options", "Displays available options for the selected module"},
        {"  set <option> <value>", "Sets a value for a module option"},
        {"  run [&]","Executes the selected module in foreground (wait) or background (no wait)"},
        {"  stop [module_name]","stop the selected module in background"},
        {"  show [module_name]", "Show results of a module. Module name is optional when inside a module."},
        {"  save <filename>", "Saves the module output to the specified file"},
        {"  back", "Returns to core (exit module)"},
    }

    fmt.Println(s.Tui.Table(&tui.Table{LineSeparator: false, Padding: 1}, coreHelp))

    if s.isModuleActive() {
        module := *s.activeModule
        fmt.Println(s.Tui.Table(&tui.Table{LineSeparator: false, Padding: 1}, module.Help()))
    }
}

// handleSearch searches modules based on name, author, or description.
func (s *Session) handleSearch(args []string) {
    term := "*"
    if len(args) > 0 {
        term = strings.Join(args, " ")
    }

    results := [][]string{{"Prompt", "Name", "Author", "Description"}}
    seen := make(map[string]bool)
    words := strings.Split(term, " ")
    manager := *s.Modules

    for _, word := range words {
        for _, moduleName := range manager.List() {
            module, _ := manager.Get(moduleName)
            if strings.Contains(strings.ToLower(module.Name()), strings.ToLower(word)) ||
                strings.Contains(strings.ToLower(module.Description()), strings.ToLower(word)) ||
                strings.Contains(strings.ToLower(module.Author()), strings.ToLower(word)) ||
                word == "*" {

                values := []string{module.Prompt(), module.Name(), module.Author(), module.Description()}
                key := strings.Join(values, "|")

                if !seen[key] {
                    results = append(results, values)
                    seen[key] = true
                }
            }
        }
    }

    if len(results) > 1 {
        fmt.Println(s.Tui.Table(&tui.Table{LineSeparator: true, Padding: 1}, results))
    }
}

// handleUse sets the currently active module.
func (s *Session) handleUse(args []string) {
    if len(args) != 1 {
        fmt.Println(s.Tui.Red("Usage: use <module>"))
        return
    }

    modulePrompt := args[0]
    manager := *s.Modules
    module, ok := manager.Get(modulePrompt)
    if !ok {
        fmt.Println(s.Tui.Red("Module not found: " + modulePrompt))
    } else {
        s.activeModule = &module
        s.updatePrompt()
    }
}

// handleOptions displays available options for the currently active module.
func (s *Session) handleOptions(args []string) {
    if !s.isModuleActive() {
        fmt.Println(s.Tui.Red("No active module."))
        return
    }

    module := *s.activeModule
    optionsTable := [][]string{
        {"Module options " + module.Prompt(), "", "", ""},
        {"  Name", "Current Setting", "Required", "Description"},
        {"  ----", "---------------", "--------", "-----------"},
    }

    for _, opt := range module.Options() {
        val := opt["value"]
        if val == "<nil>" {
            val = ""
        }
        line := []string{"  " + opt["name"], val, opt["required"], opt["description"]}
        optionsTable = append(optionsTable, line)
    }

    fmt.Println(s.Tui.Table(&tui.Table{
        LineSeparator: false,
        Padding:       1,
        MaxWidth:      s.terminalWidth / 3,
    }, optionsTable))
}

// handleSet updates an option value for the active module.
func (s *Session) handleSet(args []string) {
    if len(args) < 2 || !s.isModuleActive() {
        fmt.Println(s.Tui.Red("Usage: set <option> <value>"))
        return
    }

    key := args[0]
    value := strings.Join(args[1:], " ")
    module := *s.activeModule
    result := module.Set(key, value)

    if len(result) == 2 {
        fmt.Println(s.Tui.Yellow(result[0] + " => " + result[1]))
    }
}

// handleRun executes the active module and prints the results.
func (s *Session) handleRun(args []string) {
    // Check if a module is active
    if !s.isModuleActive() {
        fmt.Println(s.Tui.Red("No active module."))
        return
    }

    module := *s.activeModule

    prompt := module.Prompt()

        // Se è già in esecuzione, segnalo…
    s.mu.Lock()
    if _, running := s.runningCancels[prompt]; running {
        s.mu.Unlock()
        fmt.Println(s.Tui.Yellow("Module "+prompt+" is already running."))
        return
    }
    s.mu.Unlock()

    ctx, cancel := context.WithCancel(context.Background())

    runInBackground := len(args)>0 && args[0]=="&"
    if runInBackground {
        go func() {
            module.Start()
            defer module.Stop()
            defer func() {
                // quando finisce, rimuovo dal registro
                s.mu.Lock()
                delete(s.runningCancels, prompt)
                s.mu.Unlock()
            }()
            module.Run(ctx)
            fmt.Println(s.Tui.Green("\nModule "+prompt+" finished in background"))
            s.Refresh()
        }()
        // prima di tornare, registro il cancel
        s.mu.Lock()
        s.runningCancels[prompt] = cancel
        s.mu.Unlock()

        fmt.Println(s.Tui.Yellow("Module "+prompt+" started in background."))
    } else {
        module.Start()
        defer module.Stop()
        // Esecuzione in foreground non richiede registrazione a meno che tu non voglia fermarla
        results := module.Run(ctx)
        fmt.Println(s.Tui.Table(&tui.Table{
            LineSeparator: false,
            Padding:       1,
            MaxWidth:      s.terminalWidth / 3,
        }, results))
    }
}

func (s *Session) handleStop(args []string) {
    name := ""
    if len(args) > 0 {
        name = args[0]
    } else if s.isModuleActive() {
        name = (*s.activeModule).Prompt()
    } else {
        fmt.Println(s.Tui.Red("Usage: stop <module>"))
        return
    }

    s.mu.Lock()
    cancel, ok := s.runningCancels[name]
    s.mu.Unlock()
    if !ok {
        fmt.Println(s.Tui.Red("Module "+name+" is not running."))
        return
    }

    cancel() // invoca ctx.Done() per quel modulo
    s.mu.Lock()
    delete(s.runningCancels, name)
    s.mu.Unlock()

    fmt.Println(s.Tui.Green("Sent stop signal to module "+name))
}



// handleShow displays the output of a module (active or specified by name)
func (s *Session) handleShow(args []string) {
    var module modules.Module
    var ok bool

    if len(args) > 0 {
        manager := *s.Modules
        module, ok = manager.Get(args[0])
        if !ok {
            fmt.Println(s.Tui.Red("No module found. Usage: show <module>"))
            return
        }
    } else {
        if !s.isModuleActive() {
            fmt.Println(s.Tui.Red("Usage: show <module>"))
            return
        }
        module = *s.activeModule
    }

    if module.Running() {
        fmt.Println(s.Tui.Yellow(fmt.Sprintf("Module %s is running in the background. Try later.", module.Prompt())))
        return
    }

    results := module.Results()
    fmt.Println(s.Tui.Table(&tui.Table{
        LineSeparator: false,
        Padding:       1,
        MaxWidth:      s.terminalWidth / 3,
    }, results))
}


// handleSave saves the output of the active module to a file.
func (s *Session) handleSave(args []string) {

    if !s.isModuleActive() {
        fmt.Println(s.Tui.Red("No active module."))
        return
    }

    if len(args) != 1 {
        fmt.Println(s.Tui.Red("Usage: save <filename>"))
        return
    }

    filename := args[0]
    module := *s.activeModule
    err := module.Save(filename)

    if err != nil {
        fmt.Println(s.Tui.Red("Error saving: " + err.Error()))
        s.logError(err, "saving module output")
    } else {
        fmt.Println(s.Tui.Green("Results saved to: " + filename))
    }
}

// handleBack exits the currently active module.
func (s *Session) handleBack(args []string) {
    if s.isModuleActive() {
        s.activeModule = nil
        s.updatePrompt()
        fmt.Println(s.Tui.Green("Returned to core."))
    }
}

// handleExit terminates the session and exits the application.
func (s *Session) handleExit(args []string) {
    fmt.Println(s.Tui.Green("Exiting Oblivion."))
    s.Stop()
    os.Exit(0)
}
