# ARCHITECTURE.md

# System Architecture Documentation

## Overview

This document describes the architecture of the Go microservice platform, including infrastructure design, application architecture, and deployment strategies.

## Infrastructure Architecture

### AWS Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                          Internet                               │
└─────────────────────┬───────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────┐
│                Internet Gateway                                 │
└─────────────────────┬───────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────┐
│                Application Load Balancer                       │
│             (ap-south-1a, ap-south-1b)                          │
└─────────────────────┬───────────────────────────────────────────┘
                      │
        ┌─────────────┴─────────────┐
        │                           │
┌───────▼────────┐         ┌────────▼───────┐
│ Public Subnet  │         │ Public Subnet  │
│ 10.0.1.0/24    │         │ 10.0.2.0/24    │
│ ap-south-1a     │         │ ap-south-1b     │
│ NAT Gateway    │         │ NAT Gateway    │
└───────┬────────┘         └────────┬───────┘
        │                           │
┌───────▼────────┐         ┌────────▼───────┐
│ Private Subnet │         │ Private Subnet │
│ 10.0.10.0/24   │         │ 10.0.11.0/24   │
│ ap-south-1a     │         │ ap-south-1b     │
│ ECS Tasks      │         │ ECS Tasks      │
└───────┬────────┘         └────────┬───────┘
        │                           │
        └─────────────┬─────────────┘
                      │
              ┌───────▼────────┐
              │ Private Subnet │
              │ 10.0.20.0/24   │
              │ RDS PostgreSQL │
              │ Multi-AZ       │
              └────────────────┘
```

## Application Architecture

### Microservice Design

The application follows clean architecture principles with clear separation of concerns:

```
application/
├── cmd/server/           # Application entry point
├── internal/
│   ├── handlers/         # HTTP request handlers
│   ├── models/          # Data models
│   ├── database/        # Database connection and operations
│   └── config/          # Configuration management
└── pkg/
    └── logger/          # Shared logging utilities
```

### Design Patterns Used

1. **Dependency Injection**: All dependencies are injected through constructors
2. **Repository Pattern**: Database operations are abstracted through interfaces
3. **Middleware Pattern**: Cross-cutting concerns (logging, metrics, rate limiting)
4. **12-Factor App**: Environment-based configuration, stateless processes

### API Design

RESTful API following OpenAPI 3.0 specifications:

- **GET /health**: Health check endpoint
- **GET /ready**: Readiness probe
- **GET /metrics**: Prometheus metrics
- **GET /api/users**: List users
- **POST /api/users**: Create user
- **GET /api/users/{id}**: Get user by ID

## Data Architecture

### Database Design

PostgreSQL database with the following schema:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);
```

### Connection Pooling

- **Max Open Connections**: 25
- **Max Idle Connections**: 5
- **Connection Lifetime**: 5 minutes

## Security Architecture

### Network Security

1. **VPC Isolation**: All resources in private VPC
2. **Security Groups**: Minimal required access rules
3. **Private Subnets**: Database and application in private subnets
4. **NAT Gateways**: Controlled internet access for updates

### Application Security

1. **Secrets Management**: AWS Secrets Manager for database credentials
2. **Rate Limiting**: 100 requests/second with burst of 10
3. **Request Logging**: All requests logged with correlation IDs
4. **Input Validation**: JSON schema validation for API requests

### Infrastructure Security

1. **IAM Roles**: Least privilege access for ECS tasks
2. **Encryption**: Data encrypted at rest and in transit
3. **Container Scanning**: Trivy security scans in CI/CD
4. **Dependency Scanning**: Go module vulnerability checks

## Scalability Design

### Horizontal Scaling

- **ECS Fargate**: Auto-scaling based on CPU/memory metrics
- **Load Balancer**: Distributes traffic across multiple instances
- **Database**: RDS with read replicas (can be added)

### Performance Optimization

- **Connection Pooling**: Efficient database connection management
- **Graceful Shutdown**: 30-second timeout for ongoing requests
- **Health Checks**: Fast response endpoints for load balancer

## Monitoring Architecture

### Observability Stack

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │───▶│   CloudWatch    │───▶│    Grafana      │
│   (Metrics)     │    │    (Storage)    │    │  (Visualization)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Logs        │    │     Alarms      │    │   Dashboards    │
│  (Structured)   │    │   (Proactive)   │    │   (Real-time)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Key Metrics

1. **Golden Signals**:
   - **Latency**: Request response times (p50, p95, p99)
   - **Traffic**: Requests per second
   - **Errors**: Error rate percentage
   - **Saturation**: CPU, memory, connection utilization

