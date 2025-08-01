# Testicle Container Deployment Guide

## 🐳 Container Architecture

### Design Philosophy
Testicle is designed to work seamlessly in both local development and containerized environments. When running in a container, tests are mounted at `/tests` and Testicle automatically detects the container environment.

### Container Detection
```go
// Automatic container detection
func DetectEnvironment() (*Environment, error) {
    env := &Environment{}
    
    // Check for container indicators
    if isInContainer() {
        env.IsContainer = true
        env.TestDir = "/tests"
        env.ConfigPath = "/config/testicle.yaml"
    } else {
        env.IsContainer = false
        env.TestDir = "."
        env.ConfigPath = "./testicle.yaml"
    }
    
    return env, nil
}

func isInContainer() bool {
    // Check multiple container indicators
    indicators := []string{
        "/.dockerenv",                    // Docker
        "/run/.containerenv",             // Podman
        "/proc/1/cgroup",                 // Container cgroups
    }
    
    for _, indicator := range indicators {
        if _, err := os.Stat(indicator); err == nil {
            return true
        }
    }
    
    // Check environment variables
    if os.Getenv("CONTAINER") != "" || 
       os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
        return true
    }
    
    return false
}
```

## 📦 Container Images

### Base Image Strategy

```dockerfile
# Multi-stage build for minimal runtime image
FROM golang:1.23-alpine AS builder

# Build environment
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o testicle cmd/testicle/main.go

# Runtime image
FROM golang:1.23-alpine AS runtime

# Install runtime dependencies
RUN apk --no-cache add \
    git \
    ca-certificates \
    tzdata

# Create non-root user
RUN addgroup -g 1000 testicle && \
    adduser -D -s /bin/sh -u 1000 -G testicle testicle

# Set up directories
WORKDIR /app
RUN mkdir -p /tests /config /output && \
    chown -R testicle:testicle /app /tests /config /output

# Copy binary
COPY --from=builder /build/testicle /usr/local/bin/testicle
RUN chmod +x /usr/local/bin/testicle

# Set up volumes
VOLUME ["/tests", "/config", "/output"]

# Switch to non-root user
USER testicle

# Default configuration
ENV TESTICLE_TEST_DIR=/tests
ENV TESTICLE_CONFIG_FILE=/config/testicle.yaml
ENV TESTICLE_OUTPUT_DIR=/output

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD testicle --version || exit 1

# Default command
ENTRYPOINT ["testicle"]
CMD ["--dir", "/tests", "--config", "/config/testicle.yaml", "--output", "/output/results.json"]
```

### Lightweight Image (Alpine-based)

```dockerfile
# Ultra-minimal runtime image
FROM alpine:3.18 AS minimal

# Install Go runtime (smaller than full golang image)
RUN apk --no-cache add \
    go \
    git \
    ca-certificates

# Copy binary from builder stage
COPY --from=builder /build/testicle /usr/local/bin/testicle

# Minimal setup
WORKDIR /app
VOLUME ["/tests"]

ENTRYPOINT ["testicle"]
CMD ["--dir", "/tests"]
```

## 🚀 Deployment Patterns

### 1. Local Development with Docker

```bash
# Basic usage - mount current directory
docker run --rm -v $(pwd):/tests testicle:latest

# With custom configuration
docker run --rm \
  -v $(pwd):/tests \
  -v $(pwd)/testicle.yaml:/config/testicle.yaml \
  testicle:latest

# With output persistence
docker run --rm \
  -v $(pwd):/tests \
  -v $(pwd)/reports:/output \
  testicle:latest --output /output/test-results.json

# Interactive mode with watch
docker run --rm -it \
  -v $(pwd):/tests \
  testicle:latest --watch
```

### 2. CI/CD Pipeline Integration

```yaml
# GitHub Actions
name: Test with Testicle
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Run Testicle Tests
        run: |
          docker run --rm \
            -v ${{ github.workspace }}:/tests \
            -v ${{ github.workspace }}/reports:/output \
            testicle:latest \
            --output /output/results.json \
            --format junit
      
      - name: Upload Test Results
        uses: actions/upload-artifact@v3
        with:
          name: test-results
          path: reports/
```

```yaml
# GitLab CI
test:
  image: testicle:latest
  script:
    - testicle --dir /builds/$CI_PROJECT_PATH --output results.json
  artifacts:
    reports:
      junit: results.xml
    paths:
      - results.json
    expire_in: 1 week
```

### 3. Kubernetes Deployment

```yaml
# ConfigMap for Testicle configuration
apiVersion: v1
kind: ConfigMap
metadata:
  name: testicle-config
data:
  testicle.yaml: |
    test_directory: "/tests"
    watch_enabled: false  # Disable in K8s
    timeout:
      default: "5m"
    output:
      file: "/output/results.json"
      format: "json"
    ui:
      colors: false  # Disable colors in CI
---
# Job for running tests
apiVersion: batch/v1
kind: Job
metadata:
  name: testicle-test-job
spec:
  template:
    spec:
      containers:
      - name: testicle
        image: testicle:latest
        command: ["testicle"]
        args: ["--config", "/config/testicle.yaml"]
        volumeMounts:
        - name: test-source
          mountPath: /tests
          readOnly: true
        - name: config
          mountPath: /config
          readOnly: true
        - name: output
          mountPath: /output
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: test-source
        configMap:
          name: test-source-code
      - name: config
        configMap:
          name: testicle-config
      - name: output
        emptyDir: {}
      restartPolicy: Never
```

