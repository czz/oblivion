
# Oblivion

![logo](https://oblivion.czz78.com/images/oblivion_logo_250.png)

**Oblivion is a modular framework written in Go designed for web hacking activities. It offers an interactive REPL interface and a highly extensible modular structure, ideal for developing custom tools for web content analysis and manipulation.**

---

## Key Features

* **Interactive REPL** for dynamic module management
* **Centralized module management**
* **Configurable options** for each module
* **Tabular output** of results
* **Data saving** to files

---

## Project Architecture

```
oblivion/
├── core/
│   ├── session/     # REPL logic and module management
│   └── tui/         # Text user interface and tabular rendering
├── modules/         # Specific modules (e.g., spider, parser, etc.)
├── utils/
│   ├── help/        # Help utilities
│   ├── option/      # Option management
└── main.go          # Entry point
```

---

## Usage

### Starting the REPL

```bash
$ go run main.go
```

### Available Commands

* `use <module>` - Activate a module
* `options` - Show the options for the active module
* `set <name> <value>` - Set an option for the module
* `run [&]` - Execute the module, with & ans arg will run in background
* `show [module_name]` - Show the results
* `save <file>` - Save the results
* `back` - Go back to the global context
* `exit | quit` - Exit the REPL

### Example

```bash
oblv> use webspider
spider> options
spider> set DOMAIN https://example.com
spider> run
spider> save results.txt
```

---

## Creating a Module

Each module must implement the `module.Module` interface, providing:

* `Prompt()` string
* `Options()` \[]map\[string]string
* `Set(string, string)` \[]string
* `Run()` \[]\[]string
* `Results()` \[]\[]string
* `Save(string)` error

Optional:

* `Help()` \[]\[]string

Register the module with `modules.Register("name", NewModule())`.

---

## Building the Project

1. **Clone the repository**:

    ```bash
    git clone https://github.com/czz/oblivion.git
    ```

2. **Navigate to the project folder**:

    ```bash
    cd oblivion
    ```

3. **Install the dependencies**:

    ```bash
    go mod tidy
    ```

4. **Build the executable**:

    ```bash
    go build -o oblivion main.go
    ```

5. **Run the project**:

    ```bash
    ./oblivion
    ```

## Requirements

* Go 1.23.5 or higher

---

## Author

Developed by [czz](https://github.com/czz) as part of the *Oblivion* project.

---

## License

GPL 3 license.
