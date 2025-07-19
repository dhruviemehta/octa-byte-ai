#!/bin/bash

set -e

# Script to rollback the application
ENVIRONMENT=${1:-staging}
VERSION=${2:-previous}

echo "🔄 Rolling back $ENVIRONMENT environment..."

# Check if required tools are installed
command -v aws >/dev/null 2>&1 || { echo "❌ AWS CLI not installed"; exit 1; }

# Set AWS region
export AWS_DEFAULT_REGION=ap-south-1

echo "📋 Getting current service configuration..."

# Get current task definition
CURRENT_TASK_DEF=$(aws ecs describe-services \
    --cluster $ENVIRONMENT-cluster \
    --services $ENVIRONMENT-go-microservice-service \
    --query 'services[0].taskDefinition' \
    --output text)

echo "Current task definition: $CURRENT_TASK_DEF"

if [ "$VERSION" = "previous" ]; then
    # Get task definition family
    FAMILY=$(echo $CURRENT_TASK_DEF | cut -d'/' -f2 | cut -d':' -f1)
    CURRENT_REVISION=$(echo $CURRENT_TASK_DEF | cut -d':' -f2)
    PREVIOUS_REVISION=$((CURRENT_REVISION - 1))
    
    if [ $PREVIOUS_REVISION -lt 1 ]; then
        echo "❌ No previous revision available"
        exit 1
    fi
    
    ROLLBACK_TASK_DEF="$FAMILY:$PREVIOUS_REVISION"
else
    ROLLBACK_TASK_DEF="$ENVIRONMENT-go-microservice:$VERSION"
fi

echo "🔄 Rolling back to task definition: $ROLLBACK_TASK_DEF"

# Update service to previous task definition
aws ecs update-service \
    --cluster $ENVIRONMENT-cluster \
    --service $ENVIRONMENT-go-microservice-service \
    --task-definition $ROLLBACK_TASK_DEF \
    --region $AWS_DEFAULT_REGION

echo "⏳ Waiting for rollback to complete..."
aws ecs wait services-stable \
    --cluster $ENVIRONMENT-cluster \
    --services $ENVIRONMENT-go-microservice-service \
    --region $AWS_DEFAULT_REGION

echo "✅ Rollback completed successfully!"
