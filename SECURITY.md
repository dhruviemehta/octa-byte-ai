


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

