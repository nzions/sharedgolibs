# Shared Go Libraries

Common Go packages shared between allmytails and googleemu projects.

## License

This work is dedicated to the public domain under [CC0 1.0 Universal](LICENSE).

## Packages

### pkg/util
Environment variable utilities and common helper functions.

### pkg/middleware  
HTTP middleware for CORS, logging, and request handling.

### pkg/version
Version information and utilities.

## Usage

```go
import (
    "github.com/nzions/sharedgolibs/pkg/util"
    "github.com/nzions/sharedgolibs/pkg/middleware"
    "github.com/nzions/sharedgolibs/pkg/version"
)
```

## Installation

```bash
go get github.com/nzions/sharedgolibs
```
