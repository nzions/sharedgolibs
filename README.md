# Shared Go Libraries

Common Go packages shared between allmytails and googleemu projects.

## Packages

### pkg/util
Environment variable utilities and common helper functions.

### pkg/middleware  
HTTP middleware for CORS, logging, and request handling.

## Usage

```go
import (
    "github.com/nzions/sharedgolibs/pkg/util"
    "github.com/nzions/sharedgolibs/pkg/middleware"
)
```

## Installation

```bash
go get github.com/nzions/sharedgolibs
```
