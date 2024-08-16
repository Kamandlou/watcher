# File Watcher
 File Watcher is a command-line tool written in Go that monitors specified file types in a directory and executes a command whenever a file of those types is modified.
 
 # Features
- Monitors specified file types (e.g., .go, .md) in a directory.
- Executes a custom command (e.g., go run main.go, go build) upon detecting a file change.
- Supports verbose mode for detailed logging of file change events.

# Installation
```bash
go install github.com/Kamandlou/watcher@latest
```

# Build from source
1. Clone the repository:
    ```bash
    git clone https://github.com/Kamandlou/watcher.git
    ```
2. Navigate to the project directory:
    ```bash
    cd watcher
    ```
3. Build and install the application:
    ```bash
    go install
    ```
    This will build the application and install it into your Go bin directory.

# Usage
To run the File Watcher, use the following command:
```bash
watcher [flags]
```
## Flags
- path: Specify the directory path to monitor.
- verbose: Enable verbose mode to print detailed file change events.
- command: Specify the command to execute when a file changes.
- delay: Set delay time to execute command.
- types: Specify the file types to watch, separated by commas.
- period: Set period time to watch.
- help: Display help information.

## Example Usage
```bash
# Run watcher with default settings
wacther

# Customize path, file types, command, and enable verbose mode
watcher --path="./" --types=".go" --commnad="go run main.go" --verbose

# Monitor specific directory with different file types and command
watcher --path="/path/to/directory" --types=".js,.html" --command="npm run build" --verbose
```

# Contributing
Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on [GitHub](https://github.com/Kamandlou/watcher).

# License
This project is licensed under the MIT [License](LICENSE) - see the [License](LICENSE) file for details.
