#!/bin/bash

echo "Testing Production Payload Router Plugin on port 8080"
echo "===================================================="

# Test customer endpoints
echo -e "\n1. Testing endpoint_id 'customer1' (should route to Customer-1):"
response=$(curl -s -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event": "payment.received",
    "data": {"amount": 100, "currency": "USD"},
    "endpoint_id": "customer1"
  }')
echo "Backend response: $(echo "$response" | grep -o 'Customer-[^"]*' | head -1 || echo 'No customer header found')"
echo "Status: $(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/webhook -H "Content-Type: application/json" -d '{"event": "payment.received", "endpoint_id": "customer1"}')"

echo -e "\n2. Testing endpoint_id 'customer2' (should route to Customer-2):"
response=$(curl -s -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event": "order.created", 
    "data": {"order_id": "12345", "total": 250},
    "endpoint_id": "customer2"
  }')
echo "Backend response: $(echo "$response" | grep -o 'Customer-[^"]*' | head -1 || echo 'No customer header found')"
echo "Status: $(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/webhook -H "Content-Type: application/json" -d '{"event": "order.created", "endpoint_id": "customer2"}')"

echo -e "\n3. Testing endpoint_id 'customer3' (should route to Customer-3):"
response=$(curl -s -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event": "user.signup",
    "data": {"user_id": "user789", "email": "test@example.com"},
    "endpoint_id": "customer3"
  }')
echo "Backend response: $(echo "$response" | grep -o 'Customer-[^"]*' | head -1 || echo 'No customer header found')"
echo "Status: $(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/webhook -H "Content-Type: application/json" -d '{"event": "user.signup", "endpoint_id": "customer3"}')"

echo -e "\n4. Testing unknown endpoint_id (should route to Default):"
response=$(curl -s -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event": "test.event",
    "data": {"message": "unknown customer"}, 
    "endpoint_id": "unknown_customer"
  }')
echo "Backend response: $(echo "$response" | grep -o 'Customer-[^"]*\|Default' | head -1 || echo 'No backend header found')"
echo "Status: $(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/webhook -H "Content-Type: application/json" -d '{"event": "test.event", "endpoint_id": "unknown_customer"}')"

echo -e "\n5. Testing main site path (should work normally):"
response=$(curl -s -X GET http://localhost:8080/api/users)
echo "Response: $response"
echo "Status: $(curl -s -o /dev/null -w "%{http_code}" -X GET http://localhost:8080/api/users)"

echo -e "\n6. Testing another site path (should work normally):"
response=$(curl -s -X GET http://localhost:8080/dashboard)
echo "Response: $response"
echo "Status: $(curl -s -o /dev/null -w "%{http_code}" -X GET http://localhost:8080/dashboard)"

echo -e "\n7. Testing GET request to webhook path (should pass through to main site):"
response=$(curl -s -X GET http://localhost:8080/webhook)
echo "Response: $response"
echo "Status: $(curl -s -o /dev/null -w "%{http_code}" -X GET http://localhost:8080/webhook)"

echo -e "\n8. Testing invalid JSON to webhook path (should pass through to main site):"
response=$(curl -s -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d 'invalid json')
echo "Response: $response"
echo "Status: $(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/webhook -H "Content-Type: application/json" -d 'invalid json')"

echo -e "\nTesting complete!"
echo "Traefik dashboard: http://localhost:8081"
echo "Plugin documentation: https://plugins.traefik.io/plugins/innolabsdev/traefik-payload-router"