2. **Business Metrics**:
   - User creation rate
   - API endpoint usage
   - Database query performance

## Deployment Architecture

### CI/CD Pipeline

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Git Push  │───▶│   CI Tests  │───▶│   Build     │───▶│   Deploy    │
│             │    │   Security  │    │   Image     │    │   Staging   │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                            │                                    │
                            ▼                                    ▼
                   ┌─────────────┐                      ┌─────────────┐
                   │   Parallel  │                      │   Manual    │
                   │   Testing   │                      │   Approval  │
                   └─────────────┘                      └─────────────┘
                                                                │
                                                                ▼
                                                       ┌─────────────┐
                                                       │   Deploy    │
                                                       │ Production  │
                                                       └─────────────┘
```

### Blue/Green Deployment

Production deployments use blue/green strategy:

1. **Blue Environment**: Current production
2. **Green Environment**: New version deployment
3. **Health Checks**: Validate green environment
4. **Traffic Switch**: Route traffic to green
5. **Rollback**: Switch back to blue if issues

## Disaster Recovery

### Backup Strategy

1. **Database Backups**:
   - Automated daily snapshots
   - 7-day retention period
   - Cross-region backup copying
   - Point-in-time recovery

2. **Infrastructure**:
   - Terraform state in S3 with versioning
   - Multi-AZ deployment
   - Infrastructure as Code for quick rebuild

### Recovery Procedures

1. **RTO (Recovery Time Objective)**: 15 minutes
2. **RPO (Recovery Point Objective)**: 1 hour
3. **Automated Failover**: RDS Multi-AZ automatic failover
4. **Manual Procedures**: Documented runbooks for disaster scenarios

## Cost Optimization

### Infrastructure Costs

1. **ECS Fargate**: Pay-per-use, automatic scaling
2. **RDS**: Right-sized instances, automated backups
3. **Load Balancer**: Shared across environments
4. **CloudWatch**: Optimized log retention policies

### Estimated Monthly Costs

- **Staging Environment**: ~$75/month
- **Production Environment**: ~$150/month
- **Monitoring Stack**: ~$25/month

---

# SECURITY.md

# Security Documentation

## Security Framework

This document outlines the comprehensive security measures implemented across the infrastructure, application, and deployment pipeline.

## Infrastructure Security

### Network Security

#### VPC Configuration
- **Private Subnets**: All compute resources in private subnets
- **Public Subnets**: Only load balancer and NAT gateways
- **Network ACLs**: Additional layer of network security
- **Security Groups**: Stateful firewall rules

#### Security Group Rules

**ALB Security Group**:
```hcl
Inbound:
  - Port 80/443 from 0.0.0.0/0 (HTTP/HTTPS traffic)
Outbound:
  - All traffic to application security group
```

**Application Security Group**:
```hcl
Inbound:
  - Port 8080 from ALB security group only
Outbound:
  - Port 5432 to database security group
  - Port 443 to 0.0.0.0/0 (HTTPS for API calls)
```

**Database Security Group**:
```hcl
Inbound:
  - Port 5432 from application security group only
Outbound:
  - None (default deny)
```

### Data Encryption

#### Encryption at Rest
- **RDS**: AES-256 encryption enabled
- **EBS Volumes**: Encrypted with AWS KMS
- **S3 Buckets**: Server-side encryption enabled
- **Secrets Manager**: Automatic encryption

#### Encryption in Transit
- **ALB**: HTTPS termination with TLS 1.2+
- **Database**: SSL/TLS connections required
- **Internal**: Service-to-service encryption
- **API**: HTTPS only communication

### Access Control

#### IAM Policies

**ECS Task Execution Role**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "secretsmanager:GetSecretValue"
      ],
      "Resource": "*"
    }
  ]
}
```

**Deployment Role** (Minimal Permissions):
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecs:UpdateService",
        "ecs:DescribeServices",
        "ecs:DescribeTasks"
      ],
      "Resource": [
        "arn:aws:ecs:*:*:service/*/go-microservice-service",
        "arn:aws:ecs:*:*:task/*"
      ]
    }
  ]
}
```

## Application Security

### Input Validation

#### Request Validation
```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=100"`
    Email string `json:"email" validate:"required,email"`
}
```

#### Sanitization
- SQL injection prevention through parameterized queries
- XSS prevention through proper encoding
- Input length limits enforced
- Email format validation

### Rate Limiting

```go
// 100 requests per second, burst of 10
limiter := rate.NewLimiter(rate.Limit(100), 10)

func (h *Handlers) RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### Secrets Management

#### Environment Variables (Non-sensitive)
- `PORT`
- `LOG_LEVEL`
- `DB_HOST`
- `DB_PORT`
- `DB_NAME`

