# GraphQL API Deployment Guide

Comprehensive guide for deploying the Go GraphQL API Boilerplate to production environments.

## Table of Contents

- [Pre-deployment Checklist](#pre-deployment-checklist)
- [Production Configuration](#production-configuration)
- [Docker Deployment](#docker-deployment)
- [AWS Deployment](#aws-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Serverless Deployment](#serverless-deployment)
- [Monitoring & Observability](#monitoring--observability)
- [Security Best Practices](#security-best-practices)

## Pre-deployment Checklist

Before deploying to production:

### Security
- [ ] Set `APP_ENV=production`
- [ ] Set `APP_DEBUG=false`
- [ ] Disable GraphQL Playground (`GRAPHQL_PLAYGROUND_ENABLED=false`)
- [ ] Disable introspection (`GRAPHQL_INTROSPECTION_ENABLED=false`)
- [ ] Generate strong JWT secret (32+ characters)
- [ ] Configure SSL/TLS for database connections
- [ ] Set up CORS properly
- [ ] Implement rate limiting

### Database
- [ ] Set up automated backups
- [ ] Configure connection pooling
- [ ] Enable SSL for PostgreSQL
- [ ] Set up read replicas (if needed)
- [ ] Configure database monitoring

### Monitoring
- [ ] Set up logging aggregation
- [ ] Configure error tracking (Sentry, Rollbar)
- [ ] Set up APM (Application Performance Monitoring)
- [ ] Configure health checks
- [ ] Set up alerting

### Performance
- [ ] Enable dataloader batching
- [ ] Configure proper cache headers
- [ ] Set resource limits (CPU, memory)
- [ ] Load test with expected traffic
- [ ] Optimize database queries

## Production Configuration

### Environment Variables

```env
# Application
APP_HOST=0.0.0.0
APP_PORT=8080
APP_ENV=production
APP_DEBUG=false

# Database
DB_HOST=prod-postgres.example.com
DB_PORT=5432
DB_USER=api_user
DB_PASSWORD=<use-secrets-manager>
DB_NAME=graphql_prod
DB_SSLMODE=require
DB_MAX_CONNECTIONS=50
DB_MIGRATIONS_PATH=file://migrations

# Authentication
AUTH_JWT_SECRET=<use-secrets-manager>
AUTH_JWT_EXPIRATION=24

# GraphQL
GRAPHQL_PLAYGROUND_ENABLED=false
GRAPHQL_INTROSPECTION_ENABLED=false
GRAPHQL_COMPLEXITY_LIMIT=1000
GRAPHQL_QUERY_CACHE_TTL=300

# CORS (adjust for your domain)
CORS_ALLOWED_ORIGINS=https://app.example.com,https://www.example.com
CORS_ALLOWED_METHODS=POST,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW_SECONDS=60

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Monitoring (optional)
SENTRY_DSN=https://xxx@sentry.io/xxx
NEW_RELIC_LICENSE_KEY=xxx
```

### Generating Secrets

```bash
# JWT Secret (32 bytes = 64 hex chars)
openssl rand -hex 32

# Or using Go
go run main.go generate-secret
```

## Docker Deployment

### Production Dockerfile

The existing `Dockerfile` uses multi-stage builds for optimization.

### Docker Compose for Production

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
      target: production
    image: go-graphql-api:latest
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - APP_DEBUG=false
      - GRAPHQL_PLAYGROUND_ENABLED=false
      - GRAPHQL_INTROSPECTION_ENABLED=false
      - DB_HOST=postgres
      - DB_NAME=graphql_prod
    env_file:
      - .env.production
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    networks:
      - app-network
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=graphql_prod
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD_FILE=/run/secrets/db_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    secrets:
      - db_password
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - app-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
    depends_on:
      - api
    restart: unless-stopped
    networks:
      - app-network

volumes:
  postgres_data:
    driver: local

secrets:
  db_password:
    file: ./secrets/db_password.txt

networks:
  app-network:
    driver: bridge
```

### Nginx Configuration

Create `nginx/nginx.conf`:

```nginx
events {
    worker_connections 1024;
}

http {
    upstream graphql_backend {
        least_conn;
        server api:8080 max_fails=3 fail_timeout=30s;
    }

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=graphql_limit:10m rate=10r/s;
    limit_req_status 429;

    server {
        listen 80;
        server_name api.example.com;
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name api.example.com;

        ssl_certificate /etc/nginx/ssl/fullchain.pem;
        ssl_certificate_key /etc/nginx/ssl/privkey.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers HIGH:!aNULL:!MD5;
        ssl_prefer_server_ciphers on;

        # Security headers
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header X-Frame-Options "DENY" always;
        add_header X-XSS-Protection "1; mode=block" always;

        # GraphQL endpoint
        location /query {
            limit_req zone=graphql_limit burst=20 nodelay;

            proxy_pass http://graphql_backend;
            proxy_http_version 1.1;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            # Timeouts
            proxy_connect_timeout 60s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;

            # CORS headers (adjust as needed)
            add_header 'Access-Control-Allow-Origin' 'https://app.example.com' always;
            add_header 'Access-Control-Allow-Methods' 'POST, OPTIONS' always;
            add_header 'Access-Control-Allow-Headers' 'Content-Type, Authorization' always;

            if ($request_method = 'OPTIONS') {
                return 204;
            }
        }

        # Health check
        location /health {
            proxy_pass http://graphql_backend/health;
            access_log off;
        }

        # Block playground in production
        location / {
            return 403;
        }
    }
}
```

### Deploy with Docker Compose

```bash
# Create secrets directory
mkdir -p secrets
echo "your_secure_db_password" > secrets/db_password.txt
chmod 600 secrets/db_password.txt

# Build and start
docker-compose -f docker-compose.prod.yml up -d

# Check logs
docker-compose -f docker-compose.prod.yml logs -f api

# Scale the API
docker-compose -f docker-compose.prod.yml up -d --scale api=3
```

## AWS Deployment

### Option 1: AWS App Runner (Easiest)

AWS App Runner automatically builds and deploys from GitHub.

```bash
# Install AWS CLI
aws configure

# Create App Runner service
aws apprunner create-service \
  --service-name go-graphql-api \
  --source-configuration '{
    "CodeRepository": {
      "RepositoryUrl": "https://github.com/ElioNeto/go-graphql-api-boilerplate",
      "SourceCodeVersion": {
        "Type": "BRANCH",
        "Value": "main"
      },
      "CodeConfiguration": {
        "ConfigurationSource": "API",
        "CodeConfigurationValues": {
          "Runtime": "GO_1",
          "BuildCommand": "go build -o app cmd/api/main.go",
          "StartCommand": "./app",
          "Port": "8080",
          "RuntimeEnvironmentVariables": {
            "APP_ENV": "production",
            "APP_PORT": "8080"
          }
        }
      }
    },
    "AutoDeploymentsEnabled": true
  }' \
  --instance-configuration '{
    "Cpu": "1 vCPU",
    "Memory": "2 GB"
  }'
```

### Option 2: AWS ECS with Fargate

#### Step 1: Push Image to ECR

```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

# Create repository
aws ecr create-repository --repository-name go-graphql-api --region us-east-1

# Build and push
docker build -t go-graphql-api:latest .
docker tag go-graphql-api:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/go-graphql-api:latest
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/go-graphql-api:latest
```

#### Step 2: Create Task Definition

Create `ecs-task-definition.json`:

```json
{
  "family": "go-graphql-api",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "containerDefinitions": [
    {
      "name": "api",
      "image": "<account-id>.dkr.ecr.us-east-1.amazonaws.com/go-graphql-api:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        { "name": "APP_ENV", "value": "production" },
        { "name": "APP_DEBUG", "value": "false" },
        { "name": "GRAPHQL_PLAYGROUND_ENABLED", "value": "false" },
        { "name": "GRAPHQL_INTROSPECTION_ENABLED", "value": "false" }
      ],
      "secrets": [
        {
          "name": "DB_PASSWORD",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:<account-id>:secret:db-password-xxx"
        },
        {
          "name": "AUTH_JWT_SECRET",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:<account-id>:secret:jwt-secret-xxx"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/go-graphql-api",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "api"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
```

#### Step 3: Deploy ECS Service

```bash
# Register task definition
aws ecs register-task-definition --cli-input-json file://ecs-task-definition.json

# Create service
aws ecs create-service \
  --cluster production-cluster \
  --service-name go-graphql-api \
  --task-definition go-graphql-api \
  --desired-count 2 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx],securityGroups=[sg-xxx],assignPublicIp=ENABLED}" \
  --load-balancers "targetGroupArn=arn:aws:elasticloadbalancing:...,containerName=api,containerPort=8080"
```

### Option 3: AWS Lambda with API Gateway

For serverless deployment, see [Serverless Deployment](#serverless-deployment) section.

## Kubernetes Deployment

### Deployment Manifests

Create `k8s/deployment.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: graphql-api
---
apiVersion: v1
kind: Secret
metadata:
  name: api-secrets
  namespace: graphql-api
type: Opaque
stringData:
  DB_PASSWORD: "<base64-encoded>"
  AUTH_JWT_SECRET: "<base64-encoded>"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: graphql-api
  namespace: graphql-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: graphql-api
  template:
    metadata:
      labels:
        app: graphql-api
    spec:
      containers:
      - name: api
        image: <registry>/go-graphql-api:latest
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: APP_ENV
          value: "production"
        - name: APP_DEBUG
          value: "false"
        - name: GRAPHQL_PLAYGROUND_ENABLED
          value: "false"
        - name: DB_HOST
          value: "postgres-service"
        envFrom:
        - secretRef:
            name: api-secrets
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: graphql-api-service
  namespace: graphql-api
spec:
  selector:
    app: graphql-api
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: graphql-api-hpa
  namespace: graphql-api
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: graphql-api
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Deploy to Kubernetes

```bash
# Apply manifests
kubectl apply -f k8s/

# Check status
kubectl get all -n graphql-api

# View logs
kubectl logs -f deployment/graphql-api -n graphql-api

# Port forward for testing
kubectl port-forward svc/graphql-api-service 8080:80 -n graphql-api
```

## Serverless Deployment

### AWS Lambda with API Gateway

Create `lambda/main.go`:

```go
package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/ElioNeto/go-graphql-api-boilerplate/graph"
)

var graphqlHandler *handler.Server

func init() {
	graphqlHandler = handler.NewDefaultServer(
		graph.NewExecutableSchema(graph.Config{
			Resolvers: &graph.Resolver{},
		}),
	)
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse GraphQL request
	var gqlRequest struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}

	if err := json.Unmarshal([]byte(request.Body), &gqlRequest); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"errors":[{"message":"Invalid request"}]}`,
		}, nil
	}

	// Execute GraphQL query
	// ... implementation

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: responseJSON,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
```

## Monitoring & Observability

### Health Checks

Implement a health endpoint that checks:

```go
func (r *queryResolver) Health(ctx context.Context) (string, error) {
	// Check database connection
	if err := r.DB.Ping(ctx); err != nil {
		return "", fmt.Errorf("database unhealthy: %w", err)
	}
	return "ok", nil
}
```

### Structured Logging

```go
import "log/slog"

slog.Info("graphql query executed",
	"operation", operationName,
	"duration_ms", duration.Milliseconds(),
	"complexity", complexity,
)
```

### Error Tracking with Sentry

```go
import "github.com/getsentry/sentry-go"

func init() {
	sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
		Environment: os.Getenv("APP_ENV"),
	})
}

// In error handler
sentry.CaptureException(err)
```

### Prometheus Metrics

```go
import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	graphqlDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "graphql_query_duration_seconds",
			Help: "GraphQL query duration",
		},
		[]string{"operation"},
	)
)

