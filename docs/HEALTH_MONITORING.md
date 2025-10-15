# Health Check & Monitoring Guide

## Overview

UniFi Threat Sync includes a built-in HTTP health check server that exposes endpoints for monitoring, health checks, and metrics collection.

## Configuration

Enable health checks in `config.yaml`:

```yaml
health:
  enabled: true  # Enable/disable health server
  port: 8080     # Port for health endpoints
```

## Endpoints

### `/health` or `/healthz` - Liveness Probe

Indicates if the service is running and healthy.

**Response Codes:**
- `200 OK` - Service is healthy
- `503 Service Unavailable` - Service is unhealthy

**Example Request:**
```bash
curl http://localhost:8080/health
```

**Example Response:**
```json
{
  "status": "healthy",
  "version": "v1.0.0",
  "uptime": "2h15m30s",
  "lastSync": "5m ago",
  "syncCount": 12,
  "errorCount": 0,
  "timestamp": "2025-10-13T18:45:00Z"
}
```

**Fields:**
- `status` - Current health status (`healthy` or `unhealthy`)
- `version` - Application version
- `uptime` - Time since service started
- `lastSync` - Time since last successful sync
- `syncCount` - Total number of successful syncs
- `errorCount` - Total number of errors encountered
- `timestamp` - Current server time

---

### `/ready` or `/readiness` - Readiness Probe

Indicates if the service is ready to perform syncs.

**Response Codes:**
- `200 OK` - Service is ready
- `503 Service Unavailable` - Service is not ready (waiting for first sync)

**Example Request:**
```bash
curl http://localhost:8080/ready
```

**Example Response (Ready):**
```json
{
  "ready": true,
  "message": "Ready to serve"
}
```

**Example Response (Not Ready):**
```json
{
  "ready": false,
  "message": "Waiting for first successful sync"
}
```

---

### `/metrics` - Prometheus Metrics

Exports metrics in Prometheus format.

**Example Request:**
```bash
curl http://localhost:8080/metrics
```

**Example Response:**
```
# HELP unifi_threat_sync_up Is the service up
# TYPE unifi_threat_sync_up gauge
unifi_threat_sync_up 1

# HELP unifi_threat_sync_ready Is the service ready
# TYPE unifi_threat_sync_ready gauge
unifi_threat_sync_ready 1

# HELP unifi_threat_sync_sync_total Total number of syncs
# TYPE unifi_threat_sync_sync_total counter
unifi_threat_sync_sync_total 12

# HELP unifi_threat_sync_errors_total Total number of errors
# TYPE unifi_threat_sync_errors_total counter
unifi_threat_sync_errors_total 0

# HELP unifi_threat_sync_uptime_seconds Uptime in seconds
# TYPE unifi_threat_sync_uptime_seconds gauge
unifi_threat_sync_uptime_seconds 8130
```

**Available Metrics:**
- `unifi_threat_sync_up` - Service health (1 = up, 0 = down)
- `unifi_threat_sync_ready` - Service readiness (1 = ready, 0 = not ready)
- `unifi_threat_sync_sync_total` - Total successful syncs (counter)
- `unifi_threat_sync_errors_total` - Total errors (counter)
- `unifi_threat_sync_uptime_seconds` - Uptime in seconds (gauge)

---

## Docker Integration

### Docker Run with Health Port

```bash
docker run -d \
  --name unifi-threat-sync \
  -p 8080:8080 \
  -e UNIFI_USER=admin \
  -e UNIFI_PASS=secret \
  -v $(pwd)/config.yaml:/config/config.yaml:ro \
  ghcr.io/0x4272616e646f6e/unifi-threat-sync:latest
```

### Docker Compose with Health Check

```yaml
version: '3.8'
services:
  unifi-threat-sync:
    image: ghcr.io/0x4272616e646f6e/unifi-threat-sync:latest
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

Check health status:
```bash
docker ps  # Shows health status in STATUS column
docker inspect unifi-threat-sync | jq '.[0].State.Health'
```

---

## Kubernetes Integration

### Deployment with Probes

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: unifi-threat-sync
        ports:
        - name: health
          containerPort: 8080
        
        livenessProbe:
          httpGet:
            path: /health
            port: health
          initialDelaySeconds: 10
          periodSeconds: 30
          timeoutSeconds: 5
          failureThreshold: 3
        
        readinessProbe:
          httpGet:
            path: /ready
            port: health
          initialDelaySeconds: 5
          periodSeconds: 10
          timeoutSeconds: 3
          failureThreshold: 3
```

