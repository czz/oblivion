package session

import (
    "fmt"
    "strings"
    "io"

    "github.com/chzyer/readline"
)

// ReadlineLoop is the main loop for reading user input from the terminal.
// It handles input parsing, command dispatch, and interface refresh.
func (s *Session) ReadlineLoop() {
    if !s.Active {
        return
    }

    for {
        line, err := s.ReadLine.Readline()
        if err == io.EOF {
            fmt.Println(s.Tui.Green("Exiting ..."))
            s.Stop()
            return
        }

        if err != nil && err != readline.ErrInterrupt {
          fmt.Println(err)
            s.Stop()
            fmt.Sprintf(s.Tui.Red("Error reading command line: %s"), err)
            return
        }

        line = strings.TrimSpace(line)
        if line == "" {
            continue // Skip empty input
        }

        logCommand(line) // Log the user command

        parts := strings.Fields(line)
        if len(parts) == 0 {
            continue
        }

        cmdName := strings.ToLower(parts[0])
        args := parts[1:]

        if handler, found := s.commands[cmdName]; found {
            handler(args) // Execute the matched command handler
        } else {
            fmt.Println(s.Tui.Red("Unknown command: " + cmdName))
        }

        // Refresh autocompletion and prompt after executing the command
        s.ReadLine.Config.AutoComplete = s.commandCompleter()
        s.Refresh()
    }
}

// updatePrompt updates the CLI prompt depending on the active module.
func (s *Session) updatePrompt() {
    if s.isModuleActive() {
        m := *s.activeModule
        s.ReadLine.SetPrompt(fmt.Sprintf("%s%s>", s.Tui.GetPrompt(), s.Tui.Blue(m.Prompt())))
    } else {
        s.ReadLine.SetPrompt(s.Tui.Green("oblv>"))
    }
}

// Refresh updates the CLI prompt and forces a redraw of the readline interface.
func (s *Session) Refresh() {
    s.updatePrompt()
    s.ReadLine.Refresh()
}

// Stop ends the session, closes the readline interface, and logs the shutdown.
func (s *Session) Stop() {
    s.Active = false
    s.ReadLine.Close()
    logSession("[INFO] Session stopped.")
}
