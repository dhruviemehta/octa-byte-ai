

# Deployment Guide

This guide provides comprehensive instructions for deploying the Go microservice platform across different environments.

## Prerequisites

### Required Tools
- **AWS CLI** (v2.0+): `aws --version`
- **Terraform** (v1.5.0+): `terraform --version`
- **Docker** (v20.0+): `docker --version`
- **Go** (v1.21+): `go version`
- **Git**: `git --version`

### AWS Setup
```bash
# Configure AWS CLI
aws configure set aws_access_key_id YOUR_ACCESS_KEY
aws configure set aws_secret_access_key YOUR_SECRET_KEY
aws configure set default.region ap-south-1
aws configure set default.output json

# Verify access
aws sts get-caller-identity
```

### Required Permissions
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:*",
        "ecs:*",
        "rds:*",
        "elasticloadbalancing:*",
        "iam:*",
        "secretsmanager:*",
        "ecr:*",
        "logs:*",
        "cloudwatch:*"
      ],
      "Resource": "*"
    }
  ]
}
```

## Initial Setup

### 1. Clone Repository
```bash
git clone https://github.com/yourusername/octa-byte-devops-assignment.git
cd octa-byte-devops-assignment
```

### 2. Create S3 Backend
```bash
# Create S3 bucket for Terraform state
aws s3 mb s3://octa-byte-terraform-state-staging --region ap-south-1

# Create DynamoDB table for state locking
aws dynamodb create-table \
    --table-name terraform-state-lock \
    --attribute-definitions AttributeName=LockID,AttributeType=S \
    --key-schema AttributeName=LockID,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

### 3. Create ECR Repository
```bash
# Create ECR repository
aws ecr create-repository --repository-name go-microservice --region ap-south-1

# Get repository URI
aws ecr describe-repositories --repository-names go-microservice --query 'repositories[0].repositoryUri' --output text
```

## Environment Deployment

### Staging Environment

#### Step 1: Deploy Infrastructure
```bash
cd terraform/environments/staging

# Initialize Terraform
terraform init

# Review planned changes
terraform plan -var-file="terraform.tfvars"

# Apply infrastructure
terraform apply -var-file="terraform.tfvars"

# Save important outputs
terraform output > ../../../staging-outputs.txt
```

#### Step 2: Build and Push Application
```bash
cd ../../../application

# Get AWS account ID and ECR repository
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
ECR_REPO="$ACCOUNT_ID.dkr.ecr.ap-south-1.amazonaws.com/go-microservice"

# Build Docker image
docker build -t go-microservice:staging .

# Tag for ECR
docker tag go-microservice:staging $ECR_REPO:staging
docker tag go-microservice:staging $ECR_REPO:latest

# Login to ECR
aws ecr get-login-password --region ap-south-1 | docker login --username AWS --password-stdin $ECR_REPO

# Push to ECR
docker push $ECR_REPO:staging
docker push $ECR_REPO:latest
```

#### Step 3: Deploy Application
```bash
# Update ECS service
aws ecs update-service \
    --cluster staging-cluster \
    --service staging-go-microservice-service \
    --force-new-deployment \
    --region ap-south-1

# Wait for deployment to complete
aws ecs wait services-stable \
    --cluster staging-cluster \
    --services staging-go-microservice-service \
    --region ap-south-1
```

#### Step 4: Verify Deployment
```bash
# Get load balancer URL
ALB_DNS=$(aws elbv2 describe-load-balancers \
    --names staging-alb \
    --query 'LoadBalancers[0].DNSName' \
    --output text)

echo "Application URL: http://$ALB_DNS"

# Test health endpoint
curl http://$ALB_DNS/health

# Test API endpoint
curl http://$ALB_DNS/api/users
```

### Production Environment

#### Step 1: Create Production Branch
```bash
git checkout -b production
git push origin production
```

#### Step 2: Manual Approval Process
Production deployments require manual approval through GitHub Actions:

1. Go to GitHub Actions
2. Select "Deploy to Production" workflow
3. Click "Run workflow"
4. Select production environment
5. Specify version/tag to deploy
6. Submit for approval

