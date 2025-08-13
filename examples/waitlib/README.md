# waitlib Example

This example demonstrates how to use the waitlib package in your own applications.

## Running the Example

```bash
# From the repository root
go run examples/waitlib/main.go
```

Or build and run:

```bash
go build -o waitlib-example examples/waitlib/main.go
./waitlib-example
```

## Command Line Options

```bash
# Show help
./waitlib-example --help

# Show version
./waitlib-example --version

# Run normally (waits indefinitely)
./waitlib-example
```

## Docker Usage Example

Create a simple Dockerfile:

```dockerfile
FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o waitlib-example examples/waitlib/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/waitlib-example .
CMD ["./waitlib-example"]
```

Build and run:

```bash
docker build -t waitlib-example .
docker run -d --name my-wait-container waitlib-example
```

Check the process name in `docker ps`:
```bash
docker ps
# You should see: wait example-v1.2.3 <uptime>
```

## Integration in Your Applications

```go
package main

import "github.com/nzions/sharedgolibs/pkg/waitlib"

func main() {
    // Use your application's version
    waitlib.Run("my-app-v2.1.0")
}
```

This is particularly useful for:
- Health check containers
- Init containers that need to wait for dependencies
- Long-running background services
- Container orchestration scenarios where you need to easily identify versions and uptime
