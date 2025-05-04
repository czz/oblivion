package tui

import(
  "strings"
  "unicode/utf8"
)

type lineSeparator struct {
    header [3]string   // Characters for top border: left, middle, right
    content [3]string  // Characters for middle row separator: left, middle, right
    footer [3]string   // Characters for bottom border: left, middle, right
}

type Table struct {
    LineSeparator bool // Whether to include table borders
    Padding int        // Number of spaces around cell content
    MaxWidth int       // Optional max width for each column
}

// Table renders a matrix as a formatted table using the current Tui and Table settings.
func (t *Tui) Table(table *Table, matrix [][]string) string {
    if !table.tableHasUniformColumns(matrix) {
        return "Error: can't print table, has not uniform columns"
    }

    var result strings.Builder
    widths := table.calcMaxWidths(matrix)

    // Apply MaxWidth constraints to each column
    for j := range widths {
        if table.MaxWidth > 0 && widths[j]+2*table.Padding > table.MaxWidth {
            widths[j] = table.MaxWidth - 2*table.Padding
        }
    }

    // Define border styles (box-drawing characters)
    sep := &lineSeparator{
        header:  [3]string{"┌", "┬", "┐"},
        content: [3]string{"├", "┼", "┤"},
        footer:  [3]string{"└", "┴", "┘"},
    }

    if table.LineSeparator {
        result.WriteString(table.buildSep(sep.header[0], sep.header[1], sep.header[2], widths) + "\n")
    }

    for i, row := range matrix {
        // Wrap each cell to handle multiline content
        wrappedCells := make([][]string, len(row))
        maxLines := 1
        for j, cell := range row {
            wrapWidth := widths[j]
            wrapped := wrapCell(cell, wrapWidth)
            wrappedCells[j] = wrapped
            if len(wrapped) > maxLines {
                maxLines = len(wrapped)
            }
        }

        // Print each visual row line-by-line
        for line := 0; line < maxLines; line++ {
            if table.LineSeparator {
                result.WriteString("│")
            }
            for j := range row {
                var content string
                if line < len(wrappedCells[j]) {
                    content = wrappedCells[j][line]
                } else {
                    content = ""
                }
                padding := widths[j] - utf8.RuneCountInString(content)
                result.WriteString(strings.Repeat(" ", table.Padding) + content + strings.Repeat(" ", table.Padding+padding))
                if table.LineSeparator {
                    result.WriteString("│")
                }
            }
            result.WriteString("\n")
        }

        // Draw separator between rows or at the end
        if table.LineSeparator {
            if i < len(matrix)-1 {
                result.WriteString(table.buildSep(sep.content[0], sep.content[1], sep.content[2], widths) + "\n")
            } else {
                result.WriteString(table.buildSep(sep.footer[0], sep.footer[1], sep.footer[2], widths) + "\n")
            }
        }
    }

    return result.String()
}

// Checks if all rows in the matrix have the same number of columns
func (t *Table) tableHasUniformColumns(matrix [][]string) bool {
    if len(matrix) == 0 {
        return true
    }

    expectedCols := len(matrix[0])
    for _, row := range matrix {
        if len(row) != expectedCols {
            return false
        }
    }

    return true
}

// Builds a border/separator line using the given characters
func (t *Table) buildSep(left, mid, right string, widths []int) string {
    var sb strings.Builder
    sb.WriteString(left)
    for i, w := range widths {
        if i > 0 {
            sb.WriteString(mid)
        }
        sb.WriteString(strings.Repeat("─", w+2*t.Padding))
    }
    sb.WriteString(right)
    return sb.String()
}

// Calculates the maximum width of each column based on cell contents
func (t *Table) calcMaxWidths(matrix [][]string) []int {
    if len(matrix) == 0 {
        return []int{}
    }
    cols := len(matrix[0])
    widths := make([]int, cols)
    for _, row := range matrix {
        for j, cell := range row {
            w := utf8.RuneCountInString(cell)
            if w > widths[j] {
                widths[j] = w
            }
        }
    }
    return widths
}

// Truncates a string to fit within maxLen runes (adds "..." if space allows)
func truncateCell(s string, maxLen int) string {
    if utf8.RuneCountInString(s) <= maxLen {
        return s
    }
    truncated := []rune(s)
    if maxLen > 3 {
        return string(truncated[:maxLen-3]) + "..."
    }
    return string(truncated[:maxLen])
}

// Splits a string into multiple lines of maxWidth runes each
func wrapCell(s string, maxWidth int) []string {
    var lines []string
    runes := []rune(s)
    for i := 0; i < len(runes); i += maxWidth {
        end := i + maxWidth
        if end > len(runes) {
            end = len(runes)
        }
        lines = append(lines, string(runes[i:end]))
    }
    return lines
}
