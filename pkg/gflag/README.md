# gflag - GNU-style Flag Parsing for Go

The `gflag` package provides command-line flag parsing with support for both POSIX-style short flags (`-x`) and GNU-style long flags (`--flag`), extending the functionality of Go's standard `flag` package.

## Features

- **Short flags**: `-v`, `-p 8080`, `-n name`
- **Long flags**: `--verbose`, `--port=8080`, `--name=name`
- **Combined short flags**: `-vdq` (equivalent to `-v -d -q`)
- **Mixed formats**: `-v --port=8080 -n name`
- **Argument separation**: Everything after `--` is treated as non-flag arguments
- **Compatible API**: Similar interface to Go's standard `flag` package

## Installation

```bash
go get github.com/nzions/sharedgolibs/pkg/gflag
```

## Basic Usage

```go
package main

import (
    "fmt"
    "github.com/nzions/sharedgolibs/pkg/gflag"
)

func main() {
    // Define flags with both long and short names
    var verbose = gflag.BoolP("verbose", "v", false, "enable verbose output")
    var port = gflag.IntP("port", "p", 8080, "server port")
    var name = gflag.StringP("name", "n", "myserver", "server name")
    
    // Parse command line arguments
    gflag.Parse()
    
    // Use the flags
    if *verbose {
        fmt.Println("Verbose mode enabled")
    }
    fmt.Printf("Server '%s' will listen on port %d\n", *name, *port)
    
    // Access remaining arguments
    args := gflag.Args()
    if len(args) > 0 {
        fmt.Printf("Additional arguments: %v\n", args)
    }
}
```

## Supported Flag Formats

### Long Flags
```bash
./myapp --verbose --port=8080 --name=myserver
./myapp --verbose --port 8080 --name myserver
```

### Short Flags
```bash
./myapp -v -p 8080 -n myserver
```

### Combined Short Flags
```bash
./myapp -vp 8080 -n myserver  # -v and -p combined
./myapp -vqd                  # multiple boolean flags combined
```

### Mixed Formats
```bash
./myapp -v --port=8080 -n myserver
./myapp --verbose -p 8080 --name=myserver
```

### Argument Separation
```bash
./myapp -v --port=8080 -- --not-a-flag argument
# Everything after -- is treated as non-flag arguments
```

## API Reference

### Package Functions

#### Defining Flags

```go
// String flags
var name = gflag.StringP("name", "n", "default", "description")
var config = gflag.String("config", "/etc/app.conf", "config file path")

// Boolean flags
var verbose = gflag.BoolP("verbose", "v", false, "enable verbose output")
var debug = gflag.Bool("debug", false, "enable debug mode")

// Integer flags
var port = gflag.IntP("port", "p", 8080, "server port")
var workers = gflag.Int("workers", 4, "number of workers")
```

#### Parsing and Accessing Arguments

```go
// Parse command line
gflag.Parse()

// Access non-flag arguments
args := gflag.Args()           // []string of remaining arguments
count := gflag.NArg()          // number of remaining arguments
first := gflag.Arg(0)          // first remaining argument (or "" if none)
```

### FlagSet for Advanced Usage

For more control, you can create your own `FlagSet`:

```go
fs := gflag.NewFlagSet("myapp", gflag.ExitOnError)

verbose := fs.BoolP("verbose", "v", false, "enable verbose output")
port := fs.IntP("port", "p", 8080, "server port")

// Parse custom arguments
err := fs.Parse([]string{"-v", "--port=9000"})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Verbose: %t, Port: %d\n", *verbose, *port)
```

### Error Handling

The `NewFlagSet` function accepts an error handling mode:

```go
// Exit on error (default for CommandLine)
fs := gflag.NewFlagSet("myapp", gflag.ExitOnError)

// Continue on error (return error from Parse)
fs := gflag.NewFlagSet("myapp", gflag.ContinueOnError)

// Panic on error
fs := gflag.NewFlagSet("myapp", gflag.PanicOnError)
```

## Examples

### Web Server

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "github.com/nzions/sharedgolibs/pkg/gflag"
)

func main() {
    port := gflag.IntP("port", "p", 8080, "server port")
    host := gflag.StringP("host", "h", "localhost", "server host")
    verbose := gflag.BoolP("verbose", "v", false, "verbose logging")
    
    gflag.Parse()
    
    addr := fmt.Sprintf("%s:%d", *host, *port)
    
    if *verbose {
        log.Printf("Starting server on %s", addr)
    }
    
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello from %s!", addr)
    })
    
    log.Fatal(http.ListenAndServe(addr, nil))
}
```

Usage:
```bash
./server -p 9000 -h 0.0.0.0 -v
./server --port=9000 --host=0.0.0.0 --verbose
./server -vp 9000 -h 0.0.0.0
```

### File Processing Tool

```go
package main

import (
    "fmt"
    "os"
    "github.com/nzions/sharedgolibs/pkg/gflag"
)

func main() {
    recursive := gflag.BoolP("recursive", "r", false, "process directories recursively")
    output := gflag.StringP("output", "o", "", "output file")
    force := gflag.BoolP("force", "f", false, "force overwrite")
    quiet := gflag.BoolP("quiet", "q", false, "suppress output")
    
    gflag.Parse()
    
    files := gflag.Args()
    if len(files) == 0 {
        fmt.Fprintf(os.Stderr, "Error: no input files specified\n")
        os.Exit(1)
    }
    
    if !*quiet {
        fmt.Printf("Processing %d files\n", len(files))
        if *recursive {
            fmt.Println("Recursive mode enabled")
        }
        if *output != "" {
            fmt.Printf("Output will be written to: %s\n", *output)
        }
    }
    
    // Process files...
}
```

Usage:
```bash
./processor -r -o result.txt file1.txt file2.txt
./processor --recursive --output=result.txt --force file1.txt file2.txt
./processor -rqfo result.txt file1.txt file2.txt
```

## Comparison with Standard Flag Package

| Feature                                | gflag | standard flag |
| -------------------------------------- | ----- | ------------- |
| Short flags (`-v`)                     | ✅     | ✅             |
| Long flags (`--verbose`)               | ✅     | ❌             |
| Combined short flags (`-abc`)          | ✅     | ❌             |
| POSIX-style arguments                  | ✅     | ❌             |
| Flag/value separation (`--flag=value`) | ✅     | ❌             |
| Compatible API                         | ✅     | ✅             |

## Examples

For more comprehensive examples, see the `examples/` directory:

- **`examples/demo/`** - Simple demonstration of all gflag features
- **`examples/file-processor/`** - Full-featured file processing tool showing real-world usage

### Building Examples

```bash
# Build the demo
cd examples/demo && go build .

# Build the file processor  
cd examples/file-processor && go build .
```

### Web Server

```go

## Version

Current version: **0.1.0**

## License

This project is licensed under the CC0-1.0 License.
