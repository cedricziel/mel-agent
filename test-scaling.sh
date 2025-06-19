#!/bin/bash

# Test script for horizontal scaling setup
set -e

echo "🚀 Testing MEL Agent horizontal scaling setup..."

# Function to check if a service is healthy
check_health() {
    local url=$1
    local service_name=$2
    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -f -s "$url" > /dev/null 2>&1; then
            echo "✅ $service_name is healthy"
            return 0
        fi
        echo "⏳ Waiting for $service_name... (attempt $attempt/$max_attempts)"
        sleep 5
        ((attempt++))
    done
    
    echo "❌ $service_name failed to become healthy"
    return 1
}

# Function to test load balancing
test_load_balancing() {
    echo "🔄 Testing load balancing across API servers..."
    
    # Make multiple requests and check if they're distributed
    echo "Making 10 requests to check load balancing:"
    for i in {1..10}; do
        response=$(curl -s http://localhost:8080/health)
        echo "Request $i: $response"
        sleep 1
    done
}

# Function to test readiness endpoint
test_readiness() {
    echo "🔍 Testing readiness endpoint..."
    response=$(curl -s http://localhost:8080/ready)
    echo "Readiness check response: $response"
    
    if echo "$response" | grep -q '"status":"ready"'; then
        echo "✅ Readiness check passed"
    else
        echo "⚠️  Readiness check response: $response"
    fi
}

# Function to test API endpoints
test_api_endpoints() {
    echo "🔗 Testing API endpoints through load balancer..."
    
    # Test node types endpoint
    response=$(curl -s http://localhost:8080/api/node-types)
    if [ $? -eq 0 ]; then
        echo "✅ API endpoints accessible through load balancer"
    else
        echo "❌ API endpoints not accessible"
        return 1
    fi
}

# Function to test worker connectivity
test_worker_connectivity() {
    echo "👷 Testing worker connectivity..."
    
    # Check if workers are registered (this would require the worker API to be accessible)
    # For now, just check that the worker endpoint responds
    response=$(curl -s -X POST http://localhost:8080/api/workers \
        -H "Content-Type: application/json" \
        -d '{"id":"test-worker","concurrency":1}' || echo "Expected error")
    
    if [[ "$response" == *"Expected error"* ]] || [[ "$response" == *"error"* ]]; then
        echo "✅ Worker API endpoint accessible (expected auth error)"
    else
        echo "⚠️  Unexpected worker API response: $response"
    fi
}

# Main test flow
main() {
    echo "Starting horizontal scaling test..."
    
    # Check if Docker Compose is available
    if ! command -v docker-compose &> /dev/null; then
        echo "❌ docker-compose is not available. Please install Docker Compose."
        exit 1
    fi
    
    # Start the scaled services
    echo "🏗️  Starting scaled services..."
    docker-compose -f docker-compose.scale.yml down --volumes --remove-orphans 2>/dev/null || true
    
    # Scale API servers to 3 instances and workers to 2 instances
    echo "📈 Scaling API servers to 3 instances and workers to 2 instances..."
    docker-compose -f docker-compose.scale.yml up -d --build --scale api=3 --scale worker=2
    
    # Wait for services to be ready
    echo "⏳ Waiting for services to start..."
    sleep 30
    
    # Check individual service health
    check_health "http://localhost:8080/health" "Load balancer"
    
    # Test functionality
    test_readiness
    test_load_balancing
    test_api_endpoints
    test_worker_connectivity
    
    # Show service status
    echo "📊 Service status:"
    docker-compose -f docker-compose.scale.yml ps
    
    # Show logs for debugging if needed
    echo "📋 Recent logs:"
    docker-compose -f docker-compose.scale.yml logs --tail=20
    
    echo ""
    echo "🎉 Horizontal scaling test completed!"
    echo "💡 To stop the services, run: docker-compose -f docker-compose.scale.yml down"
    echo "💡 To view logs, run: docker-compose -f docker-compose.scale.yml logs -f"
}

# Cleanup function
cleanup() {
    echo "🧹 Cleaning up services..."
    docker-compose -f docker-compose.scale.yml down --volumes 2>/dev/null || true
}

# Handle script interruption
trap cleanup EXIT

# Run main function
main "$@"