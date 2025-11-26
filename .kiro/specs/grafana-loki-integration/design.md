# Design Document: Grafana Loki Integration

## Overview

This design document outlines the architecture and implementation approach for integrating Grafana Loki log aggregation into the to-do list application. The solution leverages Docker Compose to orchestrate Loki, Grafana, and the application, using the Docker Loki logging driver for seamless log shipping. The design prioritizes simplicity for local development while maintaining flexibility for multi-environment deployments.

## Implementation Approach

We're using a **Docker-native approach** with the Docker Loki logging driver instead of Promtail. This approach is simpler, more reliable, and requires less configuration. The Docker daemon automatically sends container logs to Loki without needing a separate log shipping agent.

**Key Benefits:**
- No additional agent (Promtail) to configure and maintain
- Automatic log shipping from all containers
- Built-in retry and buffering mechanisms
- Simpler Docker Compose configuration
- Less resource overhead

**Architecture Pattern:**
We follow a **sidecar pattern** where Loki and Grafana run alongside the application in the same Docker Compose stack. This keeps everything self-contained and easy to start/stop as a unit.

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     Docker Host (localhost)                  │
│                                                              │
│  ┌────────────────┐         ┌──────────────┐               │
│  │  Task Manager  │────────▶│    Loki      │               │
│  │  Application   │  logs   │  (port 3100) │               │
│  │  (port 8080)   │         │              │               │
│  └────────────────┘         └──────┬───────┘               │
│         │                           │                        │
│         │ HTTP API                  │ queries                │
│         │                           │                        │
│  ┌──────▼──────────────────────────▼───────┐               │
│  │           Grafana (port 3000)            │               │
│  │  - Data source: Loki                     │               │
│  │  - Dashboards                            │               │
│  │  - Query interface                       │               │
│  └──────────────────────────────────────────┘               │
│         │                                                    │
└─────────┼────────────────────────────────────────────────────┘
          │
          ▼
    Developer Browser
    http://localhost:3000
```

### Log Flow

```
Application Code
    │
    ├─ slog.Info("message", fields...)
    │
    ▼
JSON to stdout
    │
    ├─ {"time":"...","level":"INFO","msg":"...","request_id":"..."}
    │
    ▼
Docker Loki Driver
    │
    ├─ Adds labels: service=task-manager, environment=local
    ├─ Buffers logs
    ├─ Retries on failure
    │
    ▼
Loki HTTP API
    │
    ├─ POST /loki/api/v1/push
    ├─ Stores logs with labels
    ├─ Indexes by time and labels
    │
    ▼
Loki Storage
    │
    ├─ Chunks (compressed log data)
    ├─ Index (label lookups)
    │
    ▼
Grafana Queries
    │
    ├─ LogQL: {service="task-manager"} | json | level="ERROR"
    ├─ Time range filtering
    ├─ Field extraction
    │
    ▼
Developer Browser
```

## Components and Interfaces

### 1. Docker Compose Configuration

**File:** `examples/docker-compose.loki.yaml`

**Purpose:** Orchestrates all services (Application, Loki, Grafana) with proper networking and dependencies.

**Key Configuration:**

```yaml
services:
  task-manager:
    build: ..
    ports:
      - "8080:8080"
    environment:
      TASKMANAGER_LOGGING_FORMAT: "json"
      TASKMANAGER_LOGGING_OUTPUT: "stdout"
      TASKMANAGER_LOGGING_LEVEL: "info"
      TASKMANAGER_LOGGING_SERVICE_NAME: "task-manager-api"
      TASKMANAGER_LOGGING_ENVIRONMENT: "local"
    logging:
      driver: loki
      options:
        loki-url: "http://localhost:3100/loki/api/v1/push"
        loki-batch-size: "100"
        loki-retries: "2"
        loki-max-backoff: "1s"
        loki-timeout: "1s"
        labels: "service=task-manager-api,environment=local"
    depends_on:
      - loki

  loki:
    image: grafana/loki:2.9.0
    ports:
      - "3100:3100"
    command: -config.file=/etc/loki/local-config.yaml
    volumes:
      - loki-data:/loki

  grafana:
    image: grafana/grafana:10.2.0
    ports:
      - "3000:3000"
    environment:
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Admin
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana-datasources.yaml:/etc/grafana/provisioning/datasources/datasources.yaml
      - ./grafana-dashboards.yaml:/etc/grafana/provisioning/dashboards/dashboards.yaml
      - ./dashboards:/var/lib/grafana/dashboards
    depends_on:
      - loki

