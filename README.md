# Traefik Webhook Endpoint Redirector

A Traefik middleware plugin that forwards webhook requests to different backends based on configurable JSON fields in the webhook payload.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/innolabsdev/traefik-payload-router)](https://goreportcard.com/report/github.com/innolabsdev/traefik-payload-router)

## Overview

This plugin allows you to route webhook requests to different backend services based on any field in the JSON payload. It's perfect for multi-tenant applications, client-specific routing, or any scenario where you need to dynamically route webhooks based on their content.

### Key Features

- 🎯 **Dynamic Field Routing** - Route based on any JSON field in your webhook payload
- 🔀 **Flexible Mapping** - Map field values to specific backend URLs
- 🏠 **Default Fallback** - Configure a default backend for unmapped values
- 🛡️ **Request Preservation** - Maintains original headers, body, and query parameters
- 📝 **Configurable Path** - Specify which path should be processed (default: `/webhooks`)
- ⚡ **High Performance** - Minimal overhead with efficient JSON parsing
- 🔧 **Easy Configuration** - Simple Traefik labels or file configuration

## Quick Start

### Installation

Add the plugin to your Traefik configuration:

#### Static Configuration (traefik.yml)
```yaml
experimental:
  plugins:
    webhook-redirect:
      moduleName: github.com/innolabsdev/traefik-payload-router
      version: v1.0.0
```

#### Docker Compose Labels
```yaml
services:
  traefik:
    image: traefik:v3.0
    command:
      - "--experimental.plugins.webhook-redirect.moduleName=github.com/innolabsdev/traefik-payload-router"
      - "--experimental.plugins.webhook-redirect.version=v1.0.0"

  webhook-service:
    image: your-webhook-service
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.webhooks.rule=PathPrefix(`/webhooks`)"
      - "traefik.http.routers.webhooks.middlewares=webhook-redirect"
      
      # Plugin configuration
      - "traefik.http.middlewares.webhook-redirect.plugin.webhook-redirect.fieldName=client_id"
      - "traefik.http.middlewares.webhook-redirect.plugin.webhook-redirect.redirectMappings.foo=http://foo-service/webhooks"
      - "traefik.http.middlewares.webhook-redirect.plugin.webhook-redirect.redirectMappings.bar=http://bar-service/webhooks"
      - "traefik.http.middlewares.webhook-redirect.plugin.webhook-redirect.defaultRedirect=http://default-service/webhooks"
```

## Configuration

### Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `fieldName` | string | `"endpoint_id"` | JSON field name to use for routing decisions |
| `redirectMappings` | map[string]string | `{}` | Map of field values to target URLs |
| `defaultRedirect` | string | `""` | Default URL when no mapping matches |
| `webhookPath` | string | `"/webhooks"` | Path prefix to process |
| `statusCode` | int | `302` | HTTP status code for redirects |

### Example Configurations

#### Client-based Routing
```yaml
# Route based on client_id field
- "traefik.http.middlewares.webhook-redirect.plugin.webhook-redirect.fieldName=client_id"
- "traefik.http.middlewares.webhook-redirect.plugin.webhook-redirect.redirectMappings.foo=http://foo-backend:8080/webhooks"
- "traefik.http.middlewares.webhook-redirect.plugin.webhook-redirect.redirectMappings.bar=http://bar-backend:8080/webhooks"
```

#### Tenant-based Routing
```yaml
# Route based on tenant_id field
- "traefik.http.middlewares.webhook-redirect.plugin.webhook-redirect.fieldName=tenant_id"
- "traefik.http.middlewares.webhook-redirect.plugin.webhook-redirect.redirectMappings.tenant-foo=http://foo-tenant-service/api/webhooks"
- "traefik.http.middlewares.webhook-redirect.plugin.webhook-redirect.redirectMappings.tenant-bar=http://bar-tenant-service/api/webhooks"
```

#### Environment-based Routing
```yaml
# Route based on environment field
- "traefik.http.middlewares.webhook-redirect.plugin.webhook-redirect.fieldName=environment"
- "traefik.http.middlewares.webhook-redirect.plugin.webhook-redirect.redirectMappings.prod=http://prod-webhook-handler/webhooks"
- "traefik.http.middlewares.webhook-redirect.plugin.webhook-redirect.redirectMappings.staging=http://staging-webhook-handler/webhooks"
- "traefik.http.middlewares.webhook-redirect.plugin.webhook-redirect.defaultRedirect=http://dev-webhook-handler/webhooks"
```

## Usage Examples

### Example Webhook Payloads

#### Client-based routing payload:
```json
{
  "event": "order.created",
  "client_id": "foo",
  "data": {
    "order_id": "12345",
    "amount": 99.99
  }
}
```

#### Tenant-based routing payload:
```json
{
  "type": "user.signup",
  "tenant_id": "tenant-bar", 
  "user": {
    "id": "user-456",
    "email": "user@example.com"
  }
}
```

### Request Flow

1. **Incoming Request**: POST to `/webhooks` with JSON payload
2. **Field Extraction**: Plugin reads the configured `fieldName` from JSON
3. **Route Matching**: Looks up the field value in `redirectMappings`
4. **Request Forwarding**: Forwards the complete request to the matched backend
5. **Response Relay**: Returns the backend's response to the original client

## Local Development

### Prerequisites
- Docker and Docker Compose
- Go 1.21+

### Setup
```bash
# Clone the repository
git clone https://github.com/innolabsdev/traefik-payload-router.git
cd traefik-payload-router

# Set up local development environment  
./setup-local.sh

# Start test environment
docker-compose up -d

# Run tests
./test-webhook-local.sh
```

### Testing Different Configurations

Test with different field names:
```bash
# Test with client_id field
curl -X POST http://localhost:8080/webhooks \
  -H "Content-Type: application/json" \
  -d '{"event": "test", "client_id": "foo"}'

# Test with tenant_id field  
curl -X POST http://localhost:8080/webhooks \
  -H "Content-Type: application/json" \
  -d '{"event": "test", "tenant_id": "bar"}'
```

## Advanced Configuration

### File Configuration
```yaml
# traefik.yml
http:
  middlewares:
    webhook-redirect:
      plugin:
        webhook-redirect:
          fieldName: "client_id"
          redirectMappings:
            "foo": "http://foo-backend:8080/webhooks"
            "bar": "http://bar-backend:8080/webhooks" 
          defaultRedirect: "http://default-backend:8080/webhooks"
          webhookPath: "/api/webhooks"
          statusCode: 307
```

### Multiple Middleware Instances
```yaml
# Different configurations for different routes
services:
  webhook-service:
    labels:
      # Client routing middleware
      - "traefik.http.middlewares.client-router.plugin.webhook-redirect.fieldName=client_id"
      - "traefik.http.middlewares.client-router.plugin.webhook-redirect.redirectMappings.foo=http://foo-service/hooks"
      
      # Tenant routing middleware  
      - "traefik.http.middlewares.tenant-router.plugin.webhook-redirect.fieldName=tenant_id"
      - "traefik.http.middlewares.tenant-router.plugin.webhook-redirect.redirectMappings.tenant1=http://tenant1-service/hooks"
      
      # Apply different middleware to different paths
      - "traefik.http.routers.client-webhooks.rule=PathPrefix(`/client-webhooks`)"
      - "traefik.http.routers.client-webhooks.middlewares=client-router"
      - "traefik.http.routers.tenant-webhooks.rule=PathPrefix(`/tenant-webhooks`)"  
      - "traefik.http.routers.tenant-webhooks.middlewares=tenant-router"
```

## Error Handling

The plugin handles various error scenarios gracefully:

- **Invalid JSON**: Passes request through to next handler
- **Missing Field**: Passes request through to next handler  
- **No Mapping Found**: Uses `defaultRedirect` if configured, otherwise passes through
- **Invalid URLs**: Returns 500 Internal Server Error
- **Backend Unreachable**: Returns 502 Bad Gateway

## Performance Considerations

- **Memory Usage**: Minimal - only stores configuration mappings
- **CPU Usage**: Low - efficient JSON parsing with early exits
- **Latency**: ~1-2ms additional latency for JSON parsing and field extraction
- **Concurrency**: Fully concurrent, no shared state between requests

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices
- Add tests for new functionality
- Update documentation for configuration changes
- Test with multiple Traefik versions

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- 📚 [Documentation](https://github.com/innolabsdev/traefik-payload-router/wiki)
- 🐛 [Issue Tracker](https://github.com/innolabsdev/traefik-payload-router/issues)
- 💬 [Discussions](https://github.com/innolabsdev/traefik-payload-router/discussions)

## Changelog

### v1.0.0
- Initial release
- Configurable field name routing
- Dynamic JSON payload parsing
- Comprehensive test suite
- Docker Compose examples

---

Made with ❤️ by [Innolabs](https://innolabs.dev)