### Service for Health Checks

```yaml
apiVersion: v1
kind: Service
metadata:
  name: unifi-threat-sync
spec:
  ports:
  - name: health
    port: 8080
    targetPort: health
  selector:
    app: unifi-threat-sync
```

Test from within cluster:
```bash
kubectl run -it --rm debug --image=busybox --restart=Never -- \
  wget -O- http://unifi-threat-sync:8080/health
```

---

## Prometheus Integration

### ServiceMonitor (Prometheus Operator)

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: unifi-threat-sync
spec:
  selector:
    matchLabels:
      app: unifi-threat-sync
  endpoints:
  - port: health
    path: /metrics
    interval: 30s
```

### Prometheus Scrape Config (Static)

```yaml
scrape_configs:
  - job_name: 'unifi-threat-sync'
    static_configs:
    - targets: ['unifi-threat-sync:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

---

## Monitoring Best Practices

### 1. Liveness Probe
- **Purpose**: Detect if the service is alive
- **Action**: Restart container if unhealthy
- **Recommendation**: 
  - Initial delay: 10-30s
  - Period: 30s
  - Failure threshold: 3

### 2. Readiness Probe
- **Purpose**: Detect if service can handle traffic
- **Action**: Remove from load balancer if not ready
- **Recommendation**:
  - Initial delay: 5-10s
  - Period: 10s
  - Failure threshold: 3

### 3. Metrics Collection
- **Scrape Interval**: 30-60s
- **Retention**: Based on your monitoring system
- **Alerts**: Set up based on error rate and sync success

---

## Example Monitoring Queries

### Prometheus Queries

```promql
# Check if service is up
unifi_threat_sync_up

# Sync success rate
rate(unifi_threat_sync_sync_total[5m])

# Error rate
rate(unifi_threat_sync_errors_total[5m])

# Service uptime
unifi_threat_sync_uptime_seconds / 3600  # In hours
```

### Alert Rules

```yaml
groups:
- name: unifi_threat_sync
  rules:
  - alert: UnifiThreatSyncDown
    expr: unifi_threat_sync_up == 0
    for: 5m
    annotations:
      summary: "UniFi Threat Sync is down"
  
  - alert: UnifiThreatSyncErrors
    expr: rate(unifi_threat_sync_errors_total[5m]) > 0.1
    for: 10m
    annotations:
      summary: "High error rate in UniFi Threat Sync"
  
  - alert: UnifiThreatSyncNotReady
    expr: unifi_threat_sync_ready == 0
    for: 15m
    annotations:
      summary: "UniFi Threat Sync not ready"
```

---

## Troubleshooting

### Health endpoint not responding

```bash
# Check if port is exposed
docker port unifi-threat-sync

# Check if health server is enabled in config
cat config.yaml | grep -A2 health

# Check logs for errors
docker logs unifi-threat-sync | grep health
```

### Service always not ready

The service becomes "ready" after the first successful sync. If it's stuck in not-ready:

1. Check logs for sync errors:
   ```bash
   docker logs unifi-threat-sync | grep -i error
   ```

2. Verify UniFi controller is accessible:
   ```bash
   curl -k https://your-controller:8443
   ```

3. Check feed URLs are accessible:
   ```bash
   curl https://www.spamhaus.org/drop/drop.txt
   ```

### Metrics not appearing in Prometheus

1. Verify Prometheus can reach the service:
   ```bash
   kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
     curl http://unifi-threat-sync:8080/metrics
   ```

2. Check ServiceMonitor/scrape config is correct

3. Check Prometheus targets page for errors

---

## Security Considerations

1. **Network Policy**: Restrict access to health port in production
   ```yaml
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: unifi-threat-sync
   spec:
     ingress:
     - from:
       - namespaceSelector:
           matchLabels:
             name: monitoring
       ports:
       - port: 8080
   ```

2. **Authentication**: Health endpoints don't require auth (by design for K8s probes)
   - Keep them internal to your cluster
   - Don't expose externally without additional auth

3. **Rate Limiting**: Built-in HTTP server has timeouts but no rate limiting
   - Use a reverse proxy if exposed externally

---

## Additional Resources

- [Kubernetes Liveness/Readiness Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [Prometheus Metrics Best Practices](https://prometheus.io/docs/practices/naming/)
- [Docker Healthcheck Reference](https://docs.docker.com/engine/reference/builder/#healthcheck)
