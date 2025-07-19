#!/bin/bash

set -e

# Script to deploy the application
ENVIRONMENT=${1:-staging}
VERSION=${2:-latest}

echo "🚀 Deploying to $ENVIRONMENT environment..."

# Check if required tools are installed
command -v aws >/dev/null 2>&1 || { echo "❌ AWS CLI not installed"; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "❌ Docker not installed"; exit 1; }

# Set AWS region
export AWS_DEFAULT_REGION=ap-south-1

# Get AWS account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
ECR_REPOSITORY="$ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com/$ENVIRONMENT-go-microservice"

echo "📦 Building Docker image..."
cd application

# Build the Docker image
docker build -t go-microservice:$VERSION .

# Tag for ECR
docker tag go-microservice:$VERSION $ECR_REPOSITORY:$VERSION
docker tag go-microservice:$VERSION $ECR_REPOSITORY:latest

echo "🔐 Logging into ECR..."
aws ecr get-login-password --region $AWS_DEFAULT_REGION | docker login --username AWS --password-stdin $ECR_REPOSITORY

# Create ECR repository if it doesn't exist
aws ecr describe-repositories --repository-names go-microservice 2>/dev/null || \
aws ecr create-repository --repository-name go-microservice

echo "⬆️ Pushing image to ECR..."
docker push $ECR_REPOSITORY:$VERSION
docker push $ECR_REPOSITORY:latest

echo "🔄 Updating ECS service..."
aws ecs update-service \
    --cluster $ENVIRONMENT-cluster \
    --service $ENVIRONMENT-go-microservice-service \
    --force-new-deployment \
    --region $AWS_DEFAULT_REGION

echo "⏳ Waiting for deployment to stabilize..."
aws ecs wait services-stable \
    --cluster $ENVIRONMENT-cluster \
    --services $ENVIRONMENT-go-microservice-service \
    --region $AWS_DEFAULT_REGION

echo "✅ Deployment completed successfully!"