#### Step 3: Automated Deployment
Once approved, the pipeline will:

1. **Pre-deployment Checks**:
   - Verify staging environment
   - Check for critical alerts
   - Validate database migrations

2. **Database Backup**:
   - Create RDS snapshot
   - Verify backup completion

3. **Blue/Green Deployment**:
   - Deploy to green environment
   - Run health checks
   - Switch traffic gradually

4. **Post-deployment Tests**:
   - Smoke tests
   - Performance validation
   - Error rate monitoring

## Database Management

### Running Migrations

#### Local Development
```bash
cd application

# Install migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations -database "postgres://postgres:password@localhost:5432/appdb?sslmode=disable" up

# Check migration status
migrate -path migrations -database "postgres://postgres:password@localhost:5432/appdb?sslmode=disable" version
```

#### Production Migrations
```bash
# Connect to production database
DB_HOST=$(aws rds describe-db-instances \
    --db-instance-identifier production-postgres \
    --query 'DBInstances[0].Endpoint.Address' \
    --output text)

# Get credentials from Secrets Manager
SECRET=$(aws secretsmanager get-secret-value \
    --secret-id production-db-password \
    --query SecretString --output text)

DB_USER=$(echo $SECRET | jq -r .username)
DB_PASS=$(echo $SECRET | jq -r .password)

# Run migration
migrate -path migrations \
    -database "postgres://$DB_USER:$DB_PASS@$DB_HOST:5432/appdb?sslmode=require" \
    up
```

### Database Backup and Restore

#### Create Manual Backup
```bash
# Create snapshot
aws rds create-db-snapshot \
    --db-instance-identifier production-postgres \
    --db-snapshot-identifier manual-backup-$(date +%Y%m%d%H%M%S)
```

#### Restore from Backup
```bash
# List available snapshots
aws rds describe-db-snapshots \
    --db-instance-identifier production-postgres

# Restore from snapshot
aws rds restore-db-instance-from-db-snapshot \
    --db-instance-identifier production-postgres-restored \
    --db-snapshot-identifier manual-backup-20231201120000
```

## Monitoring Setup

### CloudWatch Alarms
```bash
# Create CPU alarm
aws cloudwatch put-metric-alarm \
    --alarm-name "staging-high-cpu" \
    --alarm-description "High CPU utilization" \
    --metric-name CPUUtilization \
    --namespace AWS/ECS \
    --statistic Average \
    --period 300 \
    --threshold 80 \
    --comparison-operator GreaterThanThreshold \
    --evaluation-periods 2 \
    --dimensions Name=ServiceName,Value=staging-go-microservice-service

# Create memory alarm
aws cloudwatch put-metric-alarm \
    --alarm-name "staging-high-memory" \
    --alarm-description "High memory utilization" \
    --metric-name MemoryUtilization \
    --namespace AWS/ECS \
    --statistic Average \
    --period 300 \
    --threshold 85 \
    --comparison-operator GreaterThanThreshold \
    --evaluation-periods 2
```


## Troubleshooting

### Common Issues

#### 1. Terraform State Lock
```bash
# List locks
aws dynamodb scan --table-name terraform-state-lock

# Force unlock (use carefully)
terraform force-unlock LOCK_ID
```

#### 2. ECS Service Issues
```bash
# Check service status
aws ecs describe-services \
    --cluster staging-cluster \
    --services staging-go-microservice-service

# Check task definition
aws ecs describe-task-definition \
    --task-definition staging-go-microservice

# View logs
aws logs tail /ecs/staging-go-microservice --follow
```

#### 3. Database Connection Issues
```bash
# Test database connectivity
aws rds describe-db-instances \
    --db-instance-identifier staging-postgres \
    --query 'DBInstances[0].DBInstanceStatus'

# Check security groups
aws ec2 describe-security-groups \
    --group-names staging-db-sg
```

