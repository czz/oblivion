package tui

import(
    "os"
)

// Tui represents a simple text-based UI system.
type Tui struct {
    effects bool   // Determines whether terminal effects (like colors) are enabled.
    prompt  string // The prompt string displayed before user input.
}

// NewTui creates a new Tui instance.
// If effects are disabled or terminal is not appropriate, disables effects accordingly.
func NewTui(effects ...bool) *Tui {
    check_term := true

    // If one argument is passed, skip terminal checks and use that value
    if len(effects) == 1 {
        check_term = false
    }

    // Disable effects if TERM is not set or is a dumb terminal
    if term := os.Getenv("TERM"); term == "" {
        check_term = false
    } else if term == "dumb" {
        check_term = false
    }

    return &Tui{
        effects: check_term,
        prompt:  "tui>",
    }
}

// HasEffectsEnable returns whether terminal effects are enabled.
func (t *Tui) HasEffectsEnable() bool {
    return t.effects
}

// SetPrompt sets the terminal prompt to the given string.
func (t *Tui) SetPrompt(s string) {
    t.prompt = s
}

// GetPrompt returns the current terminal prompt string.
func (t *Tui) GetPrompt() string {
    return t.prompt
}
