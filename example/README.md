# Production Example: Traefik Payload Router Plugin

This example demonstrates how to use the Traefik Payload Router plugin in a production environment using the official plugin from the Traefik Plugin Catalog.

## Features

- **Production Plugin**: Uses the official plugin from Traefik Plugin Catalog (not local development version)
- **Customer Routing**: Routes webhooks to different customer backends based on `endpoint_id` in JSON payload
- **Fallback Handling**: Non-webhook requests are served by nginx with dynamic responses
- **Real-world Scenarios**: Includes realistic webhook payloads (payments, orders, user signups)

## Quick Start

1. **Start the services:**
   ```bash
   docker-compose up -d
   ```

2. **Run the tests:**
   ```bash
   ./test-production.sh
   ```

3. **Access Traefik dashboard:**
   ```
   http://localhost:8081
   ```

## How It Works

### Plugin Configuration
- **Field Name**: `endpoint_id` (field in JSON payload used for routing)
- **Webhook Path**: `/webhook` (only POST requests to this path are processed)
- **Customer Mappings**:
  - `customer1` → `backend-customer1`
  - `customer2` → `backend-customer2` 
  - `customer3` → `backend-customer3`
- **Default Route**: Unknown `endpoint_id` values route to `default-webhook`

### Request Flow
1. **Webhook POST** to `/webhook` with JSON payload → Plugin processes and routes based on `endpoint_id`
2. **Other requests** (GET, other paths, invalid JSON) → Pass through to nginx main site

### Test Scenarios
- ✅ Customer-specific webhook routing (3 different customers)
- ✅ Default fallback for unknown customers
- ✅ Main site functionality (API endpoints, dashboard)
- ✅ Non-POST webhook requests pass through normally
- ✅ Invalid JSON passes through to main site

## Plugin Details

- **Plugin Name**: `payload-router`
- **Source**: [github.com/innolabsdev/traefik-payload-router](https://github.com/innolabsdev/traefik-payload-router)
- **Catalog**: [Traefik Plugin Catalog](https://plugins.traefik.io/plugins/innolabsdev/traefik-payload-router)
- **Version**: v1.0.0

## Production Considerations

- Update plugin version in `docker-compose.yml` as needed
- Configure real backend URLs instead of echo servers
- Add proper error handling and monitoring
- Consider using environment variables for sensitive configuration
- Set up proper logging and observability

## Cleanup

```bash
docker-compose down
```