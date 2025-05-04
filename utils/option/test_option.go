package option

import (
    "testing"
    "fmt"
)

// TestNewOption tests the creation of a new Option.
func TestNewOption(t *testing.T) {
    // Create a new Option using the NewOption function.
    m := NewOption("prova", int(1), true, "desc")

    // Check if the values are correct using assertions.
    if m.Name != "prova" {
        t.Errorf("expected Name to be 'prova', but got %s", m.Name)
    }
    if m.Value != 1 {
        t.Errorf("expected Value to be 1, but got %v", m.Value)
    }
    if m.Required != true {
        t.Errorf("expected Required to be true, but got %v", m.Required)
    }
    if m.Description != "desc" {
        t.Errorf("expected Description to be 'desc', but got %s", m.Description)
    }

    // Optionally, print the values for debugging purposes
    fmt.Printf("Name: %s, Value: %v, Required: %v, Description: %s\n", m.Name, m.Value, m.Required, m.Description)
}
