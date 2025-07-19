#!/bin/bash

# Script to perform health checks
ENVIRONMENT=${1:-staging}
MAX_ATTEMPTS=${2:-30}
SLEEP_INTERVAL=${3:-10}

echo "🏥 Performing health checks for $ENVIRONMENT environment..."

# Set AWS region
export AWS_DEFAULT_REGION=ap-south-1

# Get ALB DNS name
ALB_DNS=$(aws elbv2 describe-load-balancers \
    --names $ENVIRONMENT-alb \
    --query 'LoadBalancers[0].DNSName' \
    --output text 2>/dev/null)

if [ "$ALB_DNS" = "None" ] || [ -z "$ALB_DNS" ]; then
    echo "❌ Could not find load balancer for $ENVIRONMENT"
    exit 1
fi

BASE_URL="http://$ALB_DNS"
echo "🌐 Testing URL: $BASE_URL"

# Function to test endpoint
test_endpoint() {
    local endpoint=$1
    local expected_status=$2
    
    echo "Testing $endpoint..."
    
    for i in $(seq 1 $MAX_ATTEMPTS); do
        response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL$endpoint" 2>/dev/null)
        
        if [ "$response" = "$expected_status" ]; then
            echo "✅ $endpoint responded with $response"
            return 0
        else
            echo "⏳ Attempt $i/$MAX_ATTEMPTS: $endpoint responded with $response (expected $expected_status)"
            if [ $i -lt $MAX_ATTEMPTS ]; then
                sleep $SLEEP_INTERVAL
            fi
        fi
    done
    
    echo "❌ $endpoint failed health check after $MAX_ATTEMPTS attempts"
    return 1
}

# Test health endpoint
test_endpoint "/health" "200" || exit 1

# Test readiness endpoint
test_endpoint "/ready" "200" || exit 1

# Test API endpoint
test_endpoint "/api/users" "200" || exit 1

# Test metrics endpoint
test_endpoint "/metrics" "200" || exit 1

echo "🎉 All health checks passed!"

# Get some basic metrics
echo "📊 Getting basic application metrics..."

HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")
echo "Health check response: $HEALTH_RESPONSE"

echo "✅ Health check completed successfully!"