#### 4. Load Balancer Issues
```bash
# Check target group health
aws elbv2 describe-target-health \
    --target-group-arn $(aws elbv2 describe-target-groups \
        --names staging-app-tg \
        --query 'TargetGroups[0].TargetGroupArn' \
        --output text)

# Check load balancer status
aws elbv2 describe-load-balancers \
    --names staging-alb \
    --query 'LoadBalancers[0].State'
```

### Log Analysis

#### Application Logs
```bash
# Stream application logs
aws logs tail /ecs/staging-go-microservice --follow

# Search for errors
aws logs filter-log-events \
    --log-group-name /ecs/staging-go-microservice \
    --filter-pattern "ERROR"

# Get logs for specific time range
aws logs filter-log-events \
    --log-group-name /ecs/staging-go-microservice \
    --start-time $(date -d '1 hour ago' +%s)000
```

#### Infrastructure Logs
```bash
# ALB access logs (if enabled)
aws s3 ls s3://my-alb-logs/AWSLogs/$(aws sts get-caller-identity --query Account --output text)/elasticloadbalancing/ap-south-1/

# VPC Flow Logs (if enabled)
aws logs describe-log-groups --log-group-name-prefix /aws/vpc/flowlogs
```

## Rollback Procedures

### Application Rollback
```bash
# Quick rollback using script
./scripts/rollback.sh staging

# Manual rollback to specific version
./scripts/rollback.sh staging v1.2.3
```

### Infrastructure Rollback
```bash
# Rollback to previous Terraform state
cd terraform/environments/staging

# Check state history
terraform state list

# Apply previous configuration
git checkout HEAD~1 -- .
terraform plan
terraform apply
```

### Database Rollback
```bash
# Restore from automated backup
aws rds restore-db-instance-to-point-in-time \
    --source-db-instance-identifier staging-postgres \
    --target-db-instance-identifier staging-postgres-restored \
    --restore-time 2023-12-01T10:00:00.000Z
```

## Performance Optimization

### Auto Scaling Configuration
```bash
# Create auto scaling target
aws application-autoscaling register-scalable-target \
    --service-namespace ecs \
    --resource-id service/staging-cluster/staging-go-microservice-service \
    --scalable-dimension ecs:service:DesiredCount \
    --min-capacity 2 \
    --max-capacity 10

# Create scaling policy
aws application-autoscaling put-scaling-policy \
    --policy-name staging-scale-up \
    --service-namespace ecs \
    --resource-id service/staging-cluster/staging-go-microservice-service \
    --scalable-dimension ecs:service:DesiredCount \
    --policy-type TargetTrackingScaling \
    --target-tracking-scaling-policy-configuration file://scaling-policy.json
```

### Database Performance Tuning
```bash
# Enable Performance Insights
aws rds modify-db-instance \
    --db-instance-identifier staging-postgres \
    --enable-performance-insights \
    --performance-insights-retention-period 7

# Create read replica for read scaling
aws rds create-db-instance-read-replica \
    --db-instance-identifier staging-postgres-replica \
    --source-db-instance-identifier staging-postgres
```

## Maintenance

### Regular Maintenance Tasks

#### Weekly
- Review CloudWatch metrics and logs
- Check security patches for base images
- Validate backup integrity
- Review and rotate secrets

#### Monthly
- Update dependencies in go.mod
- Review and optimize costs
- Security vulnerability assessment
- Performance baseline review

#### Quarterly
- Infrastructure security audit
- Disaster recovery testing
- Capacity planning review
- Documentation updates

### Scheduled Maintenance Window
```bash

# Schedule maintenance window
aws rds modify-db-instance \
    --db-instance-identifier staging-postgres \
    --preferred-maintenance-window "sun:03:00-sun:04:00"

# Update ECS service during maintenance
aws ecs update-service \
    --cluster staging-cluster \
    --service staging-go-microservice-service \
    --desired-count 0

# Wait for tasks to stop
aws ecs wait services-stable \
    --cluster staging-cluster \
    --services staging-go-microservice-service

# Perform maintenance tasks
# ...

# Restore service
aws ecs update-service \
    --cluster staging-cluster \
    --service staging-go-microservice-service \
    --desired-count 2
```