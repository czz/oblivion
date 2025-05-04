package tui

func (t *Tui) Pack(style, text string) string {
    if t.effects {
        return style + text + RESET
    }
    return text
}

func (t *Tui) Style(text string, styles ...string) string {
    if !t.effects {
        return text
    }
    prefix := ""
    for _, s := range styles {
        prefix += s
    }
    return prefix + text + RESET
}

// Common style methods
func (t *Tui) Bold(text string) string    { return t.Pack(BOLD, text) }
func (t *Tui) Dim(text string) string     { return t.Pack(DIM, text) }
func (t *Tui) Red(text string) string     { return t.Pack(RED, text) }
func (t *Tui) Green(text string) string   { return t.Pack(GREEN, text) }
func (t *Tui) Blue(text string) string    { return t.Pack(BLUE, text) }
func (t *Tui) Yellow(text string) string  { return t.Pack(YELLOW, text) }
