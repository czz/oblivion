package option

/*
o := NewOption("MIXED", []interface{}{42, "hello", true, 3.14}, false, "mixed slice")
fmt.Println(o.Format(false)["value"]) // ➜ 42, "hello", true, 3.14
*/

import (
    "fmt"
    "reflect"
    "strings"
)

// Option represents a configuration option with a name, value, required status, and description.
type Option struct {
    Name        string      // The name of the option.
    Value       interface{} // The value of the option, which can be of any type.
    Required    bool        // Indicates if the option is required.
    Description string      // A description of the purpose of the option.
}

// NewOption creates and returns a new Option with the provided attributes.
func NewOption(name string, value interface{}, required bool, description string) *Option {
    return &Option{
        Name:        name,
        Value:       value,
        Required:    required,
        Description: description,
    }
}

// Format formats the option as a map, returning either Go-syntax format (true) or default format (false).
func (o *Option) Format(f ...bool) map[string]string {
    var c bool = false
    if len(f) > 0 {
        c = f[0]
    }

    opt := make(map[string]string)
    opt["name"] = o.Name

    if c {
        opt["value"] = fmt.Sprintf("%#v", o.Value)
        opt["required"] = fmt.Sprintf("%#v", o.Required)
    } else {
        opt["value"] = o.formatValue(o.Value)
        opt["required"] = fmt.Sprintf("%v", o.Required)
    }

    opt["description"] = o.Description
    return opt
}

// Set allows updating the value of the option
func (o *Option) Set(v interface{}) {
    o.Value = v
}

// Get returns the value of the option using reflection to dynamically access the value.
func (o *Option) Get() interface{} {
    return reflect.ValueOf(o.Value).Interface()
}

// formatValue handles the formatting of the value based on its type.
func (o *Option) formatValue(v interface{}) string {
    switch reflect.TypeOf(v).Kind() {
    case reflect.Slice:
        return formatSlice(v)
    case reflect.Map:
        return formatMap(v)
    case reflect.Struct:
        return formatStruct(v)
    case reflect.Ptr:
        return formatPointer(v)
    default:
        return fmt.Sprintf("%v", v)
    }
}

// formatSlice handles the formatting of slices, including mixed types.
func formatSlice(v interface{}) string {
    val := reflect.ValueOf(v)

    if val.Kind() != reflect.Slice {
        return fmt.Sprintf("%v", v)
    }

    if val.Len() == 0 {
        return ""
    }

    var elements []string

    for i := 0; i < val.Len(); i++ {
        elem := val.Index(i).Interface()

        switch e := elem.(type) {
        case int, int8, int16, int32, int64:
            elements = append(elements, fmt.Sprintf("%d", e))
        case uint, uint8, uint16, uint32, uint64:
            elements = append(elements, fmt.Sprintf("%d", e))
        case float32, float64:
            elements = append(elements, fmt.Sprintf("%.2f", e))
        case string:
            elements = append(elements, fmt.Sprintf("%q", e))
        case bool:
            elements = append(elements, fmt.Sprintf("%t", e))
        default:
            elements = append(elements, fmt.Sprintf("%v", e))
        }
    }

    return strings.Join(elements, ", ")
}

// formatMap handles the formatting of maps.
func formatMap(v interface{}) string {
    val := reflect.ValueOf(v)
    var elements []string
    for _, key := range val.MapKeys() {
        elements = append(elements, fmt.Sprintf("%v: %v", key, val.MapIndex(key)))
    }
    return strings.Join(elements, ", ")
}

// formatStruct handles the formatting of structs.
func formatStruct(v interface{}) string {
    val := reflect.ValueOf(v)
    var fields []string
    for i := 0; i < val.NumField(); i++ {
        fields = append(fields, fmt.Sprintf("%s: %v", val.Type().Field(i).Name, val.Field(i).Interface()))
    }
    return strings.Join(fields, ", ")
}

// formatPointer handles the formatting of pointers.
func formatPointer(v interface{}) string {
    val := reflect.ValueOf(v)
    if val.IsNil() {
        return "nil"
    }
    return fmt.Sprintf("%v", val.Elem())
}
