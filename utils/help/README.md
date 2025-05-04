
# Help Module - Documentation

## Description
The **Help Module** is designed to provide structured help information for various modules in a project. It allows the registration of help content for different modules and enables easy retrieval of this information through a unified system. The module supports the management of option tables, descriptions, and syntax, providing users with detailed information about commands and their usage.

## Features
- Register help content for different modules.
- Retrieve structured help tables with command options, syntax, and descriptions.
- Provide well-formatted help output for different modules.
- Allows for easy addition of new commands and options in the future.

## Structure
The **Help Manager** contains a list of **Help Entries**. Each **Help Entry** represents a module and contains:
- **Name**: The name of the module.
- **Description**: A description of what the module does.
- **Table**: A 2D slice containing detailed information about the options (syntax, description) for that module.

## How It Works

### `HelpManager`
The `HelpManager` handles the registration and retrieval of help entries. It provides the following methods:
- **Register**: Register a new help entry for a module.
- **Get**: Retrieve the help content for a specific module by name.
- **List**: List all registered help entries.

### `HelpEntry`
The `HelpEntry` represents the help content for a specific module. It contains:
- **Name**: Name of the module.
- **Description**: A short description of what the module does.
- **Table**: A table with options, syntax, and descriptions for the commands within the module.

## Installation
To use the Help Module in your project, simply include the `help` package in your Go project.

```bash
go get github.com/czz/help
```

## Usage

### Register a New Help Entry
You can register a new help entry for a module using the `Register` method. Here's an example of how to do this:

```go
helpManager := help.NewHelpManager()

table := [][]string{
    {"  Option", "Syntax", "Description"},
    {"  -------", "------", "-----------"},
    {"  TARGETS", "<target>", "Targets to scan (comma-separated list or file path)"},
    {"  PORTS", "80, 443", "Ports to scan (comma-separated list or range)"},
}

helpManager.Register("Port Scanner", "This is a simple port scanner tool.", table)
```

### Retrieve Help Information
You can retrieve the help information for a specific module using the `Get` method:

```go
helpContent, found := helpManager.Get("Port Scanner")
if found {
    // Process helpContent
    // This will contain the formatted table for "Port Scanner"
}
```

### List All Registered Help Entries
You can list all the registered help entries using the `List` method:

```go
entries := helpManager.List()
for _, entry := range entries {
    fmt.Println(entry.Name, entry.Description)
}
```

## Example Output

### For "Port Scanner" Module
If the help entry for the "Port Scanner" module is retrieved, the output will look like this:

```
Options Port Scanner
====================
  Option    Syntax    Description
  -------   ------    -----------
  TARGETS   <target>  Targets to scan (comma-separated list or file path)
  PORTS     80, 443  Ports to scan (comma-separated list or range)
```

### For Other Modules
You can define additional modules and their help content in a similar way. Each entry will have a corresponding table to display the options, syntax, and description.

## License
This project is licensed under the MIT License - see the LICENSE file for details.

## Author
- Luca Cuzzolin

## Contributions
Feel free to open issues or submit pull requests. All contributions are welcome!
