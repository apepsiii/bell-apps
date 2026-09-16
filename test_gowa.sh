#!/bin/bash

# Test Gowa WhatsApp Gateway endpoints
# Run this on VPS to verify correct API format

echo "=== Testing Gowa API Endpoints ==="
echo ""

# Test 1: Send text message (JSON)
echo "1. Testing text message (JSON)..."
curl -s -u "admin:PutihAbu123!" \
  -H "X-Device-Id: Pionir" \
  -H "Content-Type: application/json" \
  -X POST http://127.0.0.1:8053/api/send/message \
  -d '{"phone":"6285158250766","message":"Test text JSON"}' | jq .

echo ""

# Test 2: Send text message (alternative endpoint)
echo "2. Testing text message (alternative endpoint)..."
curl -s -u "admin:PutihAbu123!" \
  -H "X-Device-Id: Pionir" \
  -H "Content-Type: application/json" \
  -X POST http://127.0.0.1:8053/api/whatsapp/send \
  -d '{"phone":"6285158250766","message":"Test text whatsapp"}' | jq .

echo ""

# Test 3: Send image (multipart form)
echo "3. Testing image (multipart form)..."
curl -s -u "admin:PutihAbu123!" \
  -H "X-Device-Id: Pionir" \
  -X POST http://127.0.0.1:8053/api/send/image \
  -F "phone=6285158250766" \
  -F "caption=Test image multipart" \
  -F "image=https://via.placeholder.com/150" | jq .

echo ""

# Test 4: List available endpoints
echo "4. Available devices..."
curl -s -u "admin:PutihAbu123!" \
  http://127.0.0.1:8053/api/devices | jq .

echo ""
echo "=== Test Complete ==="
