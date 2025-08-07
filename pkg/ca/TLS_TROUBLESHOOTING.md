# TLS Certificate Verification Solutions

This document explains how to resolve TLS certificate verification errors like:

```
Error starting backend service: failed to initialize common services: failed to initialize service stripe: failed direct HTTP test to Stripe: Get "https://api.stripe.com/v1/account": tls: failed to verify certificate: x509: certificate signed by unknown authority
```

## The Problem

This error occurs when your Go application tries to make HTTPS requests to external services (like Stripe's API), but the TLS certificate presented by the server isn't trusted by your system's default certificate store. This can happen in several scenarios:

1. **Corporate environments** with custom proxy certificates
2. **Containerized environments** with minimal certificate stores
3. **Development environments** with custom CA setups
4. **Systems with outdated certificate stores**

## Solutions Available in the CA Transport Package

The `pkg/ca` package provides several functions to handle TLS certificate verification:

### 1. `CreateHTTPClientWithSystemAndCustomCAs()` (Recommended)

This is the most flexible solution that allows you to create HTTP clients with different certificate trust configurations:

```go
// For public APIs like Stripe (recommended)
client, err := ca.CreateHTTPClientWithSystemAndCustomCAs(true, false)
if err != nil {
    return fmt.Errorf("failed to create HTTP client: %w", err)
}

// Use this client for your API calls
req, _ := http.NewRequest("GET", "https://api.stripe.com/v1/account", nil)
req.Header.Set("Authorization", "Bearer " + stripeAPIKey)
resp, err := client.Do(req)
```

**Parameters:**
- `includeSystemCAs`: Whether to trust the system's default certificate authorities (recommended: `true`)
- `includeCustomCA`: Whether to fetch and trust additional CA certificates from `SGL_CA` server

### 2. `UpdateTransport()` / `UpdateTransportOnlyIf()` (Global)

These functions modify the global HTTP transport to trust additional CAs:

```go
// Updates global transport if SGL_CA is configured
if err := ca.UpdateTransportOnlyIf(); err != nil {
    log.Printf("Warning: Could not update CA transport: %v", err)
}

// Now http.DefaultClient will trust the custom CA
resp, err := http.DefaultClient.Get("https://internal-service.local/api")
```

## Recommended Solutions by Use Case

### For Public APIs (Stripe, PayPal, etc.)

```go
// Create client with system certificate authorities
client, err := ca.CreateHTTPClientWithSystemAndCustomCAs(true, false)
if err != nil {
    return fmt.Errorf("failed to create HTTP client: %w", err)
}

// Use for all external API calls
stripeResp, err := client.Get("https://api.stripe.com/v1/account")
```

### For Mixed Environments (Public + Internal APIs)

```go
// Create client with both system and custom CAs
client, err := ca.CreateHTTPClientWithSystemAndCustomCAs(true, true)
if err != nil {
    return fmt.Errorf("failed to create HTTP client: %w", err)
}

// Works for both public and internal services
stripeResp, err := client.Get("https://api.stripe.com/v1/account")
internalResp, err := client.Get("https://internal-api.company.local/data")
```

### For Internal-Only Environments

```go
// Create client with only custom CA
client, err := ca.CreateHTTPClientWithSystemAndCustomCAs(false, true)
if err != nil {
    return fmt.Errorf("failed to create HTTP client: %w", err)
}
```

## Environment Variables

The CA transport system uses these environment variables:

- `SGL_CA`: URL of your CA server (e.g., `http://ca-server.local:8090`)
- `SGL_CA_API_KEY`: Optional API key for CA server authentication

## Integration Examples

### Service Initialization

```go
func initializeServices() error {
    // Create HTTP client for external APIs
    httpClient, err := ca.CreateHTTPClientWithSystemAndCustomCAs(true, false)
    if err != nil {
        return fmt.Errorf("failed to create HTTP client: %w", err)
    }

    // Initialize Stripe service with custom client
    stripeService := &StripeService{
        client: httpClient,
        apiKey: os.Getenv("STRIPE_API_KEY"),
    }

    // Test connection
    if err := stripeService.TestConnection(); err != nil {
        return fmt.Errorf("failed to initialize Stripe service: %w", err)
    }

    return nil
}
```

### Service Struct with Custom Client

```go
type StripeService struct {
    client *http.Client
    apiKey string
}

func NewStripeService(apiKey string) (*StripeService, error) {
    // Create client that trusts system CAs
    client, err := ca.CreateHTTPClientWithSystemAndCustomCAs(true, false)
    if err != nil {
        return nil, fmt.Errorf("failed to create HTTP client: %w", err)
    }

    return &StripeService{
        client: client,
        apiKey: apiKey,
    }, nil
}

func (s *StripeService) TestConnection() error {
    req, err := http.NewRequest("GET", "https://api.stripe.com/v1/account", nil)
    if err != nil {
        return err
    }
    req.Header.Set("Authorization", "Bearer "+s.apiKey)

    resp, err := s.client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to connect to Stripe: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status: %s", resp.Status)
    }

    return nil
}
```

## Debugging TLS Issues

### Check Certificate Chain

```bash
# Check what certificates Stripe presents
openssl s_client -connect api.stripe.com:443 -showcerts

# Check system certificate store
openssl version -d
ls -la $(openssl version -d | cut -d'"' -f2)/certs/
```

### Test with Different Configurations

```go
// Test different client configurations
func debugTLSIssues() {
    configs := []struct {
        name       string
        systemCAs  bool
        customCA   bool
    }{
        {"System CAs only", true, false},
        {"Custom CA only", false, true},
        {"Both", true, true},
        {"Neither", false, false},
    }

    for _, config := range configs {
        fmt.Printf("Testing: %s\n", config.name)
        client, err := ca.CreateHTTPClientWithSystemAndCustomCAs(config.systemCAs, config.customCA)
        if err != nil {
            fmt.Printf("  ❌ Failed to create client: %v\n", err)
            continue
        }

        resp, err := client.Get("https://api.stripe.com/v1/account")
        if err != nil {
            fmt.Printf("  ❌ Request failed: %v\n", err)
        } else {
            resp.Body.Close()
            fmt.Printf("  ✅ Success: %s\n", resp.Status)
        }
    }
}
```

## Security Considerations

1. **Always use system CAs for public APIs** - they're maintained and updated regularly
2. **Only add custom CAs when necessary** - for internal services or corporate environments
3. **Never use `InsecureSkipVerify: true`** in production - it disables all certificate validation
4. **Keep certificate stores updated** - especially in containerized environments

## Troubleshooting Common Issues

### "certificate signed by unknown authority"
- **Solution**: Use `CreateHTTPClientWithSystemAndCustomCAs(true, false)` for public APIs
- **Cause**: System certificate store doesn't trust the certificate's CA

### "cannot validate certificate for [hostname] because it doesn't contain any IP SANs"
- **Solution**: Check that you're using the correct hostname/IP
- **Cause**: Certificate doesn't include the hostname you're connecting to

### "tls: first record does not look like a TLS handshake"
- **Solution**: Ensure you're using HTTPS, not HTTP
- **Cause**: Trying to do TLS with a non-TLS server

### Connection works locally but fails in containers
- **Solution**: Ensure container has updated CA certificates
- **Cause**: Container base image has outdated certificate store