func init() {
	prometheus.MustRegister(graphqlDuration)
}

// Expose metrics
http.Handle("/metrics", promhttp.Handler())
```

## Security Best Practices

### 1. Disable Introspection in Production

```env
GRAPHQL_INTROSPECTION_ENABLED=false
```

### 2. Query Complexity Limits

```go
import "github.com/99designs/gqlgen/graphql/handler/extension"

server.Use(extension.FixedComplexityLimit(1000))
```

### 3. Query Depth Limits

```go
server.SetQueryCache(lru.New(1000))
```

### 4. Rate Limiting

Implement per-user rate limiting:

```go
import "golang.org/x/time/rate"

var limiters = make(map[string]*rate.Limiter)

func getRateLimiter(userID string) *rate.Limiter {
	if limiter, ok := limiters[userID]; ok {
		return limiter
	}
	limiter := rate.NewLimiter(rate.Limit(10), 100)
	limiters[userID] = limiter
	return limiter
}
```

### 5. CORS Configuration

```go
import "github.com/rs/cors"

corsHandler := cors.New(cors.Options{
	AllowedOrigins: []string{"https://app.example.com"},
	AllowedMethods: []string{"POST", "OPTIONS"},
	AllowedHeaders: []string{"Content-Type", "Authorization"},
	AllowCredentials: true,
})
```

## Database Backups

### Automated Backup Script

```bash
#!/bin/bash
# backup-graphql-db.sh

BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/graphql_db_$DATE.sql.gz"

# Dump database
pg_dump -h $DB_HOST -U $DB_USER graphql_prod | gzip > $BACKUP_FILE

# Upload to S3
aws s3 cp $BACKUP_FILE s3://my-backups/graphql-db/

# Keep only last 30 days
find $BACKUP_DIR -name "graphql_db_*.sql.gz" -mtime +30 -delete

echo "Backup completed: $BACKUP_FILE"
```

Schedule with cron:

```cron
0 2 * * * /path/to/backup-graphql-db.sh
```

## Troubleshooting Production Issues

### High Memory Usage

```bash
# Check memory profile
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Top memory consumers
(pprof) top
```

### Slow Queries

Enable query logging:

```env
APP_DEBUG=true
```

Analyze slow queries in logs.

### Connection Pool Exhaustion

Increase pool size:

```env
DB_MAX_CONNECTIONS=100
```

Monitor active connections:

```sql
SELECT count(*) FROM pg_stat_activity;
```

## Rollback Strategy

### Docker Rollback

```bash
# List previous images
docker images

# Rollback to previous version
docker-compose -f docker-compose.prod.yml down
docker tag go-graphql-api:previous go-graphql-api:latest
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes Rollback

```bash
# View rollout history
kubectl rollout history deployment/graphql-api -n graphql-api

# Rollback to previous revision
kubectl rollout undo deployment/graphql-api -n graphql-api

# Rollback to specific revision
kubectl rollout undo deployment/graphql-api --to-revision=2 -n graphql-api
```

## Next Steps

- Set up CI/CD pipeline (GitHub Actions, GitLab CI)
- Implement blue-green deployments
- Configure CDN for static assets
- Set up database read replicas
- Implement distributed tracing (Jaeger, Zipkin)
