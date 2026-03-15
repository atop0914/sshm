package ssh

import (
	"testing"

	"github.com/sshm/sshm/internal/models"
)

func TestParseProxyHost(t *testing.T) {
	tests := []struct {
		input    string
		wantHost string
		wantUser string
		wantPort int
	}{
		{"example.com", "example.com", "", 22},
		{"example.com:2222", "example.com", "", 2222},
		{"user@example.com", "example.com", "user", 22},
		{"user@example.com:2222", "example.com", "user", 2222},
		{"192.168.1.1:22", "192.168.1.1", "", 22},
		{"jump.server.com", "jump.server.com", "", 22},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			host, user, port, err := parseProxyHost(tt.input)
			if err != nil {
				t.Fatalf("parseProxyHost() error = %v", err)
			}
			if host != tt.wantHost {
				t.Errorf("parseProxyHost() host = %v, want %v", host, tt.wantHost)
			}
			if user != tt.wantUser {
				t.Errorf("parseProxyHost() user = %v, want %v", user, tt.wantUser)
			}
			if port != tt.wantPort {
				t.Errorf("parseProxyHost() port = %v, want %v", port, tt.wantPort)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		input    string
		wantHome bool
	}{
		{"~/.ssh/id_rsa", true},
		{"/etc/passwd", false},
		{"relative/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := expandPath(tt.input)
			if err != nil {
				t.Fatalf("expandPath() error = %v", err)
			}
			if tt.wantHome && result == tt.input {
				t.Errorf("expandPath() did not expand path")
			}
			if !tt.wantHome && result != tt.input {
				t.Errorf("expandPath() should not have changed path")
			}
		})
	}
}

func TestHostGenerateSSHCommand(t *testing.T) {
	tests := []struct {
		host     models.Host
		expected string
	}{
		{
			host: models.Host{
				Host: "192.168.1.1",
				User: "admin",
				Port: 22,
			},
			expected: "ssh admin@192.168.1.1",
		},
		{
			host: models.Host{
				Host: "192.168.1.1",
				User: "admin",
				Port: 2222,
			},
			expected: "ssh -p 2222 admin@192.168.1.1",
		},
		{
			host: models.Host{
				Host:     "192.168.1.1",
				User:     "admin",
				Port:     22,
				Identity: "~/.ssh/id_rsa",
			},
			expected: "ssh -i ~/.ssh/id_rsa admin@192.168.1.1",
		},
		{
			host: models.Host{
				Host:   "192.168.1.1",
				User:   "admin",
				Port:   22,
				Proxy:  "jump.example.com",
			},
			expected: "ssh -J jump.example.com admin@192.168.1.1",
		},
		{
			host: models.Host{
				Host:     "192.168.1.1",
				User:     "admin",
				Port:     2222,
				Identity: "~/.ssh/id_rsa",
				Proxy:    "jump.example.com",
			},
			expected: "ssh -p 2222 -i ~/.ssh/id_rsa -J jump.example.com admin@192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.host.Host, func(t *testing.T) {
			result := tt.host.GenerateSSHCommand()
			if result != tt.expected {
				t.Errorf("GenerateSSHCommand() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPing(t *testing.T) {
	// Test connecting to a known unreachable host
	err := Ping("192.0.2.1", 22) // TEST-NET-1, should fail
	if err == nil {
		t.Error("Ping() should have failed for unreachable host")
	}

	// Test invalid port
	err = Ping("localhost", 1)
	if err == nil {
		t.Error("Ping() should have failed for invalid port")
	}
}
