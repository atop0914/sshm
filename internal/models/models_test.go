package models

import (
	"testing"
)

func TestHostFields(t *testing.T) {
	h := Host{
		ID:       "test-id",
		Name:     "test-host",
		Host:     "192.168.1.100",
		Port:     22,
		User:     "admin",
		Password: "secret",
		Identity: "~/.ssh/id_rsa",
		AuthType: AuthTypeKey,
		Proxy:    "jump.example.com",
		Group:    "production",
		Tags:     []string{"web", "production"},
	}

	if h.ID != "test-id" {
		t.Errorf("Expected ID test-id, got %s", h.ID)
	}
	if h.Name != "test-host" {
		t.Errorf("Expected Name test-host, got %s", h.Name)
	}
	if h.Host != "192.168.1.100" {
		t.Errorf("Expected Host 192.168.1.100, got %s", h.Host)
	}
	if h.Port != 22 {
		t.Errorf("Expected Port 22, got %d", h.Port)
	}
	if h.User != "admin" {
		t.Errorf("Expected User admin, got %s", h.User)
	}
	if h.AuthType != AuthTypeKey {
		t.Errorf("Expected AuthType key, got %s", h.AuthType)
	}
	if h.Proxy != "jump.example.com" {
		t.Errorf("Expected Proxy jump.example.com, got %s", h.Proxy)
	}
}

func TestHostGenerateSSHCommand(t *testing.T) {
	tests := []struct {
		name     string
		host     Host
		expected string
	}{
		{
			name: "basic",
			host: Host{
				Host: "server.com",
				User: "user",
				Port: 22,
			},
			expected: "ssh user@server.com",
		},
		{
			name: "with port",
			host: Host{
				Host: "server.com",
				User: "user",
				Port: 2222,
			},
			expected: "ssh -p 2222 user@server.com",
		},
		{
			name: "with identity",
			host: Host{
				Host:     "server.com",
				User:     "user",
				Port:     22,
				Identity: "~/.ssh/id_rsa",
			},
			expected: "ssh -i ~/.ssh/id_rsa user@server.com",
		},
		{
			name: "with proxy",
			host: Host{
				Host:  "server.com",
				User:  "user",
				Port:  22,
				Proxy: "jump.com",
			},
			expected: "ssh -J jump.com user@server.com",
		},
		{
			name: "full",
			host: Host{
				Host:     "server.com",
				User:     "user",
				Port:     2222,
				Identity: "~/.ssh/id_rsa",
				Proxy:    "jump.com",
			},
			expected: "ssh -p 2222 -i ~/.ssh/id_rsa -J jump.com user@server.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.host.GenerateSSHCommand()
			if result != tt.expected {
				t.Errorf("GenerateSSHCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestProfile(t *testing.T) {
	p := Profile{
		Name:                "test-profile",
		Timeout:             60,
		KeepAliveInterval:   30,
		KeepAliveCountMax:  5,
		ServerAliveEnabled: true,
	}

	if p.Name != "test-profile" {
		t.Errorf("Expected Name test-profile, got %s", p.Name)
	}
	if p.Timeout != 60 {
		t.Errorf("Expected Timeout 60, got %d", p.Timeout)
	}
}

func TestDefaultProfile(t *testing.T) {
	p := DefaultProfile()
	if p.Name != "default" {
		t.Errorf("Expected default profile name, got %s", p.Name)
	}
	if p.Timeout != 30 {
		t.Errorf("Expected default timeout 30, got %d", p.Timeout)
	}
	if !p.ServerAliveEnabled {
		t.Errorf("Expected ServerAliveEnabled to be true by default")
	}
}
