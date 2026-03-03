package models

import "time"

// ConnectionHistory tracks connections to SSH hosts
type ConnectionHistory struct {
	HostID     string    `json:"host_id" yaml:"host_id"`
	Timestamp  time.Time `json:"timestamp" yaml:"timestamp"`
	Success    bool      `json:"success" yaml:"success"`
	Error      string    `json:"error,omitempty" yaml:"error,omitempty"`
	Duration   int64     `json:"duration_ms,omitempty" yaml:"duration_ms,omitempty"` // connection time in milliseconds
}

// HistoryStats contains aggregated connection statistics for a host
type HistoryStats struct {
	HostID           string    `json:"host_id"`
	TotalConnections int       `json:"total_connections"`
	SuccessfulConns  int       `json:"successful_connections"`
	FailedConns      int       `json:"failed_connections"`
	LastConnected    time.Time `json:"last_connected"`
}