volumes:
  loki-data:
  grafana-data:
```

**Design Decisions:**

1. **Docker Loki Driver:** Simplest approach for local development
2. **Anonymous Grafana Access:** Convenient for local development (disable in production)
3. **Volume Persistence:** Logs and dashboards persist across container restarts
4. **Dependency Chain:** Application → Loki → Grafana ensures proper startup order
5. **Batch Size & Retries:** Optimized for local development (low latency, quick retries)

### 2. Loki Configuration

**File:** `examples/loki-config.yaml` (optional, uses defaults)

**Purpose:** Configure Loki storage, retention, and performance settings.

**Key Settings:**

```yaml
auth_enabled: false  # Disable for local development

server:
  http_listen_port: 3100

ingester:
  lifecycler:
    ring:
      kvstore:
        store: inmemory
      replication_factor: 1
  chunk_idle_period: 5m
  chunk_retain_period: 30s

schema_config:
  configs:
    - from: 2020-10-24
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

storage_config:
  boltdb_shipper:
    active_index_directory: /loki/index
    cache_location: /loki/cache
    shared_store: filesystem
  filesystem:
    directory: /loki/chunks

limits_config:
  retention_period: 168h  # 7 days
  reject_old_samples: true
  reject_old_samples_max_age: 168h

chunk_store_config:
  max_look_back_period: 168h

table_manager:
  retention_deletes_enabled: true
  retention_period: 168h
```

**Design Decisions:**

1. **No Authentication:** Simplified for local development
2. **Single Node:** No replication needed for local setup
3. **7-Day Retention:** Balance between debugging needs and disk space
4. **Filesystem Storage:** Simple, no external dependencies
5. **BoltDB Index:** Lightweight, embedded database

### 3. Grafana Data Source Provisioning

**File:** `examples/grafana-datasources.yaml`

**Purpose:** Automatically configure Loki as a data source when Grafana starts.

```yaml
apiVersion: 1

datasources:
  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    isDefault: true
    editable: true
    jsonData:
      maxLines: 1000
      derivedFields:
        - datasourceUid: loki
          matcherRegex: "request_id=(\\w+)"
          name: RequestID
          url: "$${__value.raw}"
```

**Design Decisions:**

1. **Proxy Access:** Grafana proxies requests to Loki (simpler networking)
2. **Default Data Source:** Automatically selected in query interface
3. **Derived Fields:** Extract request_id for easy correlation
4. **Max Lines:** Limit to 1000 lines to prevent browser overload

### 4. Grafana Dashboard Provisioning

**File:** `examples/grafana-dashboards.yaml`

**Purpose:** Automatically load pre-built dashboards on Grafana startup.

```yaml
apiVersion: 1

providers:
  - name: 'Task Manager Dashboards'
    orgId: 1
    folder: 'Task Manager'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards
