# CA Examples

This directory contains example applications demonstrating various features of the Certificate Authority package.

## Examples

### 1. Persistence Example (`persistence/main.go`)

Demonstrates how to configure and use the CA with disk persistence:

- Shows configuration with `PersistDir` for disk storage
- Demonstrates certificate persistence across CA restarts
- Compares RAM-only vs disk-based storage
- Shows how certificates are automatically loaded from disk on startup

**Usage:**
```bash
cd examples/persistence
go run main.go
```

**Features demonstrated:**
- Disk persistence configuration
- Certificate storage and retrieval
- Automatic loading of existing certificates
- Comparison between storage modes

### 2. Thread Safety Example (`threadsafe/main.go`)

Comprehensive test of thread safety and concurrent operations:

- Issues 50 certificates concurrently from 10 goroutines
- Tests concurrent read/write operations
- Verifies data integrity under concurrent load
- Demonstrates proper thread-safe operations

**Usage:**
```bash
cd examples/threadsafe
go run main.go
```

**Features demonstrated:**
- Concurrent certificate generation
- Thread-safe storage operations
- Data integrity verification
- Read/write concurrency testing
- Performance measurement under load

## Running All Examples

To run all examples sequentially:

```bash
# From the pkg/ca directory
cd examples/persistence && go run main.go && cd ../threadsafe && go run main.go
```

## Key Features Demonstrated

Both examples showcase:
- **Storage Abstraction**: Seamless switching between RAM and disk storage
- **Thread Safety**: All operations are protected with proper locking
- **Persistence**: Certificates survive CA restarts when using disk storage
- **Performance**: Efficient concurrent operations with minimal blocking
- **Error Handling**: Proper error handling and validation
- **Configuration**: Simple configuration options for different use cases

## Implementation Notes

The examples use temporary directories for disk storage to avoid cluttering the file system. In production, you would typically use a dedicated directory for certificate storage.

The thread safety example uses aggressive concurrency testing to verify the implementation can handle high-load scenarios without data corruption or race conditions.
