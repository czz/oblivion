package tui

// ANSI escape sequences for terminal text formatting
const (
    // Text Colors
    RED    = "\033[31m"
    GREEN  = "\033[32m"
    YELLOW = "\033[33m"
    BLUE   = "\033[34m"

    // Background Colors
    BACKGROUND_RED        = "\033[41m"
    BACKGROUND_GREEN      = "\033[42m"
    BACKGROUND_YELLOW     = "\033[43m"
    BACKGROUND_DARKGRAY   = "\033[100m"  // Fixed typo: BACKGROUD_DARKGRAY → BACKGROUND_DARKGRAY
    BACKGROUND_LIGHTBLUE  = "\033[104m"

    // Foreground Colors
    FOREGROUND_BLACK = "\033[30m"        // Fixed typo: FOREGROUD_BLACK → FOREGROUND_BLACK
    FOREGROUND_WHITE = "\033[97m"        // Fixed typo: FOREGROUD_WHITE → FOREGROUND_WHITE

    // Text Attributes
    BOLD  = "\033[1m"
    DIM   = "\033[2m"

    // Reset
    RESET = "\033[0m"  // Resets all formatting
)

// ctrl holds common escape sequence prefixes for control characters (e.g., ESC)
var ctrl = []string{"\033", "\\e", "\x1b"}