```

### 5. Pre-Built Dashboards

**Files:** `examples/dashboards/*.json`

**Purpose:** Provide ready-to-use dashboards for common use cases.

#### Dashboard 1: Application Overview

**File:** `examples/dashboards/application-overview.json`

**Panels:**
- Log volume over time (all levels)
- Error rate (errors per minute)
- Request count by endpoint
- Recent errors (last 20)
- Recent warnings (last 20)

**Query Examples:**
```logql
# Log volume
sum(count_over_time({service="task-manager-api"}[1m]))

# Error rate
sum(rate({service="task-manager-api"} | json | level="ERROR" [1m]))

# Request count by endpoint
sum by (path) (count_over_time({service="task-manager-api"} | json | msg="HTTP request completed" [5m]))
```

#### Dashboard 2: Request Performance

**File:** `examples/dashboards/request-performance.json`

**Panels:**
- Average request duration by endpoint
- 95th percentile request duration
- Slowest requests (top 10)
- Request count by status code
- Request count by method

**Query Examples:**
```logql
# Average duration by endpoint
avg by (path) (
  avg_over_time({service="task-manager-api"} | json | msg="HTTP request completed" | unwrap duration_ms [5m])
)

# 95th percentile
quantile_over_time(0.95, {service="task-manager-api"} | json | unwrap duration_ms [5m])

# Slowest requests
topk(10, 
  max_over_time({service="task-manager-api"} | json | unwrap duration_ms [1h])
)
```

#### Dashboard 3: Authentication & Security

**File:** `examples/dashboards/authentication-security.json`

**Panels:**
- Failed login attempts (last hour)
- Successful registrations (last hour)
- JWT validation failures
- Authorization failures by user
- Authentication attempts by email (masked)

**Query Examples:**
```logql
# Failed logins
count_over_time({service="task-manager-api"} | json | msg=~".*login.*" | level="WARN" [1h])

# Registrations
count_over_time({service="task-manager-api"} | json | msg=~".*registration.*" | level="INFO" [1h])

# JWT failures
{service="task-manager-api"} | json | msg=~".*JWT.*" | level="WARN"
```

#### Dashboard 4: Error Analysis

**File:** `examples/dashboards/error-analysis.json`

**Panels:**
- Error count by operation
- Error messages (grouped and counted)
- Errors by user_id
- Errors by endpoint
- Error timeline (last 24 hours)

**Query Examples:**
```logql
# Errors by operation
sum by (operation) (count_over_time({service="task-manager-api"} | json | level="ERROR" [1h]))

# Error messages grouped
topk(10,
  count by (msg) ({service="task-manager-api"} | json | level="ERROR")
)

# Errors by user
sum by (user_id) (count_over_time({service="task-manager-api"} | json | level="ERROR" | user_id != "" [24h]))
```

## Data Models

### Log Entry Structure

The application already outputs structured JSON logs. Loki will ingest these as-is:

```json
{
  "time": "2025-11-12T10:30:45.123Z",
  "level": "INFO",
  "msg": "HTTP request completed",
  "service": "task-manager-api",
  "environment": "local",
  "request_id": "req_1699785045123_a1b2c3d4",
  "user_id": 42,
  "method": "POST",
  "path": "/tasks",
  "status_code": 201,
  "duration_ms": 45
}
```

### Loki Label Schema

Labels are used for indexing and filtering. Keep labels low-cardinality (few unique values).

**Primary Labels (attached by Docker driver):**
- `service`: "task-manager-api"
- `environment`: "local" | "staging" | "production"
- `container_name`: Docker container name
- `compose_project`: Docker Compose project name

**Extracted Fields (from JSON, not labels):**
- `level`: "DEBUG" | "INFO" | "WARN" | "ERROR"
- `request_id`: Unique per request
- `user_id`: User identifier
- `method`: HTTP method
- `path`: HTTP path
- `status_code`: HTTP status
- `duration_ms`: Request duration
- `operation`: Operation name
- `task_id`: Task identifier
- `error`: Error message

**Why This Separation?**
- Labels are indexed → fast filtering, but high cardinality = high cost
- Fields are stored → can filter after retrieval, unlimited cardinality
- Use labels for broad filtering (service, environment)
- Use field extraction for detailed filtering (request_id, user_id)

## Error Handling

### Docker Loki Driver Failures

**Scenario:** Loki is unavailable or slow

**Handling:**
- Docker driver buffers logs in memory (configurable size)
- Retries with exponential backoff (configured: 2 retries, 1s max backoff)
- If buffer fills, oldest logs are dropped (prevents memory exhaustion)
- Application continues running (logging failures don't crash app)

**Configuration:**
```yaml
logging:
  driver: loki
  options:
    loki-retries: "2"
    loki-max-backoff: "1s"
    loki-timeout: "1s"
```

### Loki Storage Full

**Scenario:** Disk space exhausted

**Handling:**
- Loki stops accepting new logs
- Returns HTTP 500 to Docker driver
- Docker driver buffers and retries
- Logs warning in Loki container logs

**Prevention:**
- Configure retention period (7 days default)
- Monitor disk usage with `docker system df`
- Provide cleanup command in documentation

### Grafana Connection Issues

**Scenario:** Grafana can't reach Loki

**Handling:**
- Grafana shows "Data source error" in UI
- Queries return empty results
- Health check endpoint shows failure

**Resolution:**
- Verify Loki is running: `docker ps`
- Check Loki logs: `docker logs loki`
- Verify network connectivity: `docker exec grafana curl http://loki:3100/ready`

### Query Timeout

**Scenario:** Query takes too long (large time range, complex filter)

**Handling:**
- Loki returns HTTP 504 Gateway Timeout
- Grafana shows timeout error
- Partial results may be displayed

**Prevention:**
- Limit time ranges (use last 1 hour instead of last 7 days)
- Add more specific filters (service, environment, level)
- Use aggregation instead of raw logs for large ranges

## Testing Strategy

### Manual Testing Checklist

1. **Setup Verification:**
   - [ ] Docker Loki plugin installed
   - [ ] Docker Compose starts all services
   - [ ] Grafana accessible at http://localhost:3000
   - [ ] Loki accessible at http://localhost:3100
   - [ ] Application accessible at http://localhost:8080

2. **Log Ingestion:**
   - [ ] Make HTTP request to application
   - [ ] Verify logs appear in Grafana within 5 seconds
   - [ ] Verify JSON fields are parsed correctly
   - [ ] Verify request_id is present and unique

3. **Query Functionality:**
   - [ ] Query all logs: `{service="task-manager-api"}`
   - [ ] Filter by level: `{service="task-manager-api"} | json | level="ERROR"`
   - [ ] Filter by request_id: `{service="task-manager-api"} | json | request_id="req_..."`
   - [ ] Search text: `{service="task-manager-api"} |= "HTTP request"`

4. **Dashboard Functionality:**
   - [ ] Application Overview dashboard loads
   - [ ] Panels show data
   - [ ] Time range selector works
   - [ ] Live tail mode works

5. **Error Scenarios:**
   - [ ] Stop Loki, verify application continues running
   - [ ] Restart Loki, verify logs resume flowing
   - [ ] Query with invalid LogQL, verify error message

### Integration Test Scenarios

**Test 1: End-to-End Log Flow**
```bash
# Start stack
docker-compose -f examples/docker-compose.loki.yaml up -d

# Wait for services
sleep 10

# Make request
curl http://localhost:8080/health

# Query Loki API
curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={service="task-manager-api"}' \
  | jq '.data.result[0].values'

# Expected: Array of log entries with timestamps
```

**Test 2: Request Correlation**
```bash
# Make authenticated request
TOKEN=$(curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}' \
  | jq -r '.token')

curl -X POST http://localhost:8080/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"description":"Test task"}'

# Query logs for this request
# Extract request_id from logs, then query:
curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={service="task-manager-api"} | json | request_id="req_..."'

# Expected: Multiple log entries (request start, auth, DB operation, request complete)
```

**Test 3: Multi-Environment Labels**
```bash
# Start with local environment
docker-compose -f examples/docker-compose.loki.yaml up -d

# Start second instance with staging environment
docker-compose -f examples/docker-compose.loki.yaml \
  -p staging \
  up -d task-manager

# Query by environment
curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={environment="local"}'

curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={environment="staging"}'

# Expected: Separate log streams for each environment
```

### Performance Testing

**Test 1: Log Volume**
```bash
# Generate 1000 requests
for i in {1..1000}; do
  curl -s http://localhost:8080/health > /dev/null
done

# Measure Loki ingestion rate
docker stats loki --no-stream

# Expected: CPU < 50%, Memory < 500MB
```

**Test 2: Query Performance**
```bash
# Query large time range
time curl -G -s "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={service="task-manager-api"}' \
  --data-urlencode 'start=1699785045000000000' \
  --data-urlencode 'end=1699871445000000000'

# Expected: < 5 seconds for 24 hours of logs
```

## Deployment Considerations

### Local Development

**Recommended Setup:**
- Run Grafana Loki only when needed (not always)
- Use Docker Compose profiles to start/stop easily
- Keep 7-day retention (balance debugging vs disk space)
- Use anonymous Grafana access (convenient)

**Commands:**
```bash
# Start everything
docker-compose -f examples/docker-compose.loki.yaml up -d

# Stop everything
docker-compose -f examples/docker-compose.loki.yaml down

# View logs
docker-compose -f examples/docker-compose.loki.yaml logs -f task-manager

# Clear all data
docker-compose -f examples/docker-compose.loki.yaml down -v
```

### Staging Environment

**Recommended Setup:**
- Run Grafana Loki continuously
- Increase retention to 30 days
- Enable Grafana authentication
- Use separate Loki instance or environment label

**Configuration Changes:**
```yaml
# Longer retention
limits_config:
  retention_period: 720h  # 30 days

# Grafana authentication
environment:
  - GF_AUTH_ANONYMOUS_ENABLED=false
  - GF_SECURITY_ADMIN_PASSWORD=secure-password
```

### Production Environment

**Recommended Setup:**
- Use managed Grafana Cloud or dedicated Loki cluster
- Configure authentication and TLS
- Set up alerting rules
- Implement log sampling for high-volume services
- Use remote storage (S3, GCS) instead of local filesystem

**Not Covered in This Spec:**
- Production deployment is out of scope
- Focus is on local development and learning
- Provide guidance document for production considerations

## Security Considerations

### Local Development

**Current Approach:**
- No authentication (Grafana, Loki)
- No TLS/HTTPS
- Exposed only on localhost

**Rationale:**
- Simplified setup for learning
- No external network exposure
- Acceptable risk for local development

### Sensitive Data

**Handled by Application:**
- Email masking (already implemented in logger)
- Token masking (already implemented in logger)
- Password never logged

**Loki Storage:**
- Logs stored in Docker volumes (local filesystem)
- No encryption at rest (not needed for local dev)
- Volumes can be deleted to remove all logs

### Production Considerations

**Required Changes:**
- Enable Grafana authentication
- Enable Loki authentication (multi-tenancy)
- Use TLS for all connections
- Implement log retention policies
- Set up access controls and audit logging

## Documentation Structure

### Quick Start Guide

**File:** `docs/GRAFANA_LOKI_QUICKSTART.md`

**Contents:**
1. Prerequisites (Docker, Docker Compose, Docker Loki plugin)
2. Installation steps (5-minute setup)
3. Verification steps
4. First query examples
5. Troubleshooting common issues

### LogQL Query Reference

**File:** `docs/LOGQL_QUERY_REFERENCE.md`

**Contents:**
1. Basic query syntax
2. Label filtering
3. JSON field extraction
4. Text search
5. Aggregation functions
6. Time range queries
7. Common query patterns

### Dashboard Guide

**File:** `docs/GRAFANA_DASHBOARDS.md`

**Contents:**
1. Overview of pre-built dashboards
2. How to create custom dashboards
3. Panel types and use cases
4. Query examples for each panel type
5. Sharing and exporting dashboards

### Troubleshooting Guide

**File:** `docs/GRAFANA_LOKI_TROUBLESHOOTING.md`

**Contents:**
1. Docker Loki plugin issues
2. No logs appearing in Grafana
3. Loki connection errors
4. Query timeout issues
5. Storage and retention issues
6. Performance problems

## Summary

This design provides a complete, production-ready approach to integrating Grafana Loki for local development. The key design decisions prioritize simplicity and ease of use while maintaining flexibility for future expansion to staging and production environments.

**Core Principles:**
1. **Docker-native:** Use Docker Loki driver, not Promtail
2. **Self-contained:** Everything runs in Docker Compose
3. **Zero-config:** Automatic provisioning of data sources and dashboards
4. **Developer-friendly:** Simple commands, clear documentation
5. **Production-ready foundation:** Easy to extend for production use

The implementation will focus on getting developers up and running quickly while providing powerful log analysis capabilities that integrate seamlessly with the existing structured logging system.
