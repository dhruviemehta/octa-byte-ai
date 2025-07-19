
# System Architecture Documentation

## Overview

This document describes the architecture of the Go microservice platform, including infrastructure design, application architecture, and deployment strategies.

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
- **GET /metrics**:  metrics
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
