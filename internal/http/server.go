package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// HealthServer provides health check endpoints
type HealthServer struct {
	server      *http.Server
	port        int
	healthy     atomic.Bool
	ready       atomic.Bool
	lastSync    atomic.Value // stores time.Time
	syncCount   atomic.Int64
	errorCount  atomic.Int64
	version     string
	startTime   time.Time
}

// HealthStatus represents the health check response
type HealthStatus struct {
	Status      string    `json:"status"`
	Version     string    `json:"version"`
	Uptime      string    `json:"uptime"`
	LastSync    string    `json:"lastSync,omitempty"`
	SyncCount   int64     `json:"syncCount"`
	ErrorCount  int64     `json:"errorCount"`
	Timestamp   time.Time `json:"timestamp"`
}

// ReadinessStatus represents the readiness check response
type ReadinessStatus struct {
	Ready     bool   `json:"ready"`
	Message   string `json:"message,omitempty"`
}

// NewHealthServer creates a new health check server
func NewHealthServer(port int, version string) *HealthServer {
	hs := &HealthServer{
		port:      port,
		version:   version,
		startTime: time.Now(),
	}
	
	// Initially healthy but not ready (until first sync)
	hs.healthy.Store(true)
	hs.ready.Store(false)
	
	mux := http.NewServeMux()
	mux.HandleFunc("/health", hs.handleHealth)
	mux.HandleFunc("/healthz", hs.handleHealth) // Kubernetes alias
	mux.HandleFunc("/ready", hs.handleReady)
	mux.HandleFunc("/readiness", hs.handleReady) // Kubernetes alias
	mux.HandleFunc("/metrics", hs.handleMetrics)
	
	hs.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	
	return hs
}

// Start starts the health check server
func (hs *HealthServer) Start() error {
	go func() {
		if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Health server error: %v\n", err)
		}
	}()
	fmt.Printf("Health check server listening on :%d\n", hs.port)
	return nil
}

// Stop gracefully stops the health check server
func (hs *HealthServer) Stop(ctx context.Context) error {
	return hs.server.Shutdown(ctx)
}

// SetHealthy sets the healthy status
func (hs *HealthServer) SetHealthy(healthy bool) {
	hs.healthy.Store(healthy)
}

// SetReady sets the ready status
func (hs *HealthServer) SetReady(ready bool) {
	hs.ready.Store(ready)
}

// RecordSync records a successful sync
func (hs *HealthServer) RecordSync() {
	hs.lastSync.Store(time.Now())
	hs.syncCount.Add(1)
	hs.ready.Store(true) // Ready after first successful sync
}

// RecordError records an error
func (hs *HealthServer) RecordError() {
	hs.errorCount.Add(1)
}

// handleHealth handles the /health endpoint
func (hs *HealthServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	status := HealthStatus{
		Status:     "healthy",
		Version:    hs.version,
		Uptime:     time.Since(hs.startTime).Round(time.Second).String(),
		SyncCount:  hs.syncCount.Load(),
		ErrorCount: hs.errorCount.Load(),
		Timestamp:  time.Now(),
	}
	
	if lastSync := hs.lastSync.Load(); lastSync != nil {
		if t, ok := lastSync.(time.Time); ok {
			status.LastSync = time.Since(t).Round(time.Second).String() + " ago"
		}
	}
	
	if !hs.healthy.Load() {
		status.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleReady handles the /ready endpoint
func (hs *HealthServer) handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	ready := hs.ready.Load()
	status := ReadinessStatus{
		Ready: ready,
	}
	
	if !ready {
		status.Message = "Waiting for first successful sync"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		status.Message = "Ready to serve"
		w.WriteHeader(http.StatusOK)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleMetrics handles the /metrics endpoint (basic text format)
func (hs *HealthServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	
	fmt.Fprintf(w, "# HELP unifi_threat_sync_up Is the service up\n")
	fmt.Fprintf(w, "# TYPE unifi_threat_sync_up gauge\n")
	if hs.healthy.Load() {
		fmt.Fprintf(w, "unifi_threat_sync_up 1\n")
	} else {
		fmt.Fprintf(w, "unifi_threat_sync_up 0\n")
	}
	
	fmt.Fprintf(w, "# HELP unifi_threat_sync_ready Is the service ready\n")
	fmt.Fprintf(w, "# TYPE unifi_threat_sync_ready gauge\n")
	if hs.ready.Load() {
		fmt.Fprintf(w, "unifi_threat_sync_ready 1\n")
	} else {
		fmt.Fprintf(w, "unifi_threat_sync_ready 0\n")
	}
	
	fmt.Fprintf(w, "# HELP unifi_threat_sync_sync_total Total number of syncs\n")
	fmt.Fprintf(w, "# TYPE unifi_threat_sync_sync_total counter\n")
	fmt.Fprintf(w, "unifi_threat_sync_sync_total %d\n", hs.syncCount.Load())
	
	fmt.Fprintf(w, "# HELP unifi_threat_sync_errors_total Total number of errors\n")
	fmt.Fprintf(w, "# TYPE unifi_threat_sync_errors_total counter\n")
	fmt.Fprintf(w, "unifi_threat_sync_errors_total %d\n", hs.errorCount.Load())
	
	fmt.Fprintf(w, "# HELP unifi_threat_sync_uptime_seconds Uptime in seconds\n")
	fmt.Fprintf(w, "# TYPE unifi_threat_sync_uptime_seconds gauge\n")
	fmt.Fprintf(w, "unifi_threat_sync_uptime_seconds %.0f\n", time.Since(hs.startTime).Seconds())
}