### 4. Docker Compose for Development

```yaml
# docker-compose.yml
version: '3.8'

services:
  testicle:
    build: .
    volumes:
      - .:/tests:ro
      - ./config:/config:ro
      - ./reports:/output
    environment:
      - TESTICLE_WATCH=true
      - TESTICLE_PARALLEL=4
    stdin_open: true
    tty: true

  # Optional: database for integration tests
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: test_db
      POSTGRES_USER: test_user
      POSTGRES_PASSWORD: test_pass
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  # Optional: Redis for cache tests
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

## 🔧 Configuration for Containers

### Container-Optimized Configuration

```yaml
# testicle-container.yaml
test_directory: "/tests"
watch_enabled: false  # Usually disabled in CI
timeout:
  default: "10m"  # Longer timeouts for CI
  per_test:
    integration_test: "15m"
output:
  file: "/output/results.json"
  format: "json"
  include_output: true
  include_timing: true
discovery:
  patterns:
    - "**/*_test.go"
  exclude:
    - "**/vendor/**"
    - "**/.git/**"
ui:
  colors: false  # Disable colors in CI
  progress: "dots"  # Minimal progress for logs
container:
  mode: true
  mount_point: "/tests"
  output_dir: "/output"
  config_dir: "/config"
```

### Environment Variable Configuration

```bash
# Container environment variables
TESTICLE_TEST_DIR=/tests
TESTICLE_CONFIG_FILE=/config/testicle.yaml
TESTICLE_OUTPUT_FILE=/output/results.json
TESTICLE_WATCH=false
TESTICLE_PARALLEL=4
TESTICLE_TIMEOUT=10m
TESTICLE_FORMAT=json
TESTICLE_VERBOSE=false
TESTICLE_COLORS=false
```

## 📊 Monitoring and Observability

### Health Checks

```go
// Health check endpoint for containers
func (a *App) HealthCheck() error {
    // Check if Go binary is available
    if _, err := exec.LookPath("go"); err != nil {
        return fmt.Errorf("go binary not found: %w", err)
    }
    
    // Check if test directory is accessible
    if _, err := os.Stat(a.config.TestDir); err != nil {
        return fmt.Errorf("test directory not accessible: %w", err)
    }
    
    // Check if output directory is writable
    testFile := filepath.Join(a.config.OutputDir, ".health")
    if err := os.WriteFile(testFile, []byte("ok"), 0644); err != nil {
        return fmt.Errorf("output directory not writable: %w", err)
    }
    os.Remove(testFile)
    
    return nil
}
```

### Logging for Containers

```go
// Structured logging for container environments
type ContainerLogger struct {
    logger *slog.Logger
    format string  // "json" or "text"
}

func (l *ContainerLogger) LogTestResult(result *TestResult) {
    l.logger.Info("test_completed",
        slog.String("test_name", result.Test.Name),
        slog.String("package", result.Test.Package),
        slog.String("status", string(result.Status)),
        slog.Duration("duration", result.Duration),
        slog.Bool("timed_out", result.TimedOut),
    )
}
```

### Metrics Collection

```go
// Prometheus metrics for container monitoring
var (
    testsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "testicle_tests_total",
            Help: "Total number of tests executed",
        },
        []string{"status", "package"},
    )
    
    testDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "testicle_test_duration_seconds",
            Help: "Test execution duration in seconds",
        },
        []string{"test_name", "package"},
    )
)
```

## 🔐 Security Considerations

### Container Security

```dockerfile
# Security-hardened container
FROM golang:1.23-alpine AS runtime

# Create non-root user with specific UID/GID
RUN addgroup -g 10001 testicle && \
    adduser -D -s /bin/sh -u 10001 -G testicle testicle

# Install minimal required packages
RUN apk --no-cache add --update \
    ca-certificates \
    git \
    && rm -rf /var/cache/apk/*

# Set up secure directories with proper permissions
WORKDIR /app
RUN mkdir -p /tests /config /output && \
    chown -R testicle:testicle /app /tests /config /output && \
    chmod 755 /tests /config /output

# Copy binary and set permissions
COPY --from=builder --chown=testicle:testicle /build/testicle /usr/local/bin/testicle
RUN chmod 555 /usr/local/bin/testicle

# Drop privileges
USER 10001:10001

# Security labels
LABEL security.non-root=true
LABEL security.minimal=true
```

### Runtime Security

```yaml
# Kubernetes security context
apiVersion: v1
kind: Pod
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 10001
    runAsGroup: 10001
    fsGroup: 10001
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: testicle
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
        - ALL
```

This container deployment guide provides comprehensive coverage for running Testicle in containerized environments, from local development to production CI/CD pipelines.