#### AWS Secrets Manager (Sensitive)
- `DB_USER`
- `DB_PASSWORD`
- API keys (if any)
- Third-party credentials

### Security Headers

```go
func securityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000")
        next.ServeHTTP(w, r)
    })
}
```

## Container Security

### Base Image Security
- **Alpine Linux**: Minimal attack surface
- **Non-root User**: Application runs as non-root
- **Multi-stage Build**: Reduces final image size
- **Dependency Scanning**: Trivy scans for vulnerabilities

### Dockerfile Security Best Practices

```dockerfile
# Use specific version tags
FROM golang:1.21-alpine AS builder

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Copy only necessary files
COPY go.mod go.sum ./

# Use non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=5s \
  CMD curl -f http://localhost:8080/health || exit 1
```

### Container Runtime Security

#### ECS Task Definition Security
```json
{
  "requiresCompatibilities": ["FARGATE"],
  "networkMode": "awsvpc",
  "containerDefinitions": [{
    "name": "go-microservice",
    "user": "1000:1000",
    "readonlyRootFilesystem": true,
    "linuxParameters": {
      "capabilities": {
        "drop": ["ALL"]
      }
    }
  }]
}
```

## CI/CD Security

### Pipeline Security

#### Secret Scanning
```yaml
- name: Run secret scan
  uses: trufflesecurity/trufflehog@main
  with:
    path: ./
    base: main
    head: HEAD
```

#### Dependency Scanning
```yaml
- name: Run Gosec Security Scanner
  uses: securecodewarrior/github-action-gosec@master
  with:
    args: '-fmt sarif -out gosec.sarif ./...'
```

#### Container Image Scanning
```yaml
- name: Scan Docker image
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: ${{ env.ECR_REPOSITORY }}:latest
    format: 'sarif'
    output: 'trivy-results.sarif'
```

### GitHub Security

#### Branch Protection Rules
- Require pull request reviews
- Require status checks to pass
- Require branches to be up to date
- Include administrators in restrictions

#### Environment Protection
- Required reviewers for production
- Wait timer before deployment
- Environment-specific secrets

### Supply Chain Security

#### Dependency Management
```go
// go.mod with specific versions
require (
    github.com/gorilla/mux v1.8.0
    github.com/lib/pq v1.10.9
    // ... other dependencies with pinned versions
)
```

#### Image Signing (Future Enhancement)
- Cosign for container image signing
- Policy enforcement in deployment
- Attestation generation

## Monitoring Security

### Security Metrics

#### Failed Authentication Attempts
```promql
increase(http_requests_total{status_code="401"}[5m])
```

#### Unusual Traffic Patterns
```promql
rate(http_requests_total[5m]) > 1000
```

#### Error Rate Spikes
```promql
rate(http_requests_total{status_code=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.05
```

### Log Security

#### Structured Logging
```go
logger.Info("User action",
    "user_id", userID,
    "action", "create_user",
    "ip_address", request.RemoteAddr,
    "correlation_id", correlationID,
)
```

#### Log Retention
- Application logs: 30 days
- Access logs: 90 days
- Audit logs: 1 year
- Security logs: 7 years

### Incident Response

#### Security Incident Runbook

1. **Detection**: Automated alerts trigger investigation
2. **Assessment**: Determine scope and impact
3. **Containment**: Isolate affected systems
4. **Eradication**: Remove threat and vulnerabilities
5. **Recovery**: Restore normal operations
6. **Lessons Learned**: Post-incident review

#### Emergency Contacts
- Security Team: security@company.com
- DevOps Team: devops@company.com
- Management: management@company.com

## Compliance

### Security Standards
- **OWASP Top 10**: Application security best practices
- **CIS Benchmarks**: Infrastructure hardening
- **AWS Well-Architected**: Security pillar compliance
- **SOC 2**: Security controls framework

### Audit Requirements
- Quarterly security assessments
- Annual penetration testing
- Continuous vulnerability scanning
- Security training for developers

### Data Protection
- **Data Classification**: Public, internal, confidential
- **Data Retention**: Automated cleanup policies
- **Data Minimization**: Collect only necessary data
- **Right to Deletion**: User data removal procedures

---

# DEPLOYMENT.md

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

### Grafana Dashboard Import
```bash
# Start Grafana locally
docker run -d \
    --name=grafana \
    -p 3000:3000 \
    -e "GF_SECURITY_ADMIN_PASSWORD=admin" \
    grafana/grafana

# Access Grafana at http://localhost:3000
# Username: admin, Password: admin

# Import dashboards from monitoring/grafana/dashboards/
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